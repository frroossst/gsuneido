// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package compile

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"

	"os"

	"github.com/apmckinlay/gsuneido/compile/ast"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/util/ascii"
)

// Constant compiles an anonymous Suneido constant
func Constant(src string) Value {
	return NamedConstant("", "", src, nil)
}

// NamedConstant compiles a Suneido constant with a name
// e.g. a library record
func NamedConstant(lib, name, src string, prevDef Value) Value {
	p := NewParser(src)
	p.lib = lib
	p.name = name
	p.prevDef = prevDef
	result := p.constant()
	if p.Token != tok.Eof {
		p.Error("did not parse all input")
	}
	return result
}

func Checked(th *Thread, src string) (Value, []string) {
	// can't do AST check after compile because that would miss nested functions
	p := CheckParser(src, th)
	v := p.constant()
	if p.Token != tok.Eof {
		p.Error("did not parse all input")
	}
	return v, p.CheckResults()
}

type Node_t struct {
	// the branch of the node (e.g. Binary, Unary, etc.)
	Tag string
	// the text Value from the parser
	Value string
	// the type of the node (e.g. string, number, etc.)
	Type_t string
	// Args
	Args []Node_t
	// UUID for downstream processing
	ID string
}

type Function_t struct {
	Node_t
	// the Name of the function
	Name string
	// the Parameters of the function
	Parameters []string
	// the Body of the function
	Body [][]Node_t
}

func (p *Parser) TypeFunction() Function_t {
	return p.typeFunction()
}

type Class_t struct {
	Node_t
	// the Name of the class
	Name string
	// the Base class of the class
	Base string
	// the Methods of the class
	Methods map[string][]Function_t
	// the Attributes of the class
	Attributes map[string][]Node_t
}

func (p *Parser) TypeClass() Class_t {
	return p.typeClass()
}

func (p *Parser) TypeConst() []Node_t {
	type_arr := []Node_t{}
	for p.Token != tok.Eof {
		// type_arr = append(type_arr, p.typeConst())
		// p.typeConst() returns an array, merge the arrays
		type_arr = append(type_arr, p.typeConst()...)
	}
	return type_arr
}

func (p *Parser) typeConst() []Node_t {
	switch p.Token {
	case tok.String:
		content := p.Text
		p.string()
		return []Node_t{{Tag: "Constant", Type_t: "String", Value: content, ID: uuid.New().String()}}
	case tok.Number:
		content := p.Text
		p.number()
		return []Node_t{{Tag: "Constant", Type_t: "Number", Value: content, ID: uuid.New().String()}}
	case tok.Identifier:
		content := p.Text
		p.MatchIdent()
		return []Node_t{{Tag: "Identifier", Type_t: "Variable", Value: content, ID: uuid.New().String()}}
	case tok.Eq:
		content := p.Text
		p.Match(tok.Eq)
		return []Node_t{{Tag: "Operator", Type_t: "Operator", Value: content, ID: uuid.New().String()}}
	case tok.Add:
		content := p.Text
		p.Match(tok.Add)
		return []Node_t{{Tag: "Operator", Type_t: "Operator", Value: content, ID: uuid.New().String()}}
	case tok.Sub:
		content := p.Text
		p.Match(tok.Sub)
		return []Node_t{{Tag: "Operator", Type_t: "Operator", Value: content, ID: uuid.New().String()}}
	default:
		panic(p.Error("invalid constant, unexpected " + p.Token.String()))
	}
}

func (p *Parser) typeFunction() Function_t {
	ast := p.WithoutKeywordFunction()
	parameters := []string{}
	for i := 0; i < len(ast.Params); i++ {
		parameters = append(parameters, ast.Params[i].Name.Name)
	}

	body := [][]Node_t{}
	for i := 0; i < len(ast.Body); i++ {
		body = append(body, getExprType(ast.Body[i]))
	}

	// mark anonymous functions with empty name
	return Function_t{Node_t: Node_t{Tag: "Function", Type_t: "Function", Value: "YW5vbnltb3Vz"}, Name: "", Parameters: parameters, Body: body}
}

