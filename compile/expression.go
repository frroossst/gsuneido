package compile

import (
	"github.com/apmckinlay/gsuneido/compile/ast"
	. "github.com/apmckinlay/gsuneido/lexer"
	. "github.com/apmckinlay/gsuneido/runtime"
	. "github.com/apmckinlay/gsuneido/util/ascii"
)

// expression parses a Suneido expression and builds an AST
func (p *parser) expr() ast.Expr {
	return p.pcExpr(1)
}

// pcExpr implements precedence climbing
// each call processes at least one atom
// a given call processes everything >= minprec
// it recurses to process the right hand side of each operator
func (p *parser) pcExpr(minprec int8) ast.Expr {
	e := p.atom()
	// fmt.Println("pcExpr minprec", minprec, "atom", e)
	for p.Token != EOF {
		tok := p.Token
		prec := precedence[tok]
		// fmt.Println("loop ", p.Item, "prec", prec)
		if prec < minprec {
			break
		}
		if p.newline {
			break
		}
		p.next()
		switch {
		case tok == DOT:
			id := p.Text
			p.matchIdent()
			e = p.Mem(e, p.Constant(SuStr(id)))
			if p.Token == L_CURLY && !p.expectingCompound { // a.F { }
				e = p.Call(e, p.arguments(p.Token))
			}
		case tok == INC || tok == DEC: // postfix
			ckLvalue(e)
			e = p.Unary(tok+1, e) // +1 must be POSTINC/DEC
		case tok == IN:
			e = p.in(e)
		case tok == NOT:
			p.match(IN)
			e = p.Unary(NOT, p.in(e))
		case tok == L_BRACKET:
			var expr ast.Expr
			if p.Token == RANGETO || p.Token == RANGELEN {
				expr = nil
			} else {
				expr = p.expr()
			}
			if p.Token == RANGETO || p.Token == RANGELEN {
				rtype := p.Token
				p.next()
				var expr2 ast.Expr
				if p.Token == R_BRACKET {
					expr2 = nil
				} else {
					expr2 = p.expr()
				}
				if rtype == RANGETO {
					e = &ast.RangeTo{E: e, From: expr, To: expr2}
				} else {
					e = &ast.RangeLen{E: e, From: expr, Len: expr2}
				}
			} else {
				e = p.Mem(e, expr)
			}
			p.match(R_BRACKET)
		case ASSIGN_START < tok && tok < ASSIGN_END:
			ckLvalue(e)
			rhs := p.expr()
			if tok == EQ {
				if id, ok := e.(*ast.Ident); ok {
					if c, ok := rhs.(*ast.Constant); ok {
						if named, ok := c.Val.(Named); ok {
							named.SetName(id.Name)
						}
					}
				}
			}
			e = p.Binary(e, tok, rhs)
		case tok == Q_MARK:
			t := p.expr()
			p.match(COLON)
			f := p.expr()
			e = p.Trinary(e, t, f)
		case tok == L_PAREN: // function call
			e = p.Call(e, p.arguments(tok))
		case ASSOC_START < tok && tok < ASSOC_END:
			// for associative operators, collect a list of contiguous
			es := []ast.Expr{e}
			listtype := flip(tok)
			for {
				rhs := p.pcExpr(prec + 1) // +1 for left associative
				// invert SUB and DIV to combine as ADD and MUL
				if tok == SUB || tok == DIV {
					rhs = p.Unary(tok, rhs)
				}
				es = append(es, rhs)

				tok = p.Token
				if !p.same(listtype, tok) {
					break
				}
				p.next()
			}
			e = p.Nary(listtype, es)
		default: // other left associative binary operators
			rhs := p.pcExpr(prec + 1) // +1 for left associative
			e = p.Binary(e, tok, rhs)
		}
	}
	return e
}

func (p *parser) in(e ast.Expr) ast.Expr {
	list := []ast.Expr{}
	p.match(L_PAREN)
	for p.Token != R_PAREN {
		list = append(list, p.expr())
		if p.Token == COMMA {
			p.next()
		}
	}
	p.match(R_PAREN)
	return p.In(e, list)
}

func ckLvalue(e ast.Expr) {
	switch e := e.(type) {
	case *ast.Mem:
		return
	case *ast.Ident:
		if isLocal(e.Name) {
			return
		}
	}
	panic("syntax error: lvalue required")
}

func flip(tok Token) Token {
	switch tok {
	case SUB:
		return ADD
	case DIV:
		return MUL
	default:
		return tok
	}
}

func (p *parser) same(listtype Token, next Token) bool {
	if p.newline {
		return false
	}
	return next == listtype ||
		(next == SUB && listtype == ADD) || (next == DIV && listtype == MUL)
}

func (p *parser) atom() ast.Expr {
	switch tok := p.Token; tok {
	case TRUE, FALSE, NUMBER, STRING, HASH:
		return p.Constant(p.constant())
	case L_PAREN:
		p.next()
		e := p.expr()
		p.match(R_PAREN)
		return e
	case L_CURLY:
		return p.block()
	case L_BRACKET:
		return p.record()
	case ADD, SUB, NOT, BITNOT:
		p.next()
		return p.Unary(tok, p.pcExpr(precedence[L_PAREN]))
	case INC, DEC:
		p.next()
		e := p.pcExpr(precedence[DOT])
		ckLvalue(e)
		return p.Unary(tok, e)
	case DOT: // unary, i.e. implicit "this"
		// does not absorb DOT
		p.newline = false
		return p.Ident("this")
	case FUNCTION:
		return p.Constant(codegen(p.function()))
	case CLASS:
		return p.Constant(p.class())
	case NEW:
		p.next()
		expr := p.pcExpr(precedence[DOT])
		var args []ast.Arg
		if p.matchIf(L_PAREN) {
			args = p.arguments(L_PAREN)
		} else {
			args = []ast.Arg{}
		}
		expr = p.Mem(expr, p.Constant(SuStr("*new*")))
		return p.Call(expr, args)
	default:
		if IsIdent[p.Token] {
			// MyClass { ... } => class
			if !p.expectingCompound &&
				okBase(p.Text) && p.lxr.AheadSkip(0).Token == L_CURLY {
				return p.Constant(p.class())
			}
			e := p.Ident(p.Text)
			p.next()
			return e
		}
	}
	panic(p.error("syntax error: unexpected " + p.Item.String()))
}

