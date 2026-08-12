package engine

// narrowing across switch statements, in both cond mode (case exprs are
// predicates) and scrutinee mode (case exprs are literals to match).

import "github.com/apmckinlay/gsuneido/compile/ast"

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
