package engine

import (
	"github.com/apmckinlay/gsuneido/compile/ast"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
	"github.com/apmckinlay/gsuneido/core"
)

// Types/InGuard and Members/MemberInGuard are parallel maps: update them in lockstep or guard provenance goes wrong.
type narrowScope struct {
	Types               map[string]DynType
	InGuard             map[string]bool
	Members             map[string]DynType
	MemberInGuard       map[string]bool
	memberAssignedFalse map[string]bool // computed once per pass, shared read-only across forks; never mutate
	writes              *classMemberWrites
	postconds           boolPostconds
	postHook            func(*ast.Return, narrowScope)
}

// set by tests to enforce the lockstep invariant at mutation points; nil in production
var lockstepCheck func(sc narrowScope, site string)

func assertLockstep(sc narrowScope, site string) {
	if lockstepCheck != nil {
		lockstepCheck(sc, site)
	}
}

func newNarrowScope(size int) narrowScope {
	return narrowScope{
		Types:         make(map[string]DynType, size),
		InGuard:       make(map[string]bool, size),
		Members:       make(map[string]DynType, size),
		MemberInGuard: make(map[string]bool, size),
	}
}

func (s narrowScope) clone() narrowScope {
	c := newNarrowScope(len(s.Types))
	for k, v := range s.Types {
		c.Types[k] = v
	}
	for k, v := range s.InGuard {
		c.InGuard[k] = v
	}
	for k, v := range s.Members {
		c.Members[k] = v
	}
	for k, v := range s.MemberInGuard {
		c.MemberInGuard[k] = v
	}
	c.memberAssignedFalse = s.memberAssignedFalse // shared, read-only
	c.writes = s.writes                           // shared, read-only
	c.postconds = s.postconds                     // shared, read-only
	c.postHook = s.postHook                       // shared, read-only
	return c
}

type narrowTarget struct {
	name     string // bare name, no leading dot for members
	isMember bool
	node     ast.Node // *ast.Ident or *ast.Mem
}

func unwrapTarget(e ast.Expr) (narrowTarget, bool) {
	if id, ok := unwrapIdent(e); ok && !isGlobalIdent(id.Name) {
		return narrowTarget{name: id.Name, node: id}, true
	}
	if name, mem, ok := unwrapThisMember(e); ok {
		return narrowTarget{name: name, isMember: true, node: mem}, true
	}
	return narrowTarget{}, false
}

// peelParens strips position and parenthesis wrappers to reach the expression
// that actually carries a type stamp.
func peelParens(e ast.Expr) ast.Expr {
	for {
		switch x := e.(type) {
		case *ast.ExprPos:
			if x.Expr == nil {
				return e
			}
			e = x.Expr
		case *ast.Unary:
			if x.Tok != tok.LParen || x.E == nil {
				return e
			}
			e = x.E
		default:
			return e
		}
	}
}

func unwrapThisMember(e ast.Expr) (string, *ast.Mem, bool) {
	for {
		switch x := e.(type) {
		case *ast.ExprPos:
			if x.Expr == nil {
				return "", nil, false
			}
			e = x.Expr
		case *ast.Unary:
			if x.Tok != tok.LParen {
				return "", nil, false
			}
			e = x.E
		default:
			mem, ok := e.(*ast.Mem)
			if !ok {
				return "", nil, false
			}
			id, ok := mem.E.(*ast.Ident)
			if !ok || id.Name != "this" {
				return "", nil, false
			}
			name, ok := memberName(mem.M)
			if !ok {
				return "", nil, false
			}
			return name, mem, true
		}
	}
}

var predicateTargets = map[string][]DynType{
	"Boolean?":  {TBoolean},
	"Number?":   {TNumber},
	"String?":   {TString},
	"Date?":     {TDate},
	"Object?":   {TObject},
	"Record?":   {TObject},
	"Class?":    {TClass},
	"Function?": {TFunction, TBlock},
}

var typeStringTargets = map[string][]DynType{
	"Boolean":  {TBoolean},
	"Number":   {TNumber},
	"String":   {TString},
	"Date":     {TDate},
	"Object":   {TObject},
	"Record":   {TObject},
	"Class":    {TClass},
	"Function": {TFunction},
	"Block":    {TBlock},
}

func FlowNarrowingPass(cls *ClassObject, env TypeEnv) {
	assignedFalse := membersAssignedFalse(cls, env)
	writes := ComputeMemberWrites(cls)
	postconds := ComputeBoolPostconditions(cls, env, assignedFalse, writes)
	for _, fn := range cls.SortedMethods {
		sc := initialNarrowScope(fn, env)
		sc.memberAssignedFalse = assignedFalse
		sc.writes = writes
		sc.postconds = postconds
		walkBlock(fn.Body, env, sc)
	}
}

func initialNarrowScope(fn *ast.Function, env TypeEnv) narrowScope {
	sc := newNarrowScope(len(fn.Params) + 8)
	for i := range fn.Params {
		p := &fn.Params[i]
		name := p.Name.ParamName()
		if t, ok := env.Params[p]; ok {
			sc.Types[name] = t
		} else if len(p.Name.Name) > 0 && p.Name.Name[0] == '.' {
			if t, ok := env.LookupMember(name); ok {
				sc.Types[name] = t
			}
		}
	}
	return sc
}