var precedence = [Ntokens]int8{
	Q_MARK:    2,
	OR:        3,
	AND:       4,
	IN:        5,
	NOT:       5, // not in
	BITOR:     6,
	BITXOR:    7,
	BITAND:    8,
	IS:        9,
	ISNT:      9,
	MATCH:     9,
	MATCHNOT:  9,
	LT:        10,
	LTE:       10,
	GT:        10,
	GTE:       10,
	LSHIFT:    11,
	RSHIFT:    11,
	CAT:       12,
	ADD:       12,
	SUB:       12,
	MUL:       13,
	DIV:       13,
	MOD:       13,
	INC:       14,
	DEC:       14,
	L_PAREN:   15,
	DOT:       16,
	L_BRACKET: 16,
	EQ:        16,
	ADDEQ:     16,
	SUBEQ:     16,
	CATEQ:     16,
	MULEQ:     16,
	DIVEQ:     16,
	MODEQ:     16,
	LSHIFTEQ:  16,
	RSHIFTEQ:  16,
	BITOREQ:   16,
	BITANDEQ:  16,
	BITXOREQ:  16,
}

var call = Item{Text: "call"}

func (p *parser) arguments(opening Token) []ast.Arg {
	var args []ast.Arg
	if opening == L_PAREN {
		if p.matchIf(AT) {
			return p.atArgument()
		}
		args = p.argumentList(R_PAREN)
	}
	if p.Token == L_CURLY && !p.expectingCompound {
		args = append(args, ast.Arg{Name: blockArg, E: p.block()})
	}
	return args
}

var atArg = SuStr("@")
var at1Arg = SuStr("@+1")
var blockArg = SuStr("block")

func (p *parser) atArgument() []ast.Arg {
	which := atArg
	if p.matchIf(ADD) {
		if p.Item.Text != "1" {
			panic("only @+1 is supported")
		}
		p.match(NUMBER)
		which = at1Arg
	}
	expr := p.expr()
	p.match(R_PAREN)
	return []ast.Arg{ast.Arg{Name: which, E: expr}}
}

func (p *parser) argumentList(closing Token) []ast.Arg {
	var args []ast.Arg
	haveNamed := false
	unnamed := func(val ast.Expr) {
		if haveNamed {
			p.error("un-named arguments must come before named arguments")
		}
		args = append(args, ast.Arg{E: val})
	}
	named := func(name Value, val ast.Expr) {
		for _, a := range args {
			if name.Equal(a.Name) {
				p.error("duplicate argument name (" + name.String() + ")")
			}
		}
		args = append(args, ast.Arg{Name: name, E: val})
		haveNamed = true
	}
	var pending Value
	handlePending := func(val ast.Expr) {
		if pending != nil {
			named(pending, val)
			pending = nil
		}
	}
	for p.Token != closing {
		var expr ast.Expr
		if p.matchIf(COLON) {
			if !IsLower(p.Text[0]) {
				p.error("expecting local variable name")
			}
			handlePending(p.Constant(True))
			named(SuStr(p.Text), p.Ident(p.Text))
			p.matchIdent()
		} else {
			expr = p.expr() // could be name or value
			if name := p.argname(expr); name != nil && p.matchIf(COLON) {
				handlePending(p.Constant(True))
				pending = name // it's a name but don't know value yet
			} else if pending != nil {
				handlePending(expr)
			} else {
				unnamed(expr)
			}
		}
		if p.matchIf(COMMA) {
			handlePending(p.Constant(True))
		}
	}
	p.match(closing)
	handlePending(p.Constant(True))
	return args
}

func (p *parser) argname(expr ast.Expr) Value {
	// FIXME: queries won't be same ast node types
	if id, ok := expr.(*ast.Ident); ok {
		return SuStr(id.Name)
	}
	if c, ok := expr.(*ast.Constant); ok {
		return c.Val
	}
	return nil
}

func (p *parser) record() ast.Expr {
	p.match(L_BRACKET)
	args := p.argumentList(R_BRACKET)
	return p.Call(p.Ident("Record"), args)
}

func (p *parser) block() *ast.Block {
	p.match(L_CURLY)
	params := p.blockParams()
	body := p.statements()
	p.match(R_CURLY)
	return &ast.Block{ast.Function{Params: params, Body: body}}
}

func (p *parser) blockParams() []ast.Param {
	var params []ast.Param
	if p.matchIf(BITOR) {
		if p.matchIf(AT) {
			params = append(params, ast.Param{Name: "@" + p.Text})
			p.matchIdent()
		} else {
			for IsIdent[p.Token] {
				params = append(params, ast.Param{Name: p.Text})
				p.matchIdent()
				p.matchIf(COMMA)
			}
		}
		p.match(BITOR)
	}
	return params
}
