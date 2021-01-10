// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

// Package ast defines the node types
// used by the compiler to build syntax trees
package ast

import (
	"fmt"

	"github.com/apmckinlay/gsuneido/compile/lexer"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/str"
)

type Node interface {
	astNode()
	String() string
	// Children calls the given function for each child node
	Children(func(Node) Node)
}

type astNodeT struct{}

func (*astNodeT) astNode() {}

func (*astNodeT) Children(func(Node) Node) {
}

// Expr is implemented by expression nodes
type Expr interface {
	Node
	exprNode()
}
type exprNodeT struct {
	astNodeT
}

func (*exprNodeT) exprNode() {}

type Ident struct {
	exprNodeT
	Name string
	Pos  int32
}

func (a *Ident) String() string {
	return a.Name
}

func (a *Ident) ParamName() string {
	name := a.Name
	if name[0] == '.' {
		name = str.UnCapitalize(name[1:])
	} else if name[0] == '@' || name[0] == '_' {
		name = name[1:]
	}
	return name
}

type Constant struct {
	exprNodeT
	Val Value
}

func (a *Constant) String() string {
	return a.Val.String()
}

type Unary struct {
	exprNodeT
	Tok tok.Token
	E   Expr
}

func (a *Unary) String() string {
	return "Unary(" + a.Tok.String() + " " + a.E.String() + ")"
}

func (a *Unary) Children(fn func(Node) Node) {
	applyExpr(fn, &a.E)
}

type Binary struct {
	exprNodeT
	Lhs Expr
	Tok tok.Token
	Rhs Expr
}

func (a *Binary) String() string {
	return "Binary(" + a.Tok.String() + " " + a.Lhs.String() + " " + a.Rhs.String() + ")"
}

func applyExpr(fn func(Node) Node, pexpr *Expr) {
	if *pexpr != nil {
		*pexpr = fn(*pexpr).(Expr)
	}
}

func (a *Binary) Children(fn func(Node) Node) {
	applyExpr(fn, &a.Lhs)
	applyExpr(fn, &a.Rhs)
}

type Trinary struct {
	exprNodeT
	Cond Expr
	T    Expr
	F    Expr
}

func (a *Trinary) String() string {
	return "Trinary(" + a.Cond.String() + " " + a.T.String() + " " + a.F.String() + ")"
}

func (a *Trinary) Children(fn func(Node) Node) {
	applyExpr(fn, &a.Cond)
	applyExpr(fn, &a.T)
	applyExpr(fn, &a.F)
}

// Nary is used for associative binary operators e.g. add, multiply, and, or
type Nary struct {
	exprNodeT
	Tok   tok.Token
	Exprs []Expr
}

func (a *Nary) String() string {
	s := "Nary(" + a.Tok.String()
	for _, e := range a.Exprs {
		s += " " + e.String()
	}
	return s + ")"
}

func (a *Nary) Children(fn func(Node) Node) {
	for i := range a.Exprs {
		applyExpr(fn, &a.Exprs[i])
	}
}

type RangeTo struct {
	exprNodeT
	E    Expr
	From Expr
	To   Expr
}

func (a *RangeTo) String() string {
	return "RangeTo(" + a.E.String() + " " + fmt.Sprint(a.From) + " " + fmt.Sprint(a.To) + ")"
}

func (a *RangeTo) Children(fn func(Node) Node) {
	applyExpr(fn, &a.E)
	applyExpr(fn, &a.From)
	applyExpr(fn, &a.To)
}

type RangeLen struct {
	exprNodeT
	E    Expr
	From Expr
	Len  Expr
}

func (a *RangeLen) String() string {
	return "RangeLen(" + a.E.String() + " " + fmt.Sprint(a.From) + " " + fmt.Sprint(a.Len) + ")"
}

func (a *RangeLen) Children(fn func(Node) Node) {
	applyExpr(fn, &a.E)
	applyExpr(fn, &a.From)
	applyExpr(fn, &a.Len)
}

type Mem struct {
	exprNodeT
	E Expr
	M Expr
}

func (a *Mem) String() string {
	return "Mem(" + a.E.String() + " " + a.M.String() + ")"
}