//nolint:gocognit,gocyclo,funlen // exhaustive AST dispatch
func narrowWalk(n ast.Node, env TypeEnv, sc narrowScope) {
	if n == nil {
		return
	}
	switch x := n.(type) {
	case *ast.ExprPos:
		if x.Expr != nil {
			narrowWalk(x.Expr, env, sc)
			if t, ok := env.Nodes[x.Expr]; ok {
				env.SetType(x, t)
			}
		}
		return
	case *ast.Compound:
		walkBlock(x.Body, env, sc)
		return
	case *ast.If:
		narrowIf(x, env, sc)
		return
	case *ast.Trinary:
		narrowWalk(x.Cond, env, sc)
		tScope := refineCond(x.Cond, sc, true, env, true)
		narrowWalk(x.T, env, tScope)
		fScope := refineCond(x.Cond, sc, false, env, true)
		narrowWalk(x.F, env, fScope)
		if t := trinaryType(x, env); t != TUnknown {
			env.SetType(x, t)
		}
		return
	case *ast.Switch:
		narrowSwitch(x, env, sc)
		return
	case *ast.While:
		loopSc := loopEntryScope(x.Body, nil, sc, env)
		narrowWalk(x.Cond, env, loopSc)
		body := refineCond(x.Cond, loopSc, true, env, true)
		narrowWalk(x.Body, env, body)
		return
	case *ast.For:
		loopSc := loopEntryScope(x.Body, x.Inc, sc, env)
		for _, e := range x.Init {
			narrowWalk(e, env, loopSc)
		}
		body := loopSc
		if x.Cond != nil {
			narrowWalk(x.Cond, env, loopSc)
			body = refineCond(x.Cond, loopSc, true, env, true)
		}
		narrowWalk(x.Body, env, body)
		for _, e := range x.Inc {
			narrowWalk(e, env, body)
		}
		return
	case *ast.DoWhile:
		loopSc := loopEntryScope(x.Body, nil, sc, env)
		narrowWalk(x.Body, env, loopSc)
		narrowWalk(x.Cond, env, loopSc)
		return
	case *ast.ForIn, *ast.Forever:
		// no boolean guard to exploit; clone + strip loop-carried.
		loopSc := loopEntryScope(loopBodyOf(n), nil, sc, env)
		n.Children(func(c ast.Node) ast.Node {
			narrowWalk(c, env, loopSc)
			return c
		})
		return
	case *ast.Nary:
		if x.Tok == tok.And || x.Tok == tok.Or {
			siblingPolarity := x.Tok == tok.And
			sub := sc.clone()
			for _, e := range x.Exprs {
				narrowWalk(e, env, sub)
				applyRefinement(e, sub, siblingPolarity, env, true)
			}
			return
		}
	case *ast.Binary:
		if x.Tok == tok.Eq {
			narrowEqAssign(x, env, sc)
			return
		}
		if x.Tok.IsAssign() {
			n.Children(func(c ast.Node) ast.Node {
				narrowWalk(c, env, sc)
				return c
			})
			if name, _, ok := unwrapThisMember(x.Lhs); ok {
				delete(sc.Members, name)
				delete(sc.MemberInGuard, name)
			} else if id, ok := x.Lhs.(*ast.Ident); ok && !isGlobalIdent(id.Name) {
				delete(sc.Types, id.Name)
				delete(sc.InGuard, id.Name)
			}
			return
		}
	case *ast.Unary:
		if x.Tok == tok.Inc || x.Tok == tok.Dec ||
			x.Tok == tok.PostInc || x.Tok == tok.PostDec {
			n.Children(func(c ast.Node) ast.Node {
				narrowWalk(c, env, sc)
				return c
			})
			if name, _, ok := unwrapThisMember(x.E); ok {
				delete(sc.Members, name)
				delete(sc.MemberInGuard, name)
			} else if id, ok := x.E.(*ast.Ident); ok && !isGlobalIdent(id.Name) {
				delete(sc.Types, id.Name)
				delete(sc.InGuard, id.Name)
			}
			return
		}
	case *ast.Ident:
		if !isGlobalIdent(x.Name) && sc.InGuard[x.Name] {
			if t, ok := sc.Types[x.Name]; ok && t != nil {
				env.SetType(x, t)
			}
		}
		return
	case *ast.Mem:
		if name, _, ok := unwrapThisMember(x); ok && sc.MemberInGuard[name] {
			if t, ok2 := sc.Members[name]; ok2 && t != nil {
				env.SetType(x, t)
			}
		}
	case *ast.Return:
		n.Children(func(c ast.Node) ast.Node {
			narrowWalk(c, env, sc)
			return c
		})
		if sc.postHook != nil {
			sc.postHook(x, sc)
		}
		return
	case *ast.Call:
		n.Children(func(c ast.Node) ast.Node {
			narrowWalk(c, env, sc)
			return c
		})
		eff, opaque := sc.writes.callEffect(x)
		for k := range sc.Members {
			if !opaque && !eff[k] {
				continue // refinement survives: call can't touch k
			}
			if t, keep := memberTypeAcrossCall(k, sc, env); keep {
				sc.Members[k] = t
				continue
			}
			delete(sc.Members, k)
			delete(sc.MemberInGuard, k)
		}
		assertLockstep(sc, "call")
		return
	}
	n.Children(func(c ast.Node) ast.Node {
		narrowWalk(c, env, sc)
		return c
	})
}

func narrowEqAssign(x *ast.Binary, env TypeEnv, sc narrowScope) {
	narrowWalk(x.Rhs, env, sc)
	if id, ok := x.Lhs.(*ast.Ident); ok && !isGlobalIdent(id.Name) {
		// new value invalidates any prior guard refinement
		delete(sc.Types, id.Name)
		delete(sc.InGuard, id.Name)
		rhsT := env.GetType(x.Rhs)
		lhsT := env.GetType(id)
		if isNarrower(rhsT, lhsT) {
			sc.Types[id.Name] = rhsT
			sc.InGuard[id.Name] = true
			env.SetType(id, rhsT)
		}
	} else if name, mem, ok := unwrapThisMember(x.Lhs); ok {
		delete(sc.Members, name)
		delete(sc.MemberInGuard, name)
		rhsT := env.GetType(x.Rhs)
		lhsT := env.GetType(mem)
		if isNarrower(rhsT, lhsT) {
			sc.Members[name] = rhsT
			sc.MemberInGuard[name] = true
			env.SetType(mem, rhsT)
		}
	} else {
		narrowWalk(x.Lhs, env, sc)
	}
	assertLockstep(sc, "assign")
}

func loopEntryScope(body ast.Statement, inc []ast.Expr, sc narrowScope, env TypeEnv) narrowScope {
	out := sc.clone()
	w := collectLoopWrites(body, sc.writes)
	for _, e := range inc {
		w.scan(e)
	}
	for name := range w.locals {
		delete(out.Types, name)
		delete(out.InGuard, name)
	}
	for name := range out.Members {
		if !w.opaque && !w.callMembers[name] && !w.members[name] {
			continue
		}
		if !w.members[name] {
			if t, keep := memberTypeAcrossCall(name, sc, env); keep {
				out.Members[name] = t
				continue
			}
		}
		delete(out.Members, name)
		delete(out.MemberInGuard, name)
	}
	assertLockstep(out, "loopEntryScope")
	return out
}

type loopWrites struct {
	locals      map[string]bool
	members     map[string]bool
	callMembers map[string]bool
	opaque      bool
	writes      *classMemberWrites
}

