// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import (
	"slices"

	. "github.com/apmckinlay/gsuneido/compile/ast"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
	"github.com/apmckinlay/gsuneido/util/assert"
)

// renameExpr renames identifiers in an expression.
// It does not modify the expression.
// If any renames are done, it returns a new expression.
// It is used by Where Transform
func renameExpr(expr Expr, from, to []string) Expr {
	switch e := expr.(type) {
	case *Constant:
		return expr
	case *Ident:
		// this is the actual rename
		// the other cases are just traversal and path copying
		if i := slices.Index(from, e.Name); i != -1 {
			return &Ident{Name: to[i]}
		}
		return expr
	case *Unary:
		newExpr := renameExpr(e.E, from, to)
		if newExpr == expr {
			return expr
		}
		return &Unary{Tok: e.Tok, E: newExpr}
	case *Binary:
		lhs := renameExpr(e.Lhs, from, to)
		rhs := renameExpr(e.Rhs, from, to)
		if lhs == e.Lhs && rhs == e.Rhs {
			return expr
		}
		return &Binary{Tok: e.Tok, Lhs: lhs, Rhs: rhs}
	case *Mem:
		e2 := renameExpr(e.E, from, to)
		m := renameExpr(e.M, from, to)
		if e2 == e.E && m == e.M {
			return expr
		}
		return &Mem{E: e2, M: m}
	case *Trinary:
		cond := renameExpr(e.Cond, from, to)
		t := renameExpr(e.T, from, to)
		f := renameExpr(e.F, from, to)
		if cond == e.Cond && t == e.T && f == e.F {
			return expr
		}
		return &Trinary{Cond: cond, T: t, F: f}
	case *RangeTo:
		cond := renameExpr(e.E, from, to)
		f := renameExpr(e.From, from, to)
		t := renameExpr(e.To, from, to)
		if cond == e.E && f == e.From && t == e.To {
			return expr
		}
		return &RangeTo{E: cond, From: f, To: t}
	case *RangeLen:
		cond := renameExpr(e.E, from, to)
		f := renameExpr(e.From, from, to)
		n := renameExpr(e.Len, from, to)
		if cond == e.E && f == e.From && n == e.Len {
			return expr
		}
		return &RangeLen{E: cond, From: f, Len: n}
	case *Nary:
		exprs := renameExprs(e.Exprs, from, to)
		if exprs == nil {
			return expr
		}
		return &Nary{Tok: e.Tok, Exprs: exprs}
	case *Call:
		fn := renameExpr(e.Fn, from, to)
		args := renameArgs(e.Args, from, to)
		if fn == e.Fn && args == nil {
			return expr
		}
		if args == nil {
			args = e.Args
		}
		return &Call{Fn: fn, Args: args}
	case *In:
		e2 := renameExpr(e.E, from, to)
		exprs := renameExprs(e.Exprs, from, to)
		if e2 == e.E && exprs == nil {
			return expr
		}
		if exprs == nil {
			exprs = e.Exprs
		}
		return &In{E: e2, Exprs: exprs}
	case *InRange:
		e2 := renameExpr(e.E, from, to)
		if e2 == e.E {
			return expr
		}
		return &InRange{E: e2, OrgTok: e.OrgTok, Org: e.Org,
			EndTok: e.EndTok, End: e.End}
	default:
		panic(assert.ShouldNotReachHere())
	}
}

func renameExprs(exprs []Expr, from, to []string) []Expr {
	var newExprs []Expr
	for i, e := range exprs {
		e2 := renameExpr(e, from, to)
		if e2 != e {
			if newExprs == nil {
				newExprs = make([]Expr, len(exprs))
				copy(newExprs, exprs[:i])
			}
		}
		if newExprs != nil {
			newExprs[i] = e2
		}
	}
	return newExprs
}

func renameArgs(args []Arg, from, to []string) []Arg {
	var newArgs []Arg
	for i, a := range args {
		e2 := renameExpr(a.E, from, to)
		if e2 != a.E {
			if newArgs == nil {
				newArgs = make([]Arg, len(args))
				copy(newArgs, args[:i])
			}
		}
		if newArgs != nil {
			newArgs[i] = Arg{E: e2, Name: a.Name}
		}
	}
	return newArgs
}

var aFolder Folder