func (a *Mem) Children(fn func(Node) Node) {
	applyExpr(fn, &a.E)
	applyExpr(fn, &a.M)
}

type In struct {
	exprNodeT
	E     Expr
	Exprs []Expr
}

func (a *In) String() string {
	s := "In(" + a.E.String() + " ["
	sep := ""
	for _, e := range a.Exprs {
		s += sep + e.String()
		sep = " "
	}
	return s + "])"
}

func (a *In) Children(fn func(Node) Node) {
	applyExpr(fn, &a.E)
	for i := range a.Exprs {
		applyExpr(fn, &a.Exprs[i])
	}
}

// InVals is used by queries
type InVals struct {
	exprNodeT
	E    Expr
	Vals []Value
}

func (a *InVals) String() string {
	s := "In(" + a.E.String() + " ["
	sep := ""
	for _, e := range a.Vals {
		s += sep + e.String()
		sep = " "
	}
	return s + "])"
}

func (a *InVals) Children(fn func(Node) Node) {
	applyExpr(fn, &a.E)
}

type Call struct {
	exprNodeT
	Fn   Expr
	Args []Arg
}

func (a *Call) String() string {
	s := "Call(" + a.Fn.String()
	for _, arg := range a.Args {
		s += " " + arg.String()
	}
	return s + ")"
}

func (a *Call) Children(fn func(Node) Node) {
	applyExpr(fn, &a.Fn)
	for i := range a.Args {
		applyExpr(fn, &a.Args[i].E)
	}
}

type Arg struct {
	Name Value // nil if not named
	E    Expr
}

func (a *Arg) String() string {
	s := ""
	if a.Name != nil {
		if ks, ok := a.Name.(SuStr); ok && lexer.IsIdentifier(string(ks)) {
			s += string(ks) + ":"
		} else {
			s += a.Name.String() + ":"
		}
	}
	return s + a.E.String()
}

type Function struct {
	exprNodeT
	Params      []Param
	Body        []Statement
	Final       map[string]int
	Base        Gnum
	Pos         int32
	HasBlocks   bool
	IsNewMethod bool
	// Id is set and used by codegen for block parents
	Id uint32
}

func (a *Function) String() string {
	return a.str("Function")
}

func (a *Function) str(which string) string {
	s := which + "("
	if len(a.Params) > 0 {
		sep := ""
		for _, p := range a.Params {
			if sep == "" && p.String() == "this" {
				continue
			}
			s += sep + p.String()
			sep = ","
		}
	}
	s += ""
	for _, stmt := range a.Body {
		if stmt != nil {
			s += "\n\t" + stmt.String()
		}
	}
	return s + ")"
}

func applyStmt(fn func(Node) Node, pstmt *Statement) {
	if *pstmt != nil {
		stmt := fn(*pstmt)
		if stmt == nil {
			*pstmt = nil
		} else {
			*pstmt = stmt.(Statement)
		}
	}
}

func (a *Function) Children(fn func(Node) Node) {
	for i := range a.Body {
		applyStmt(fn, &a.Body[i])
	}
}

type Param struct {
	Name   Ident // including prefix @ . _
	DefVal Value // may be nil
	// Unused is set if the parameter was followed by /*unused*/
	Unused bool
}

func (p *Param) String() string {
	s := p.Name.Name
	if p.DefVal != nil {
		s += "=" + p.DefVal.String()
	}
	return s
}

type Block struct {
	Function
	CompileAsFunction bool
}

func (a *Block) String() string {
	s := "Block"
	if a.CompileAsFunction {
		s += "-func"
	}
	return a.Function.str(s)
}

func (a *Block) Children(fn func(Node) Node) {
	a.Function = *fn(&a.Function).(*Function)
}

type Factory interface {
	Ident(name string, pos int32) Expr
	Constant(val Value) Expr
	Unary(tok tok.Token, expr Expr) Expr
	Binary(lhs Expr, tok tok.Token, rhs Expr) Expr
	Mem(e Expr, m Expr) Expr
	Trinary(cond Expr, t Expr, f Expr) Expr
	Nary(tok tok.Token, exprs []Expr) Expr
	In(e Expr, exprs []Expr) Expr
	Call(fn Expr, args []Arg) Expr
}