// `v = v` cannot change v's value on any iteration, so it does not
// invalidate a loop-entry fact. only plain `=` qualifies: `v += v` widens.
func selfAssign(b *ast.Binary) bool {
	if b.Tok != tok.Eq {
		return false
	}
	lhs, ok := b.Lhs.(*ast.Ident)
	if !ok {
		return false
	}
	rhs, ok := unwrapIdent(b.Rhs)
	return ok && rhs.Name == lhs.Name
}

func collectLoopWrites(body ast.Statement, writes *classMemberWrites) *loopWrites {
	w := &loopWrites{
		locals:      map[string]bool{},
		members:     map[string]bool{},
		callMembers: map[string]bool{},
		writes:      writes,
	}
	if body != nil {
		w.scan(body)
	}
	return w
}

func (w *loopWrites) scan(n ast.Node) {
	if n == nil {
		return
	}
	switch x := n.(type) {
	case *ast.Call:
		eff, opaque := w.writes.callEffect(x)
		if opaque {
			w.opaque = true
		}
		for m := range eff {
			w.callMembers[m] = true
		}
	case *ast.Binary:
		if x.Tok.IsAssign() && !selfAssign(x) {
			w.mark(x.Lhs)
		}
	case *ast.Unary:
		if x.Tok == tok.Inc || x.Tok == tok.Dec ||
			x.Tok == tok.PostInc || x.Tok == tok.PostDec {
			w.mark(x.E)
		}
	}
	n.Children(func(c ast.Node) ast.Node {
		w.scan(c)
		return c
	})
}

func (w *loopWrites) mark(e ast.Expr) {
	if id, ok := e.(*ast.Ident); ok && !isGlobalIdent(id.Name) {
		w.locals[id.Name] = true
	} else if name, _, ok := unwrapThisMember(e); ok {
		w.members[name] = true
	}
}

func loopBodyOf(n ast.Node) ast.Statement {
	switch x := n.(type) {
	case *ast.ForIn:
		return x.Body
	case *ast.Forever:
		return x.Body
	}
	return nil
}

func memberTypeAcrossCall(name string, sc narrowScope, env TypeEnv) (DynType, bool) {
	if sc.memberAssignedFalse == nil || sc.memberAssignedFalse[name] {
		return nil, false // can't prove it stayed non-false
	}
	cw, ok := env.LookupMember(name)
	if !ok {
		return nil, false
	}
	rf := removeFalse(cw)
	if rf == TUnknown || typeHasBoolish(rf) {
		return nil, false
	}
	return rf, true
}

// typeHasBoolish reports whether t has a True or Boolean arm.
func typeHasBoolish(t DynType) bool {
	switch x := t.(type) {
	case Primitive:
		return x == TTrue || x == TBoolean
	case Union:
		for _, m := range x.Types {
			if m == TTrue || m == TBoolean {
				return true
			}
		}
	}
	return false
}

func walkBlock(stmts []ast.Statement, env TypeEnv, sc narrowScope) {
	for _, stmt := range stmts {
		if ifStmt, ok := stmt.(*ast.If); ok {
			narrowIf(ifStmt, env, sc)
			continue
		}
		narrowWalk(stmt, env, sc)
		if call, ok := assertStmtCall(stmt); ok {
			applyAssertStmt(call, sc, env)
		}
	}
}

// the Assert(...) call of an expression-statement, if stmt is one.
func assertStmtCall(stmt ast.Statement) (*ast.Call, bool) {
	es, ok := stmt.(*ast.ExprStmt)
	if !ok || es.E == nil {
		return nil, false
	}
	e := es.E
	if ep, ok := e.(*ast.ExprPos); ok && ep.Expr != nil {
		e = ep.Expr
	}
	call, ok := e.(*ast.Call)
	if !ok {
		return nil, false
	}
	if id, ok := call.Fn.(*ast.Ident); !ok || id.Name != "Assert" {
		return nil, false
	}
	if len(call.Args) == 0 || call.Args[0].Name != nil {
		return nil, false
	}
	return call, true
}

// executing past an Assert proves its claim (a failure throws), so the fact
// holds for the rest of the block. two shapes (see stdlib Assert):
//
// ```suneido
// Assert(String?(x))          // old style: a condition, msg optional
// Assert(x isString:)         // matcher style: Matcher_<name>.Match(x, args)
//
//	^ ^^^^^^^^^ the sole non-msg named arg names the matcher
//
// ```
func applyAssertStmt(call *ast.Call, sc narrowScope, env TypeEnv) {
	name, matcherArg, ok := assertMatcher(call)
	if !ok {
		applyRefinement(call.Args[0].E, sc, true, env, true)
		return
	}
	tgt, tok := unwrapTarget(call.Args[0].E)
	if !tok {
		return
	}
	existing := existingType(sc, tgt, env)
	if targets, ok := assertMatcherTargets[name]; ok {
		storeRefinement(sc, tgt, narrowTowardSet(existing, targets), true)
		return
	}
	switch name {
	case "is": // Matcher_is: equality with the expected literal
		if c, ok := unwrapConstant(matcherArg); ok {
			storeRefinement(sc, tgt,
				narrowTowardSet(existing, []DynType{DynTypeOfSuValue(c.Val)}), true)
		}
	case "isnt":
		if c, ok := unwrapConstant(matcherArg); ok {
			if t := DynTypeOfSuValue(c.Val); t == TFalse || t == TTrue {
				storeRefinement(sc, tgt, narrowAwaySet(existing, []DynType{t}), true)
			}
		}
	case "isType": // Matcher_isType: Type(value) is args
		if c, ok := unwrapConstant(matcherArg); ok {
			if targets, ok := typeStringTargets[core.ToStr(c.Val)]; ok {
				storeRefinement(sc, tgt, narrowTowardSet(existing, targets), true)
			}
		}
	}
}

// the sole non-msg named arg of a matcher-style Assert; ok=false means old
// style (condition form). unrecognized matchers still return ok=true so the
// condition path is not misapplied to a matcher subject.
func assertMatcher(call *ast.Call) (name string, arg ast.Expr, ok bool) {
	for i := range call.Args {
		a := &call.Args[i]
		if a.Name == nil || isAtArg(a) {
			continue
		}
		n := core.ToStr(a.Name)
		if n == "msg" {
			continue
		}
		return n, a.E, true
	}
	return "", nil, false
}