func getExprType(expr ast.Statement) []Node_t {
	// coerce ast.Expr to ast.Node
	node := expr.(ast.Node)

	id := uuid.New().String()

	// node matches with *ast.ExprStmt in a switch-case
	// convert node to type node.Expr
	switch t := node.(type) {
	case *ast.ExprStmt:
		basic_expr := node.(*ast.ExprStmt).E
		return getNodeType(basic_expr)
	case *ast.If:
		if_stmt := node.(*ast.If)
		cond := getNodeType(if_stmt.Cond)
		then := getNodeType(if_stmt.Then)
		if if_stmt.Else != nil {
			els := getNodeType(if_stmt.Else)
			return []Node_t{{Tag: "If", Type_t: "If", Value: "nil", Args: append(append(cond, then...), els...), ID: id}}
		}
		return []Node_t{{Tag: "If", Type_t: "If", Value: "nil", Args: append(cond, then...), ID: id}}
	case *ast.Return:
		stmt := t.E
		return_stmt := getNodeType(stmt)
		return []Node_t{{Tag: "Return", Type_t: "Return", Value: "nil", Args: return_stmt, ID: id}}
	default:
		fmt.Println(reflect.TypeOf(node))
		fmt.Println(t.String())
		panic("not implemented in getExprType " + node.String())
	}
}

func getNodeType(node ast.Node) []Node_t {

	id := uuid.New().String()

	switch n := node.(type) {
	case *ast.Unary:
		una := getNodeType(n.E)
		ops := n.Tok
		return []Node_t{{Tag: "Unary", Type_t: ops.String(), Value: ops.String(), Args: una, ID: id}}
	case *ast.Binary:
		lhs := getNodeType(n.Lhs)
		rhs := getNodeType(n.Rhs)
		ops := n.Tok
		return []Node_t{{Tag: "Binary", Type_t: "Operator", Value: ops.String(), Args: append(lhs, rhs...), ID: id}}
	case *ast.Nary:
		args := []Node_t{}
		ops := n.Tok
		for i := 0; i < len(n.Exprs); i++ {
			args = append(args, getNodeType(n.Exprs[i])...)
		}
		return []Node_t{{Tag: "Nary", Type_t: "Operator", Value: ops.String(), Args: args, ID: id}}
	case *ast.Compound:
		cmps := []Node_t{}
		for i := 0; i < len(n.Body); i++ {
			cmps = append(cmps, getExprType(n.Body[i])...)
		}
		return []Node_t{{Tag: "Compound", Type_t: "Compound", Value: "nil", Args: cmps, ID: id}}
	case *ast.Ident:
		return []Node_t{{Tag: "Identifier", Type_t: "Variable", Value: n.Name, ID: id}}
	case *ast.Constant:
		return []Node_t{{Tag: "Constant", Type_t: n.Val.Type().String(), Value: n.Val.String(), ID: id}}
	case *ast.Call:
		var value string
		switch v := n.Fn.(type) {
		case *ast.Ident:
			value = v.Name
		case *ast.Mem:
			// remove double quoted strings from the following
			noqute := v.M.String()[1 : len(v.M.String())-1]
			value = v.E.(*ast.Ident).Name + "." + noqute
		default:
			fmt.Println(reflect.TypeOf(v))
			fmt.Fprintln(os.Stdout, []any{v}...)
			panic("not implemented in getNodeType " + v.String())
		}
		return []Node_t{{Tag: "Call", Type_t: "Operator", Value: "nil",
			Args: []Node_t{{Tag: "Identifier", Type_t: "Callable", Value: value, ID: id}}}}
	case *ast.Mem:
		noqute := n.M.String()[1 : len(n.M.String())-1]
		return []Node_t{{Tag: "Member", Type_t: "Member", Value: noqute, ID: id}}
	default:
		fmt.Println(reflect.TypeOf(n))
		panic("not implemented in getNodeType " + n.String())
	}
}

func (p *Parser) Const() (result Value) {
	defer func(org int32) {
		if r, ok := result.(iSetPos); ok {
			SetPos(r, org, p.EndPos)
		}
	}(p.Pos)
	return p.constant()
}

func (p *Parser) constant() Value {
	switch p.Token {
	case tok.String:
		return p.string()
	case tok.Symbol:
		s := SuStr(p.Text)
		p.Next()
		return s
	case tok.Add:
		p.Next()
		fallthrough
	case tok.Number:
		return p.number()
	case tok.Sub:
		p.Next()
		return OpUnaryMinus(p.number())
	case tok.LParen, tok.LCurly, tok.LBracket:
		return p.object()
	case tok.Hash:
		p.Next()
		switch p.Token {
		case tok.Number:
			return p.date()
		case tok.LParen, tok.LCurly, tok.LBracket:
			return p.object()
		}
		panic("not implemented #" + p.Text)
	case tok.True:
		p.Next()
		return True
	case tok.False:
		p.Next()
		return False
	case tok.Function:
		return p.functionValue()
	case tok.Class:
		return p.class()
	default:
		if p.Token.IsIdent() {
			if okBase(p.Text) && p.Lxr.AheadSkip(0).Token == tok.LCurly {
				return p.class()
			}
			if p.Lxr.Ahead(0).Token != tok.Colon &&
				(p.Text == "struct" || p.Text == "dll" || p.Text == "callback") {
				p.Error("gSuneido does not implement " + p.Text)
			}
			s := p.Text
			p.Next()
			return SuStr(s)
		}
	}
	panic(p.Error("invalid constant, unexpected " + p.Token.String()))
}