type Builder struct{}

func (Builder) Ident(name string, pos int32) Expr {
	return &Ident{Name: name, Pos: pos}
}
func (Builder) Constant(val Value) Expr {
	return &Constant{Val: val}
}
func (Builder) Unary(tok tok.Token, expr Expr) Expr {
	return &Unary{Tok: tok, E: expr}
}
func (Builder) Binary(lhs Expr, tok tok.Token, rhs Expr) Expr {
	return &Binary{Lhs: lhs, Tok: tok, Rhs: rhs}
}
func (Builder) Trinary(cond Expr, t Expr, f Expr) Expr {
	return &Trinary{Cond: cond, T: t, F: f}
}
func (Builder) Nary(tok tok.Token, exprs []Expr) Expr {
	return &Nary{Tok: tok, Exprs: exprs}
}
func (Builder) Mem(e Expr, m Expr) Expr {
	return &Mem{E: e, M: m}
}
func (Builder) In(e Expr, exprs []Expr) Expr {
	return &In{E: e, Exprs: exprs}
}
func (Builder) Call(fn Expr, args []Arg) Expr {
	return &Call{Fn: fn, Args: args}
}

var _ Factory = Builder{}

// Statement nodes implement the Stmt interface.
type Statement interface {
	Node
	Position() int
	SetPos(pos int)
	stmtNode()
}
type stmtNodeT struct {
	astNodeT
	Pos int
}

func (*stmtNodeT) stmtNode() {}
func (stmt *stmtNodeT) Position() int {
	return stmt.Pos
}
func (stmt *stmtNodeT) SetPos(pos int) {
	stmt.Pos = pos
}

type Compound struct {
	stmtNodeT
	Body []Statement
}

func (x *Compound) String() string {
	if len(x.Body) == 0 {
		return "{}"
	}
	if len(x.Body) == 1 {
		return x.Body[0].String()
	}
	s := "{\n"
	for _, stmt := range x.Body {
		if stmt != nil {
			s += stmt.String() + "\n"
		}
	}
	return s + "}"
}

func (x *Compound) Children(fn func(Node) Node) {
	for i := range x.Body {
		applyStmt(fn, &x.Body[i])
	}
}

type If struct {
	stmtNodeT
	Cond Expr
	Then Statement
	Else Statement // may be nil
}

func (x *If) String() string {
	s := "If(" + x.Cond.String() + " " + x.Then.String()
	if x.Else != nil {
		s += "\nelse " + x.Else.String()
	}
	return s + ")"
}

func (x *If) Children(fn func(Node) Node) {
	applyExpr(fn, &x.Cond)
	applyStmt(fn, &x.Then)
	applyStmt(fn, &x.Else)
}

type Return struct {
	stmtNodeT
	E Expr
}

func (x *Return) String() string {
	s := "Return("
	if x.E != nil {
		s += x.E.String()
	}
	return s + ")"
}

func (x *Return) Children(fn func(Node) Node) {
	applyExpr(fn, &x.E)
}

type Throw struct {
	stmtNodeT
	E Expr
}

func (x *Throw) String() string {
	return "Throw(" + x.E.String() + ")"
}

func (x *Throw) Children(fn func(Node) Node) {
	applyExpr(fn, &x.E)
}

type TryCatch struct {
	stmtNodeT
	Try            Statement
	CatchPos       int
	CatchVar       Ident
	CatchVarUnused bool
	CatchFilter    string
	Catch          Statement
}

func (x *TryCatch) String() string {
	s := "Try(" + x.Try.String()
	if x.Catch != nil {
		s += "\ncatch"
		if x.CatchVar.Name != "" {
			s += " (" + x.CatchVar.Name
			if x.CatchFilter != "" {
				s += ",'" + x.CatchFilter + "'"
			}
			s += ")"
		}
		s += " " + x.Catch.String()
	}
	return s + ")"
}