// matchers whose Match is exactly a type predicate (see stdlib Matcher_is*);
// the any* forms subclass the is* forms unchanged. mirror predicateTargets.
var assertMatcherTargets = map[string][]DynType{
	"isString":         {TString},
	"anyString":        {TString},
	"isNumber":         {TNumber},
	"anyNumber":        {TNumber},
	"isInt":            {TNumber},
	"isIntNonNegative": {TNumber},
	"isIntInRange":     {TNumber},
	"isObject":         {TObject},
	"anyObject":        {TObject},
	"isBoolean":        {TBoolean},
	"isDate":           {TDate},
	"isCallable":       {TFunction, TBlock}, // Matcher_isCallable is Function?
}

func narrowIf(x *ast.If, env TypeEnv, sc narrowScope) {
	narrowWalk(x.Cond, env, sc)
	// snapshot before the branches fork - joinBranchKills needs the pre-if
	// facts both as the key set to re-check and as the no-else fall-through
	preTypes := factsInGuard(sc.Types, sc.InGuard)
	preMembers := factsInGuard(sc.Members, sc.MemberInGuard)
	thenScope := refineCond(x.Cond, sc, true, env, true)
	narrowWalk(x.Then, env, thenScope)
	var elseScope narrowScope
	if x.Else != nil {
		elseScope = refineCond(x.Cond, sc, false, env, true)
		narrowWalk(x.Else, env, elseScope)
	}
	// kills first: the merge blocks below rebuild from sc, so dropping a dead
	// refinement here keeps them from resurrecting it
	joinBranchKills(x, sc, preTypes, preMembers, thenScope, elseScope)
	if x.Else == nil && !branchAlwaysExits(x.Then) {
		joinNoElseLocals(x, env, sc)
	}
	if x.Else == nil && !branchAlwaysExits(x.Then) {
		joinNoElseMembers(x, env, sc)
	}
	// early-return: if exactly one branch always exits, siblings after the
	// if inherit the other branch's refinement.
	//
	// ```suneido
	// Foo(x) {
	//     if not Number?(x)
	//         return false       // then-branch exits
	//     ^^^^^^^^^^^^^^^^^
	//     return x + 1           // <-- siblings see x narrowed to TNumber
	//            ^                   (Number? polarity flipped via the `not`)
	// }
	// ```
	// member refinements always installed - narrowWalk's Call/assign/Inc-Dec
	// handlers clear them position-aware as sibling statements walk.
	thenExits := branchAlwaysExits(x.Then)
	elseExits := x.Else != nil && branchAlwaysExits(x.Else)
	if thenExits && !elseExits {
		applyRefinement(x.Cond, sc, false, env, true)
	} else if elseExits && !thenExits {
		applyRefinement(x.Cond, sc, true, env, true)
	}
	assertLockstep(sc, "narrowIf")
}

func joinNoElseLocals(x *ast.If, env TypeEnv, sc narrowScope) {
	assigned := topLevelAssignedLocals(x.Then)
	if len(assigned) == 0 {
		return
	}
	negScope := refineCond(x.Cond, sc, false, env, false)
	for name := range assigned {
		if !negScope.InGuard[name] {
			continue
		}
		thenT := lastTopLevelAssignType(x.Then, name, env)
		if thenT == nil {
			continue
		}
		elseT := negScope.Types[name]
		var merged DynType
		switch {
		case elseT == TUnknown:
			entryT := condEntryType(x.Cond, name, env)
			if entryT == nil || entryT == TUnknown {
				continue
			}
			merged = thenT
		case elseT != nil:
			merged = U(thenT, elseT)
		default:
			continue
		}
		if merged == nil || merged == TUnknown {
			continue
		}
		sc.Types[name] = merged
		sc.InGuard[name] = true
	}
}

func joinNoElseMembers(x *ast.If, env TypeEnv, sc narrowScope) {
	assignedM := topLevelAssignedMembers(x.Then)
	if len(assignedM) == 0 {
		return
	}
	negScope := refineCond(x.Cond, sc, false, env, true)
	for name := range assignedM {
		if !negScope.MemberInGuard[name] {
			continue
		}
		thenT := lastTopLevelMemberAssignType(x.Then, name, env)
		elseT := negScope.Members[name]
		if thenT == nil || elseT == nil || elseT == TUnknown {
			continue
		}
		merged := U(thenT, elseT)
		if merged == TUnknown || typeHasBoolish(merged) {
			continue
		}
		sc.Members[name] = merged
		sc.MemberInGuard[name] = true
	}
}

// factsInGuard snapshots a scope's live refinements as a plain name->type map.
// Only in-guard entries count: the parallel guard map is what makes a
// refinement visible to reads (see narrowWalk's *ast.Ident and *ast.Mem cases).
func factsInGuard(vals map[string]DynType, guard map[string]bool) map[string]DynType {
	out := make(map[string]DynType, len(vals))
	for name, t := range vals {
		if t != nil && guard[name] {
			out[name] = t
		}
	}
	return out
}

// joinBranchKills propagates branch *removals* back to the enclosing scope.
//
// refineCond hands each branch its own clone, so a refinement a branch
// destroys - a member-writing call, a reassignment nested below the top level
// - dies with that clone and sc walks on believing the stale fact. narrowIf's
// blocks merge back what a branch *adds*; this is the missing mirror for what
// it takes away.
//
// ```suneido
//
//	M1() { .f1 = "x" }
//	M0() {
//	    .f1 = #20200101      // refines .f1 -> Date
//	    if (true) { .M1() }  // the kill lands on the clone, not on sc
//	    return .f1           // ^^^ so this read answered Date, not String
//	}
//
// ```
func joinBranchKills(x *ast.If, sc narrowScope, preTypes, preMembers map[string]DynType,
	thenScope, elseScope narrowScope) {
	var types, members []map[string]DynType
	if !branchAlwaysExits(x.Then) {
		types = append(types, factsInGuard(thenScope.Types, thenScope.InGuard))
		members = append(members, factsInGuard(thenScope.Members, thenScope.MemberInGuard))
	}
	switch {
	case x.Else == nil:
		// skipping the if is itself a route out, and it carries the pre-if facts
		types = append(types, preTypes)
		members = append(members, preMembers)
	case !branchAlwaysExits(x.Else):
		types = append(types, factsInGuard(elseScope.Types, elseScope.InGuard))
		members = append(members, factsInGuard(elseScope.Members, elseScope.MemberInGuard))
	}
	if len(types) == 0 {
		return // every route exits - no siblings below the if to protect
	}
	joinReachingFacts(preTypes, types, sc.Types, sc.InGuard)
	joinReachingFacts(preMembers, members, sc.Members, sc.MemberInGuard)
	assertLockstep(sc, "joinBranchKills")
}

