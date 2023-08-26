package main

import (
	"fmt"

	"github.com/apmckinlay/gsuneido/compile"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
	. "github.com/apmckinlay/gsuneido/runtime"
)

func generateAST(src string) Value {
//	src := ToStr(args[0])
	p := compile.AstParser(src)
	ast := p.Const()
	if p.Token != tok.Eof {
		p.Error("did not parse all input")
	}
	return ast
}

func traverseAST(ast Value) {
	// traverse the AST
}

type ReturnType struct {
	Types []tok.Token
}

func newReturnType(toks []tok.Token) *ReturnType {
	return &ReturnType{Types: toks}
}

func toString(r *ReturnType) string {
	strBuild := ""
	for _, i := range r.Types {
		strBuild += i.String()
		strBuild += " | "
	}
	// remove suffix |
	strBuild = strBuild[:len(strBuild)-3]
	return strBuild
}

// capture all explicit return types
func assembleReturnTypes(p *compile.Parser) []tok.Token {
	explicitReturns := []tok.Token{}
	for ;; {
		prevTok := p.Token
		if p.Token == tok.Eof {
			break
		} else if prevTok == tok.Return {
			p.Next()
			fmt.Println("return -> ", p.Token)
			explicitReturns = append(explicitReturns, p.Token)
		}
		p.Next()
	}
	return explicitReturns
}

func constraintReturnTypes(p *compile.Parser) {

}

func containsElement(arr []tok.Token, target tok.Token) bool {
    for _, item := range arr {
        if item == target {
            return true
        }
    }
    return false
}

func removeElement(arr []tok.Token, target tok.Token) []tok.Token {
	for i, item := range arr {
		if item != target {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func deUnioniseReturnTypes (rt* ReturnType) []tok.Token {
	// remove duplicates
	typeSet := []tok.Token{}

	for _, i := range rt.Types {
		if containsElement(typeSet, i) {
			typeSet = append(typeSet, i)
		}
	}

	// check bool; remove true and false replace with bool
	hasTrue := containsElement(typeSet, tok.True)
	hasFalse := containsElement(typeSet, tok.False)

	if hasTrue && hasFalse {
		typeSet = removeElement(typeSet, tok.False)
		typeSet = removeElement(typeSet, tok.True)
		typeSet = append(typeSet, tok.Bool)
	}

	return typeSet
}

func main() {
	// Parse processes the command line options
	// returning the remaining arguments

	// generate AST
	code := "function() { if true { return 1 } else if  { return '2' } else if { return a } else { return true } }"
//	ast := generateAST(code)
//	fmt.Println(ast)

	p := compile.AstParser(code)
	
	// First Pass
	// get explicit return types
	explicitReturns := assembleReturnTypes(p)
	rt := newReturnType(explicitReturns)
//	fmt.Println(toString(rt))
	// deunionise return types; remove duplicates, identifiers etc.
	du := deUnioniseReturnTypes(rt)
	fmt.Println(du)
}