func (x *TryCatch) Children(fn func(Node) Node) {
	// TODO what about CatchVar ?
	applyStmt(fn, &x.Try)
	applyStmt(fn, &x.Catch)
}

type Forever struct {
	stmtNodeT
	Body Statement
}

func (x *Forever) String() string {
	return "Forever(" + x.Body.String() + ")"
}

func (x *Forever) Children(fn func(Node) Node) {
	applyStmt(fn, &x.Body)
}

type ForIn struct {
	stmtNodeT
	Var  Ident
	E    Expr
	Body Statement
}

func (x *ForIn) String() string {
	return "ForIn(" + x.Var.Name + " " + x.E.String() + "\n" + x.Body.String() + ")"
}

func (x *ForIn) Children(fn func(Node) Node) {
	// TODO what about Var ?
	applyExpr(fn, &x.E)
	applyStmt(fn, &x.Body)
}

type For struct {
	stmtNodeT
	Init []Expr
	Cond Expr
	Inc  []Expr
	Body Statement
}

func (x *For) String() string {
	s := "For("
	sep := ""
	for _, e := range x.Init {
		s += sep + e.String()
		sep = ","
	}
	s += "; "
	if x.Cond != nil {
		s += x.Cond.String()
	}
	s += "; "
	sep = ""
	for _, e := range x.Inc {
		s += sep + e.String()
		sep = ","
	}
	return s + "\n" + x.Body.String() + ")"
}

func (x *For) Children(fn func(Node) Node) {
	for i := range x.Init {
		applyExpr(fn, &x.Init[i])
	}
	applyExpr(fn, &x.Cond)
	for i := range x.Inc {
		applyExpr(fn, &x.Inc[i])
	}
	applyStmt(fn, &x.Body)
}

type While struct {
	stmtNodeT
	Cond Expr
	Body Statement
}

func (x *While) String() string {
	return "While(" + x.Cond.String() + " " + x.Body.String() + ")"
}

func (x *While) Children(fn func(Node) Node) {
	applyExpr(fn, &x.Cond)
	applyStmt(fn, &x.Body)
}

type DoWhile struct {
	stmtNodeT
	Body Statement
	Cond Expr
}

func (x *DoWhile) String() string {
	return "DoWhile(" + x.Body.String() + " " + x.Cond.String() + ")"
}

func (x *DoWhile) Children(fn func(Node) Node) {
	applyStmt(fn, &x.Body)
	applyExpr(fn, &x.Cond)
}

type Break struct {
	stmtNodeT
}

func (*Break) String() string {
	return "Break"
}

type Continue struct {
	stmtNodeT
}

func (*Continue) String() string {
	return "Continue"
}

type ExprStmt struct {
	stmtNodeT
	E Expr
}

func (x *ExprStmt) String() string {
	return x.E.String()
}

func (x *ExprStmt) Children(fn func(Node) Node) {
	applyExpr(fn, &x.E)
}

type Switch struct {
	stmtNodeT
	E       Expr
	Cases   []Case
	Default []Statement // may be nil
}

type Case struct {
	Exprs []Expr
	Body  []Statement
}

func (x *Switch) String() string {
	s := "Switch(" + x.E.String()
	for _, c := range x.Cases {
		s += "\nCase("
		sep := ""
		for _, e := range c.Exprs {
			s += sep + e.String()
			sep = ","
		}
		for _, stmt := range c.Body {
			if stmt != nil {
				s += "\n" + stmt.String()
			}
		}
		s += ")"
	}
	if x.Default != nil {
		if len(x.Default) == 0 {
			s += "\n()"
		}
		for _, stmt := range x.Default {
			if stmt != nil {
				s += "\n" + stmt.String()
			}
		}
	}
	return s + ")"
}

func (x *Switch) Children(fn func(Node) Node) {
	applyExpr(fn, &x.E)
	for i := range x.Cases {
		c := &x.Cases[i]
		for j := range c.Exprs {
			applyExpr(fn, &c.Exprs[j])
		}
		for j := range c.Body {
			applyStmt(fn, &c.Body[j])
		}
	}
	for i := range x.Default {
		applyStmt(fn, &x.Default[i])
	}
}