// replaceExpr is used by Where Transform
func replaceExpr(expr Expr, from []string, to []Expr) Expr {
	switch e := expr.(type) {
	case *Constant:
		return expr
	case *Ident:
		// this is the actual replace
		// the other cases are just traversal and path copying
		if i := slices.Index(from, e.Name); i != -1 {
			return to[i]
		}
		return expr
	case *Unary:
		newExpr := replaceExpr(e.E, from, to)
		if newExpr == expr {
			return expr
		}
		return aFolder.Unary(e.Tok, newExpr)
	case *Binary:
		lhs := replaceExpr(e.Lhs, from, to)
		rhs := replaceExpr(e.Rhs, from, to)
		if lhs == e.Lhs && rhs == e.Rhs && !e.CouldEvalRaw() {
			return expr
		}
		// if it could be evaluated raw then we need to make a copy
		return aFolder.Binary(lhs, e.Tok, rhs)
	case *Mem:
		e2 := replaceExpr(e.E, from, to)
		m := replaceExpr(e.M, from, to)
		if e2 == e.E && m == e.M {
			return expr
		}
		return &Mem{E: e2, M: m}
	case *Trinary:
		cond := replaceExpr(e.Cond, from, to)
		t := replaceExpr(e.T, from, to)
		f := replaceExpr(e.F, from, to)
		if cond == e.Cond && t == e.T && f == e.F {
			return expr
		}
		return aFolder.Trinary(cond, t, f)
	case *RangeTo:
		cond := replaceExpr(e.E, from, to)
		f := replaceExpr(e.From, from, to)
		t := replaceExpr(e.To, from, to)
		if cond == e.E && f == e.From && t == e.To {
			return expr
		}
		return &RangeTo{E: cond, From: f, To: t}
	case *RangeLen:
		cond := replaceExpr(e.E, from, to)
		f := replaceExpr(e.From, from, to)
		n := replaceExpr(e.Len, from, to)
		if cond == e.E && f == e.From && n == e.Len {
			return expr
		}
		return &RangeLen{E: cond, From: f, Len: n}
	case *Nary:
		exprs := replaceExprs(e.Exprs, from, to)
		if exprs == nil {
			return expr
		}
		return aFolder.Nary(e.Tok, exprs)
	case *Call:
		fn := replaceExpr(e.Fn, from, to)
		args := replaceArgs(e.Args, from, to)
		if fn == e.Fn && args == nil {
			return expr
		}
		if args == nil {
			args = e.Args
		}
		return aFolder.Call(fn, args, 0)
	case *In:
		e2 := replaceExpr(e.E, from, to)
		exprs := replaceExprs(e.Exprs, from, to)
		if e2 == e.E && exprs == nil && !e.CouldEvalRaw() {
			return expr
		}
		if exprs == nil {
			exprs = e.Exprs
		}
		// if it could be evaluated raw then we need to make a copy
		return aFolder.In(e2, exprs)
	case *InRange:
		e2 := replaceExpr(e.E, from, to)
		if e2 == e.E && !e.CouldEvalRaw() {
			return expr
		}
		return aFolder.Nary(tok.And, []Expr{
			aFolder.Binary(e2, e.OrgTok, e.Org),
			aFolder.Binary(e2, e.EndTok, e.End)})
	default:
		panic(assert.ShouldNotReachHere())
	}
}

// replaceExprs returns nil if nothing was replaced,
// otherwise it returns a modified copy of the expression list
func replaceExprs(exprs []Expr, from []string, to []Expr) []Expr {
	var newExprs []Expr
	for i, e := range exprs {
		e2 := replaceExpr(e, from, to)
		if e2 != e {
			if newExprs == nil {
				newExprs = make([]Expr, len(exprs))
				copy(newExprs, exprs[:i])
			}
		}
		if newExprs != nil {
			newExprs[i] = e2
		}
	}
	return newExprs
}

func replaceArgs(args []Arg, from []string, to []Expr) []Arg {
	var newArgs []Arg
	for i, a := range args {
		e2 := replaceExpr(a.E, from, to)
		if e2 != a.E {
			if newArgs == nil {
				newArgs = make([]Arg, len(args))
				copy(newArgs, args[:i])
			}
		}
		if newArgs != nil {
			newArgs[i] = Arg{E: e2, Name: a.Name}
		}
	}
	return newArgs
}