func (p *Parser) functionValue() Value {
	prevClassName := p.className
	p.className = "" // prevent privatization in standalone function
	ast := p.Function()
	p.className = prevClassName
	p.CheckFunc(ast)
	return p.codegen(p.lib, p.name, ast, p.prevDef)
}

// string handles compile time concatenation
func (p *Parser) string() Value {
	s := p.Text
	p.Match(tok.String)
	if !p.moreStr() {
		return SuStr(s) // normal case
	}
	strs := []string{s}
	for p.moreStr() {
		p.Match(tok.Cat)
		strs = append(strs, p.Text)
		p.Match(tok.String)
	}
	return p.mkConcat(strs)
}

func (p *Parser) moreStr() bool {
	return p.Token == tok.Cat && p.Lxr.AheadSkip(0).Token == tok.String
}

func (p *Parser) number() Value {
	s := p.Text
	p.Match(tok.Number)
	return NumFromString(s)
}

func (p *Parser) date() Value {
	s := p.Text
	p.Match(tok.Number)
	date := DateFromLiteral(s)
	if date == NilDate {
		p.Error("bad date literal ", s)
	}
	return date
}

var closing = map[tok.Token]tok.Token{
	tok.LParen:   tok.RParen,
	tok.LCurly:   tok.RCurly,
	tok.LBracket: tok.RBracket,
}

const noBase = -1

func (p *Parser) object() Value {
	close := closing[p.Token]
	p.Next()
	var ob container
	if close == tok.RParen {
		ob = p.mkObject()
	} else {
		ob = p.mkRecord()
	}
	p.memberList(ob, close, noBase)
	if close == tok.RBracket {
		ob = p.mkRecOrOb(ob)
	}
	if p, ok := ob.(protectable); ok {
		p.SetReadOnly()
	}
	return ob.(Value)
}

type protectable interface {
	SetReadOnly()
}

func (p *Parser) memberList(ob container, closing tok.Token, base Gnum) {
	for p.Token != closing {
		pos := p.Item.Pos
		k, v := p.member(closing, base)
		if p.Token == tok.Comma || p.Token == tok.Semicolon {
			p.Next()
		}
		if k == nil {
			p.set(ob, nil, v, pos, p.EndPos)
		} else {
			p.putMem(ob, k, v, pos)
		}
	}
	p.Next()
}

func (p *Parser) member(closing tok.Token, base Gnum) (k Value, v Value) {
	start := p.Token
	m := p.constant() // might be key or value
	inClass := base != noBase
	if inClass && start.IsIdent() && p.Token == tok.LParen { // method
		name := p.privatizeDef(m)
		prevName := p.name
		p.name += "." + name
		ast := p.function(true)
		ast.Base = base
		if name == "New" {
			ast.IsNewMethod = true
		}
		p.CheckFunc(ast)
		fn := p.codegen(p.lib, p.name, ast, p.prevDef)
		p.name = prevName
		if f, ok := fn.(*SuFunc); ok {
			f.ClassName = p.className
		}
		return SuStr(name), fn
	}
	if p.MatchIf(tok.Colon) { // named
		if inClass {
			m = SuStr(p.privatizeDef(m))
		}
		if p.Token == tok.Comma || p.Token == tok.Semicolon || p.Token == closing {
			return m, True
		}
		prevName := p.name
		if s, ok := m.ToStr(); ok {
			p.name += "." + s
		}
		c := p.constant()
		p.name = prevName
		return m, c
	}
	return nil, m
}

func (p *Parser) privatizeDef(m Value) string {
	ss, ok := m.(SuStr)
	if !ok {
		p.Error("class member names must be strings")
	}
	name := string(ss)
	if strings.HasPrefix(name, "Getter_") &&
		len(name) > 7 && !ascii.IsUpper(name[7]) {
		p.Error("invalid getter (" + name + ")")
	}
	if !ascii.IsLower(name[0]) {
		return name
	}
	if strings.HasPrefix(name, "getter_") &&
		(len(name) <= 7 || !ascii.IsLower(name[7])) {
		p.Error("invalid getter (" + name + ")")
	}
	return p.privatize(name, p.className)
}