// joinReachingFacts intersects the pre-if refinements against every route that
// reaches past the if, writing the survivors back into vals/guard. A name keeps
// its refinement only if all routes still refine it, at the union of the types
// they give it; whatever any route dropped is dropped here too.
func joinReachingFacts(pre map[string]DynType, routes []map[string]DynType,
	vals map[string]DynType, guard map[string]bool) {
	for name := range pre {
		var merged DynType
		for _, r := range routes {
			t, ok := r[name]
			if !ok || t == nil || t == TUnknown {
				merged = nil
				break
			}
			if merged == nil {
				merged = t
			} else {
				merged = U(merged, t)
			}
		}
		if merged == nil || merged == TUnknown {
			delete(vals, name)
			delete(guard, name)
			continue
		}
		vals[name] = merged
		guard[name] = true
	}
}

func branchAlwaysExits(s ast.Statement) bool {
	if s == nil {
		return false
	}
	switch x := s.(type) {
	case *ast.Return, *ast.Throw, *ast.Break, *ast.Continue:
		return true
	case *ast.Compound:
		for _, st := range x.Body {
			if branchAlwaysExits(st) {
				return true
			}
		}
		return false
	case *ast.If:
		if x.Else == nil {
			return false
		}
		return branchAlwaysExits(x.Then) && branchAlwaysExits(x.Else)
	}
	return false
}

//nolint:gocognit,gocyclo // exhaustive union/primitive case analysis
func isNarrower(a, b DynType) bool {
	if a == nil || b == nil {
		return false
	}
	aP, aIsP := a.(Primitive)
	bP, bIsP := b.(Primitive)
	if aIsP && bIsP && aP == bP {
		return false
	}
	if bIsP && bP == TUnknown {
		return !aIsP || aP != TUnknown
	}
	if aIsP && aP == TUnknown {
		return false
	}
	// any clean type is strictly narrower than any dirty one: dirty admits
	// unknown values, clean is a proof. without this an assignment whose RHS
	// narrowing cleaned (e.g. a ternary arm reading a guarded param) cannot
	// rescue the local from its stale pre-narrowing dirty flow stamp.
	if _, bDirty := decomposeForCheck(b); bDirty {
		if _, aDirty := decomposeForCheck(a); !aDirty {
			return true
		}
	}
	aU, aIsU := a.(Union)
	bU, bIsU := b.(Union)
	switch {
	case !aIsU && !bIsU:
		return subtypeOf(a, b) && a != b
	case !aIsU && bIsU:
		for _, t := range bU.Types {
			if a == t || subtypeOf(a, t) {
				return true
			}
		}
		return false
	case aIsU && !bIsU:
		return false
	default: // both Unions
		for _, ta := range aU.Types {
			found := false
			for _, tb := range bU.Types {
				if ta == tb || subtypeOf(ta, tb) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		switch {
		case len(aU.Types) < len(bU.Types):
			return true
		case len(aU.Types) == len(bU.Types) && !aU.IsDirty && bU.IsDirty:
			return true
		}
		return false
	}
}

// TFalse/TTrue <: TBoolean, TSequence <: TObject.
func subtypeOf(sub, sup DynType) bool {
	subP, subOk := sub.(Primitive)
	supP, supOk := sup.(Primitive)
	if !subOk || !supOk {
		return false
	}
	if subP == supP {
		return true
	}
	if supP == TBoolean {
		return subP == TTrue || subP == TFalse
	}
	if supP == TObject {
		return subP == TSequence || subP == TClass
	}
	return false
}

func narrowSwitch(sw *ast.Switch, env TypeEnv, sc narrowScope) {
	narrowWalk(sw.E, env, sc)
	if isImplicitTrue(sw.E) {
		narrowSwitchCondMode(sw, env, sc)
	} else {
		narrowSwitchScrutineeMode(sw, env, sc)
	}
}

func isImplicitTrue(e ast.Expr) bool {
	c, ok := unwrapConstant(e)
	return ok && DynTypeOfSuValue(c.Val) == TTrue
}

func narrowSwitchCondMode(sw *ast.Switch, env TypeEnv, sc narrowScope) {
	defScope := sc.clone()
	for i := range sw.Cases {
		c := &sw.Cases[i]
		for _, e := range c.Exprs {
			narrowWalk(e, env, sc)
		}
		caseScope := sc.clone()
		if len(c.Exprs) == 1 {
			caseScope = refineCond(c.Exprs[0], sc, true, env, false)
		}
		for _, e := range c.Exprs {
			applyRefinement(e, defScope, false, env, false)
		}
		for _, stmt := range c.Body {
			narrowWalk(stmt, env, caseScope)
		}
	}
	for _, stmt := range sw.Default {
		narrowWalk(stmt, env, defScope)
	}
}

func narrowSwitchScrutineeMode(sw *ast.Switch, env TypeEnv, sc narrowScope) {
	id, isIdent := unwrapIdent(sw.E)
	canNarrow := isIdent && !isGlobalIdent(id.Name)

	defScope := sc.clone()
	var allTargets []DynType

	for i := range sw.Cases {
		c := &sw.Cases[i]
		for _, e := range c.Exprs {
			narrowWalk(e, env, sc)
		}
		caseScope := sc.clone()

		if canNarrow {
			targets := caseLiteralTargets(c.Exprs)
			if len(targets) > 0 {
				tgt := narrowTarget{name: id.Name, node: id}
				existing := existingType(sc, tgt, env)
				// allowMembers moot - scrutinee mode is always Ident
				storeRefinement(caseScope, tgt, narrowTowardSet(existing, targets), false)
				allTargets = append(allTargets, targets...)
			}
		}

		for _, stmt := range c.Body {
			narrowWalk(stmt, env, caseScope)
		}
	}

	if canNarrow && len(allTargets) > 0 {
		tgt := narrowTarget{name: id.Name, node: id}
		existing := existingType(defScope, tgt, env)
		storeRefinement(defScope, tgt, narrowAwaySet(existing, allTargets), false)
	}

	for _, stmt := range sw.Default {
		narrowWalk(stmt, env, defScope)
	}
}

func caseLiteralTargets(exprs []ast.Expr) []DynType {
	var targets []DynType
	for _, e := range exprs {
		if lit, ok := unwrapConstant(e); ok {
			targets = append(targets, DynTypeOfSuValue(lit.Val))
		}
	}
	return targets
}

func refineCond(cond ast.Expr, sc narrowScope, polarity bool, env TypeEnv, allowMembers bool) narrowScope {
	out := sc.clone()
	applyRefinement(cond, out, polarity, env, allowMembers)
	return out
}

func applyRefinement(cond ast.Expr, sc narrowScope, polarity bool, env TypeEnv, allowMembers bool) {
	if ep, ok := cond.(*ast.ExprPos); ok && ep.Expr != nil {
		applyRefinement(ep.Expr, sc, polarity, env, allowMembers)
		return
	}
	switch n := cond.(type) {
	case *ast.Unary:
		switch n.Tok {
		case tok.Not:
			applyRefinement(n.E, sc, !polarity, env, allowMembers)
		case tok.LParen:
			applyRefinement(n.E, sc, polarity, env, allowMembers)
		}
	case *ast.Nary:
		switch {
		case (n.Tok == tok.And && polarity) || (n.Tok == tok.Or && !polarity):
			for _, e := range n.Exprs {
				applyRefinement(e, sc, polarity, env, allowMembers)
			}
		case (n.Tok == tok.Or && polarity) || (n.Tok == tok.And && !polarity):
			mergeForkedRefinements(n.Exprs, sc, polarity, env, allowMembers)
		}
	case *ast.Binary:
		refineBinary(n, sc, polarity, env, allowMembers)
	case *ast.Call:
		refinePredicateCall(n, sc, polarity, env, allowMembers)
		refineHelperPostcondition(n, sc, polarity, allowMembers)
	}
}

func mergeForkedRefinements(exprs []ast.Expr, sc narrowScope, polarity bool, env TypeEnv, allowMembers bool) {
	if len(exprs) == 0 {
		return
	}
	perOp := make([]narrowScope, len(exprs))
	for i, e := range exprs {
		saved := overrideEnvWithScopeBaseline(e, sc, env)
		perOp[i] = sc.clone()
		applyRefinement(e, perOp[i], polarity, env, allowMembers)
		restoreEnvStamps(saved, env)
	}
	mergeForkedKind(sc.Types, sc.InGuard, perOp, false)
	if allowMembers {
		mergeForkedKind(sc.Members, sc.MemberInGuard, perOp, true)
	}
	assertLockstep(sc, "mergeForkedRefinements")
}

type savedNode struct {
	node    ast.Node
	present bool
	ty      DynType
}

// the chain walk stamps speculative types into env.Nodes; restoreEnvStamps must undo them or narrowed types leak outside the guard.
func overrideEnvWithScopeBaseline(e ast.Expr, sc narrowScope, env TypeEnv) []savedNode {
	var saved []savedNode
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *ast.Ident:
			if !isGlobalIdent(x.Name) {
				if t, ok := sc.Types[x.Name]; ok && t != nil {
					prev, present := env.Nodes[n]
					saved = append(saved, savedNode{node: n, present: present, ty: prev})
					env.Nodes[n] = t
				}
			}
		case *ast.Mem:
			if name, mem, ok := unwrapThisMember(x); ok {
				if t, ok2 := sc.Members[name]; ok2 && t != nil {
					prev, present := env.Nodes[mem]
					saved = append(saved, savedNode{node: mem, present: present, ty: prev})
					env.Nodes[mem] = t
				}
			}
		}
		n.Children(func(c ast.Node) ast.Node {
			walk(c)
			return c
		})
	}
	walk(e)
	return saved
}

