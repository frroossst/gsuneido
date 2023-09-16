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

type type_t uint8
const (
	Number_t type_t = iota
	String_t
	Any_t
	Bool_t
	True_t
	False_t
	Object_t
	Function_t
	Instantce_t
	Unknown_t
)


// This function returns a string representation of the given type_t slice.
func TypeTString(t []type_t) string {
	str := ""
	for _, v := range t {
		switch v {
		case Number_t:
			str += "Number "
		case String_t:
			str += "String "
		case Any_t:
			str += "Any "
		case Bool_t:
			str += "Bool "
		case True_t:
			str += "True "
		case False_t:
			str += "False "
		case Object_t:
			str += "Object "
		case Function_t:
			str += "Function "
		case Instantce_t:
			str += "Instance "
		case Unknown_t:
			str += "Unknown "
		}
	}
	return str
}

// capture all explicit return types
func captureExplicitReturns(p *compile.Parser) []tok.Token {
	explicitReturns := []tok.Token{}
	for ;; {
		prevTok := p.Token
		if p.Token == tok.Eof {
			break
		} else if prevTok == tok.Return {
			p.Next()
			if p.Token == tok.Identifier {
				continue
			}
			fmt.Println("return -> ", p.Token)
			explicitReturns = append(explicitReturns, p.Token)
		}
		p.Next()
	}
	return explicitReturns
}

// assemble return types into a list of custom types for the LSP
func assembleReturnTypes(explicitReturns []tok.Token) []type_t {
	hasTrue := false
	hasFalse := false
	returnTypes := []type_t{}
	for _, tok := range explicitReturns {
		switch tok.String() {
		case "Number":
			returnTypes = append(returnTypes, Number_t)
		case "String":
			returnTypes = append(returnTypes, String_t)
		case "True":
			returnTypes = append(returnTypes, True_t)
			hasTrue = true
		case "False":
			returnTypes = append(returnTypes, False_t)
			hasFalse = true
		default:
			// TODO: later change this to unknown when more concrete inference is implemented
			returnTypes = append(returnTypes, Any_t)
		}
	}
	// if there are both true and false return types, then the function can return a bool
	if hasTrue && hasFalse {
		returnTypes = append(returnTypes, Bool_t)
		// remove true and false from the list
		for _, tok := range returnTypes {
			if tok == True_t || tok == False_t {
				returnTypes = removeTrueAndFalse(returnTypes)
			}
		}
	}
	return returnTypes
}

func removeTrueAndFalse(returnTypes []type_t) []type_t {
	newReturnTypes := []type_t{}
	for _, tok := range returnTypes {
		if tok != True_t && tok != False_t {
			newReturnTypes = append(newReturnTypes, tok)
		}
	}
	return newReturnTypes
}

func main() {
	code := `
		function() 
			{ 
			a = 2
			if true 
				{ return 1 } 
			else if  
				{ return '2' } 
			else if 
				{ return a } 
			else 
				{ return true } 
			return false 
			}`

	p := compile.AstParser(code)
	explicitReturns := captureExplicitReturns(p)
	firstPass := assembleReturnTypes(explicitReturns)
	fmt.Println(TypeTString(firstPass))
}