// putMem checks for duplicate member and then calls p.set with endpos
func (p *Parser) putMem(ob container, m Value, v Value, pos int32) {
	if ob.HasKey(m) {
		p.ErrorAt(pos, "duplicate member name ("+m.String()+")")
	} else {
		p.set(ob, m, v, pos, p.EndPos)
	}
}

// returns key value pair of node types for the AST
func (p *Parser) typeClass() Class_t {
	if p.Token == tok.Class {
		p.Match(tok.Class)
		if p.Token == tok.Colon {
			p.Match(tok.Colon)
		}
	}
	baseName := "class"
	if p.Token.IsIdent() {
		baseName = p.Text
		p.MatchIdent()
	}
	p.Match(tok.LCurly)
	prevClassName := p.className

	// class members are stored as key value pairs separated by `:`
	// e.g. `a: 1, b: 2`
	// the key is the name of the member
	// the value is the type of the member
	// class methods are represented as functions
	// they are not `:` separated
	// Foo(x, y) { x + y }
	// bar() { 1 }

	// construct kv store where key is string and value is Node_t
	kv_store_methods := map[string][]Function_t{}
	kv_store_attrbts := map[string][]Node_t{}

	// parse class members
	for p.Token != tok.RCurly {
		// parse member name
		member_name := p.Text

		isFunc := true

		// check by looking ahead if member is a function
		if p.Lxr.AheadSkip(0).Token == tok.Colon {
			isFunc = false
			p.MatchIdent()
			p.Match(tok.Colon)
		}

		if isFunc {
			func_name := p.MatchIdent()
			func_type := p.typeFunction()
			func_node := Function_t{Node_t: Node_t{Tag: "Function", Type_t: "Function", Value: ""}, Name: func_name, Parameters: func_type.Parameters, Body: func_type.Body}

			kv_store_methods[func_name] = []Function_t{func_node}

		} else {
			// parse member type
			member_type := p.typeConst()
			// kv_store[member_name] = member_type[0]
			kv_store_attrbts[member_name] = []Node_t{member_type[0]}
		}

	}

	p.className = p.getClassName()
	p.className = prevClassName

	fmt.Println("=== AST ===")

	return Class_t{Node_t: Node_t{Tag: "Class", Type_t: "Class", Value: "nil"}, Name: p.name, Base: baseName, Methods: kv_store_methods, Attributes: kv_store_attrbts}
}

// classNum is used to generate names for anonymous classes
var classNum atomic.Int32

// class parses a class definition
// like object, it builds a value rather than an ast
func (p *Parser) class() (result Value) {
	if p.Token == tok.Class {
		p.Match(tok.Class)
		if p.Token == tok.Colon {
			p.Match(tok.Colon)
		}
	}
	var base Gnum
	baseName := "class"
	if p.Token.IsIdent() {
		baseName = p.Text
		base = p.ckBase(baseName)
		p.MatchIdent()
	}
	pos1 := p.EndPos
	p.Match(tok.LCurly)
	pos2 := p.EndPos
	prevClassName := p.className
	p.className = p.getClassName()
	mems := p.mkClass(baseName)
	p.memberList(mems, tok.RCurly, base)
	p.setPos(mems, pos1, pos2)
	p.className = prevClassName
	if cc, ok := mems.(classBuilder); ok {
		return &SuClass{Base: base, Lib: p.lib, Name: p.name,
			MemBase: MemBase{Data: cc}}
	}
	return mems.(Value)
}

func (p *Parser) ckBase(name string) Gnum {
	if !okBase(name) {
		p.Error("base class must be global defined in library, got: ", name)
	}
	if name[0] == '_' {
		if name == "_" || name[1:] != p.name {
			p.Error("invalid reference to " + name)
		}
		return Global.Overload(name, p.prevDef)
		// for _Name in expressions see codegen.go cgen.identifier
	}
	p.CheckGlobal(name, int(p.Pos))
	return Global.Num(name)
}

func okBase(name string) bool {
	return ascii.IsUpper(name[0]) ||
		(name[0] == '_' && len(name) > 1 && ascii.IsUpper(name[1]))
}

func (p *Parser) getClassName() string {
	last := p.name
	i := strings.LastIndexAny(last, " .")
	if i != -1 {
		last = last[i+1:]
	}
	if last == "" || last == "?" {
		cn := classNum.Add(1)
		className := "Class" + strconv.Itoa(int(cn))
		return className
	}
	return last
}