// undoes overrideEnvWithScopeBaseline. skipping this looks safe and is not.
func restoreEnvStamps(saved []savedNode, env TypeEnv) {
	for _, s := range saved {
		if s.present {
			env.Nodes[s.node] = s.ty
		} else {
			delete(env.Nodes, s.node)
		}
	}
}

func mergeForkedKind(scTypes map[string]DynType, scInGuard map[string]bool, perOp []narrowScope, members bool) {
	cand := map[string]bool{}
	for _, op := range perOp {
		opIn := op.InGuard
		if members {
			opIn = op.MemberInGuard
		}
		for name := range opIn {
			cand[name] = true
		}
	}
	for name := range cand {
		merged, every := mergeForkedName(name, scTypes, scInGuard, perOp, members)
		if !every || merged == nil {
			continue
		}
		scTypes[name] = merged
		scInGuard[name] = true
	}
}

func mergeForkedName(name string, scTypes map[string]DynType, scInGuard map[string]bool,
	perOp []narrowScope, members bool) (DynType, bool) {
	var merged DynType
	for _, op := range perOp {
		opTypes, opIn := op.Types, op.InGuard
		if members {
			opTypes, opIn = op.Members, op.MemberInGuard
		}
		if !opIn[name] {
			return nil, false
		}
		opT := opTypes[name]
		if scInGuard[name] && dynEqual(opT, scTypes[name]) {
			return nil, false
		}
		if merged == nil {
			merged = opT
		} else {
			merged = U(merged, opT)
		}
	}
	return merged, true
}

func dynEqual(a, b DynType) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ua, aIsU := a.(Union)
	ub, bIsU := b.(Union)
	if aIsU != bIsU {
		return false
	}
	if !aIsU {
		return a == b
	}
	if len(ua.Types) != len(ub.Types) || ua.IsDirty != ub.IsDirty {
		return false
	}
	for _, ta := range ua.Types {
		found := false
		for _, tb := range ub.Types {
			if dynEqual(ta, tb) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func existingType(sc narrowScope, tgt narrowTarget, env TypeEnv) DynType {
	if tgt.isMember {
		if sc.MemberInGuard[tgt.name] {
			if t, ok := sc.Members[tgt.name]; ok && t != nil {
				return t
			}
		}
		if t := env.GetType(tgt.node); t != TUnknown {
			return t
		}
		if t, ok := env.LookupMember(tgt.name); ok {
			return t
		}
		return TUnknown
	}
	if sc.InGuard[tgt.name] {
		if t, ok := sc.Types[tgt.name]; ok && t != nil {
			return t
		}
	}
	return env.GetType(tgt.node)
}

// ```suneido
// if x is 5    { return x }    // x narrows toward TNumber inside { ... }
//
//	^^^^
//
// if x isnt false { return x } // TFalse dropped from x's type
//
//	^^^^^^^^^
//
// ```
func refineBinary(b *ast.Binary, sc narrowScope, polarity bool, env TypeEnv, allowMembers bool) {
	switch b.Tok {
	case tok.Is:
		applyEqRefinement(b, sc, polarity, env, allowMembers)
	case tok.Isnt:
		applyEqRefinement(b, sc, !polarity, env, allowMembers)
	}
}

func applyEqRefinement(b *ast.Binary, sc narrowScope, equal bool, env TypeEnv, allowMembers bool) {
	if tgt, lit, ok := targetAndLiteral(b); ok {
		t := DynTypeOfSuValue(lit.Val)
		existing := existingType(sc, tgt, env)
		if equal {
			storeRefinement(sc, tgt, narrowTowardSet(existing, []DynType{t}), allowMembers)
		} else if t == TFalse || t == TTrue {
			storeRefinement(sc, tgt, narrowAwaySet(existing, []DynType{t}), allowMembers)
		}
		return
	}
	refineTypeStringEq(b, sc, equal, env, allowMembers)
}

func storeRefinement(sc narrowScope, tgt narrowTarget, t DynType, allowMembers bool) {
	if t == nil {
		return
	}
	if tgt.isMember {
		if !allowMembers {
			return
		}
		sc.Members[tgt.name] = t
		sc.MemberInGuard[tgt.name] = true
		return
	}
	sc.Types[tgt.name] = t
	sc.InGuard[tgt.name] = true
	assertLockstep(sc, "storeRefinement")
}

func targetAndLiteral(b *ast.Binary) (narrowTarget, *ast.Constant, bool) {
	if tgt, ok := unwrapTarget(b.Lhs); ok {
		if c, ok := unwrapConstant(b.Rhs); ok {
			return tgt, c, true
		}
	}
	if tgt, ok := unwrapTarget(b.Rhs); ok {
		if c, ok := unwrapConstant(b.Lhs); ok {
			return tgt, c, true
		}
	}
	return narrowTarget{}, nil, false
}

// peels ExprPos, parens, and inline assignments to the underlying name.
//
// ```suneido
// if false is x = .Maybe()    return  // RHS is Binary(Eq), not Ident
//
//	^^^^^^^^^^^^^^^         // we still need to refine on x
//
// if ((x = .Maybe()) is false) ...    // LHS wrapped in Unary(LParen)
//
//	^^^^^^^^^^^^^^^                    // peel paren AND assign to find x
//
// ```
func unwrapIdent(e ast.Expr) (*ast.Ident, bool) {
	for {
		switch x := e.(type) {
		case *ast.ExprPos:
			if x.Expr == nil {
				return nil, false
			}
			e = x.Expr
		case *ast.Unary:
			if x.Tok != tok.LParen {
				return nil, false
			}
			e = x.E
		case *ast.Binary:
			if x.Tok != tok.Eq {
				return nil, false
			}
			e = x.Lhs
		case *ast.Ident:
			return x, true
		default:
			return nil, false
		}
	}
}

func unwrapConstant(e ast.Expr) (*ast.Constant, bool) {
	if ep, ok := e.(*ast.ExprPos); ok && ep.Expr != nil {
		e = ep.Expr
	}
	c, ok := e.(*ast.Constant)
	return c, ok
}

func unwrapCall(e ast.Expr) (*ast.Call, bool) {
	if ep, ok := e.(*ast.ExprPos); ok && ep.Expr != nil {
		e = ep.Expr
	}
	c, ok := e.(*ast.Call)
	return c, ok
}

// ```suneido
// if Type(x) is "Number"   { ... } // x narrows toward TNumber
//
//	^      ^^^^^^^^
//
// if Type(.foo) isnt "String" { ... } // TString removed from .foo
//
//	^^^^       ^^^^^^^^
//
// ```
// x may be a local Ident or a `this`-member Mem.
func refineTypeStringEq(b *ast.Binary, sc narrowScope, equal bool, env TypeEnv, allowMembers bool) {
	tryPair := func(callExpr, strExpr ast.Expr) bool {
		call, ok := unwrapCall(callExpr)
		if !ok {
			return false
		}
		fnId, ok := unwrapIdent(call.Fn)
		if !ok || fnId.Name != "Type" || len(call.Args) != 1 {
			return false
		}
		tgt, ok := unwrapTarget(call.Args[0].E)
		if !ok {
			return false
		}
		c, ok := unwrapConstant(strExpr)
		if !ok {
			return false
		}
		s := core.ToStr(c.Val)
		targets, ok := typeStringTargets[s]
		if !ok {
			return false
		}
		existing := existingType(sc, tgt, env)
		if equal {
			storeRefinement(sc, tgt, narrowTowardSet(existing, targets), allowMembers)
		} else {
			storeRefinement(sc, tgt, narrowAwaySet(existing, targets), allowMembers)
		}
		return true
	}
	if tryPair(b.Lhs, b.Rhs) {
		return
	}
	tryPair(b.Rhs, b.Lhs)
}

// ```suneido
// if Number?(x)   { return x } // x narrows to TNumber inside the body
//
//	^^^^^^^^^^                  // polarity=true  -> narrowTowardSet
//
// if not String?(x) return     // post-this-stmt x narrows to TString
//
//	^^^^^^^^^^^^^^              // polarity=false (from outer `not`)
//
// ```
func refinePredicateCall(call *ast.Call, sc narrowScope, polarity bool, env TypeEnv, allowMembers bool) {
	fn, ok := unwrapIdent(call.Fn)
	if !ok {
		return
	}
	targets, ok := predicateTargets[fn.Name]
	if !ok || len(call.Args) != 1 {
		return
	}
	tgt, ok := unwrapTarget(call.Args[0].E)
	if !ok {
		return
	}
	existing := existingType(sc, tgt, env)
	if polarity {
		storeRefinement(sc, tgt, narrowTowardSet(existing, targets), allowMembers)
	} else {
		storeRefinement(sc, tgt, narrowAwaySet(existing, targets), allowMembers)
	}
}

func boolIntersect(targets []DynType) DynType {
	keepF, keepT := false, false
	for _, tgt := range targets {
		switch tgt {
		case TFalse:
			keepF = true
		case TTrue:
			keepT = true
		case TBoolean:
			keepF, keepT = true, true
		}
	}
	switch {
	case keepF && keepT:
		return TBoolean
	case keepF:
		return TFalse
	case keepT:
		return TTrue
	}
	return TVoid
}

// dual of boolIntersect - removing [TFalse] leaves TTrue, etc.
func boolComplement(targets []DynType) DynType {
	removeF, removeT := false, false
	for _, tgt := range targets {
		switch tgt {
		case TFalse:
			removeF = true
		case TTrue:
			removeT = true
		case TBoolean:
			removeF, removeT = true, true
		}
	}
	switch {
	case removeF && removeT:
		return TVoid
	case removeF:
		return TTrue
	case removeT:
		return TFalse
	}
	return TBoolean
}

func narrowTowardSet(existing DynType, targets []DynType) DynType {
	if existing == nil || existing == TUnknown {
		return foldTargetUnion(targets)
	}
	if u, ok := existing.(Union); ok {
		return narrowUnionToward(u, targets)
	}
	if existing == TBoolean {
		if bi := boolIntersect(targets); bi != TVoid {
			return bi
		}
		return foldTargetUnion(targets)
	}
	for _, target := range targets {
		if subtypeOf(existing, target) {
			return existing
		}
	}
	return foldTargetUnion(targets)
}

func narrowUnionToward(u Union, targets []DynType) DynType {
	var kept []DynType
	for _, t := range u.Types {
		if t == TBoolean {
			if bi := boolIntersect(targets); bi != TVoid {
				kept = append(kept, bi)
			}
			continue
		}
		for _, target := range targets {
			if subtypeOf(t, target) {
				kept = append(kept, t)
				break
			}
		}
	}
	if u.IsDirty {
		kept = append(kept, targets...)
	}
	if len(kept) == 0 {
		return foldTargetUnion(targets)
	}
	return Union{Types: kept}.Fold()
}

func isAtomicType(t DynType) bool {
	p, ok := t.(Primitive)
	if !ok {
		return false
	}
	return p == TFalse || p == TTrue || p == TBoolean
}

func narrowAwaySet(existing DynType, targets []DynType) DynType {
	if existing == nil || existing == TUnknown {
		return existing
	}
	matches := func(t DynType) bool {
		for _, target := range targets {
			if subtypeOf(t, target) {
				return true
			}
		}
		return false
	}
	if u, ok := existing.(Union); ok {
		return narrowUnionAway(u, targets, matches)
	}
	if existing == TBoolean {
		if c := boolComplement(targets); c != TVoid {
			return c
		}
		return TUnknown
	}
	if matches(existing) {
		if isAtomicType(existing) {
			return TUnknown
		}
		return existing
	}
	return existing
}

func narrowUnionAway(u Union, targets []DynType, matches func(DynType) bool) DynType {
	var kept []DynType
	allAtomicRemoved := true
	for _, t := range u.Types {
		if t == TBoolean {
			if c := boolComplement(targets); c != TVoid {
				kept = append(kept, c)
			}
			continue
		}
		if !matches(t) {
			kept = append(kept, t)
			continue
		}
		if !isAtomicType(t) {
			allAtomicRemoved = false
		}
	}
	if len(kept) == 0 {
		if allAtomicRemoved {
			return TUnknown
		}
		return u
	}
	return Union{Types: kept, IsDirty: u.IsDirty}.Fold()
}

func foldTargetUnion(targets []DynType) DynType {
	if len(targets) == 0 {
		return TUnknown
	}
	if len(targets) == 1 {
		return targets[0]
	}
	return Union{Types: targets}.Fold()
}

func topLevelAssignedLocals(s ast.Statement) map[string]bool {
	out := map[string]bool{}
	for _, st := range stmtList(s) {
		if id, ok := unwrapStmtAssign(st); ok {
			out[id.Name] = true
		}
	}
	return out
}

func lastTopLevelAssignType(s ast.Statement, name string, env TypeEnv) DynType {
	var lastT DynType
	for _, st := range stmtList(s) {
		id, ok := unwrapStmtAssign(st)
		if !ok || id.Name != name {
			continue
		}
		if t := env.GetType(id); t != TUnknown {
			lastT = t
		}
	}
	return lastT
}

func topLevelAssignedMembers(s ast.Statement) map[string]bool {
	out := map[string]bool{}
	for _, st := range stmtList(s) {
		if name, _, ok := stmtMemberAssign(st); ok {
			out[name] = true
		}
	}
	return out
}

func lastTopLevelMemberAssignType(s ast.Statement, name string, env TypeEnv) DynType {
	var lastT DynType
	for _, st := range stmtList(s) {
		mname, b, ok := stmtMemberAssign(st)
		if !ok || mname != name {
			continue
		}
		if t := env.GetType(b.Rhs); t != TUnknown {
			lastT = t
		}
	}
	return lastT
}

func stmtMemberAssign(s ast.Statement) (string, *ast.Binary, bool) {
	es, ok := s.(*ast.ExprStmt)
	if !ok || es.E == nil {
		return "", nil, false
	}
	expr := es.E
	if ep, ok := expr.(*ast.ExprPos); ok && ep.Expr != nil {
		expr = ep.Expr
	}
	b, ok := expr.(*ast.Binary)
	if !ok || b.Tok != tok.Eq {
		return "", nil, false
	}
	name, _, ok := unwrapThisMember(b.Lhs)
	if !ok {
		return "", nil, false
	}
	return name, b, true
}

func unwrapStmtAssign(s ast.Statement) (*ast.Ident, bool) {
	es, ok := s.(*ast.ExprStmt)
	if !ok || es.E == nil {
		return nil, false
	}
	expr := es.E
	if ep, ok := expr.(*ast.ExprPos); ok && ep.Expr != nil {
		expr = ep.Expr
	}
	b, ok := expr.(*ast.Binary)
	if !ok || b.Tok != tok.Eq {
		return nil, false
	}
	id, ok := b.Lhs.(*ast.Ident)
	if !ok || isGlobalIdent(id.Name) {
		return nil, false
	}
	return id, true
}

func condEntryType(cond ast.Node, name string, env TypeEnv) DynType {
	var found *ast.Ident
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		if found != nil || n == nil {
			return
		}
		if id, ok := n.(*ast.Ident); ok && id.Name == name {
			found = id
			return
		}
		n.Children(func(c ast.Node) ast.Node {
			walk(c)
			return c
		})
	}
	walk(cond)
	if found == nil {
		return nil
	}
	return env.GetType(found)
}
