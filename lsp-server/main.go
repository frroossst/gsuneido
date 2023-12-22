package main

import (
	"fmt"
	//	"os"
	"reflect"
	"regexp"

	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
	// . "github.com/apmckinlay/gsuneido/core"
)

func main() {
	/*

	 */

	src := `
		function(x) 
			{
			// test folding
			notLogic = not true
			constF = 1 + 2 + 3 + 4 + 5
			if not notLogic or notLogic {
				return x * 2
			}
			bar = -5
			if true { bar = 1 } else { bar = 2 }
			bar()
			return x + 1 + b
			}
	`

	// catching the simplest type error: `type number is not callable`
	/*
		1. Mark x as unknown (as it won't be known in the first pass)
		2. Mark num as unknown + Number (123)
		3. Evaluate x to be Number (as only then could it be added to 123)
		4. Evaluate num to be Number
		5. Throw error as Number is not callable
	*/

	src = `
			function(x)
				{
				num = x + "123"
				num++
				num()
				}
			`

	/*
		class {
			x: 0
			msg: "hello"
		}
		pvt_foo() { return .x }
		pvt_bar() { return .msg }
		SetX(x) { .x = x }
		SetMsg(msg) { .msg = msg }
		Get() { return Object(numx: .x, strmsg: .msg) }
		AddBreak() { return x + "123" }

	*/

	fmt.Println("src:", src)
	fmt.Println("compiled:", compile.AstParser(src).Const())
	p := compile.AstParser(src)
	fmt.Println(p.TypeConst())

	/*
		fmt.Println("=== AST ===")
		p := compile.AstParser(src)
		ast := p.Const()
		fmt.Println(ast)
		fmt.Println(reflect.TypeOf(ast))
		fmt.Println("=== Children ===")
		children := ast.Get(nil, SuStr("children"))
		firstChild := children.Get(nil, IntVal(10))
		firstChild_t := firstChild.Get(nil, SuStr("type"))
		fmt.Println(firstChild_t)
		fmt.Println("==============")
	*/

	/*
		p := compile.NewParser(src)
		fmt.Println(p.String())
			f := p.Function()

			ast.PropFold(f)

			// typing AST
			typeAST_norep(f)
			fmt.Println("=== type maps ===")
			for _, k := range typeInfoSet.Keys {
				b, err := base64.StdEncoding.DecodeString(k)
				if err != nil {
					panic(err)
				}
				if v, ok := typeInfoSet.Get(k); ok {
					fmt.Println("key:", string(b), "value:", v)
				}
			}
	*/

	// fmt.Println("=== KV Set ===")
	// for _, k := range visitedSet.Keys {
	// 	// decode string from Base64
	// 	b, err := base64.StdEncoding.DecodeString(k)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("key:", string(b))
	// }

	// write to file (for debugging)
	/*
		jsonData := typeInfoSet.String()

		fmt.Println("=== JSON ===")
		fmt.Println("jsonify: ", jsonData)

		// delete file if it exists
		if _, err := os.Stat("output.json"); err == nil {
			err = os.Remove("output.json")
			if err != nil {
				panic(err)
			}
		}

		// write to file
		fobj, err := os.OpenFile("output.json", os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			panic(err)
		}
		defer fobj.Close()

		_, err = fobj.WriteString(jsonData)
		if err != nil {
			panic(err)
		}
	*/

}

type SuType int

const (
	Number SuType = iota
	String
	Function
	Boolean
	Unknown
	Undetermined
	// acts like the None type in python
	Unit
	Object
)

// ==================================================================

type String_t struct {
	Value string
}

func (s *String_t) String() string {
	return s.Value
}

var typeInfoSet = KeyValueSet[*String_t]{}

type Boolean_t struct {
	Value bool
}

func (b *Boolean_t) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

var visitedSet = KeyValueSet[*Boolean_t]{}

type Stringer interface {
	String() string
}

type KeyValueSet[T Stringer] struct {
	Keys   []string
	Values []T
}

func (kvs *KeyValueSet[T]) Get(key string) (T, bool) {
	// find index of key
	for i, k := range kvs.Keys {
		if k == key {
			return kvs.Values[i], true
		}
	}
	var zero T
	return zero, false
}

func (kvs *KeyValueSet[T]) Set(key string, value T) {
	for i, k := range kvs.Keys {
		if k == key {
			kvs.Values[i] = value
			return
		}
	}
	kvs.Keys = append(kvs.Keys, key)
	kvs.Values = append(kvs.Values, value)
}

func (kvs *KeyValueSet[T]) Contains(key string) bool {
	for _, k := range kvs.Keys {
		if k == key {
			return true
		}
	}
	return false
}

func (kvs *KeyValueSet[T]) String() string {
	// serialise key and value to JSON like string
	var buffer strings.Builder
	buffer.WriteString("{")
	for i, k := range kvs.Keys {
		// decode k from Base64
		b, err := base64.StdEncoding.DecodeString(k)
		if err != nil {
			panic(err)
		}
		if i > 0 {
			buffer.WriteString(", ")
		}
		// Escape special characters in JSON string
		pattern := regexp.MustCompile(`[^\S ]`)

		keyJSON, err := json.Marshal(pattern.ReplaceAllString(string(b), ""))
		if err != nil {
			panic(err)
		}
		valJSON, err := json.Marshal(pattern.ReplaceAllString(kvs.Values[i].String(), ""))
		if err != nil {
			panic(err)
		}

		buffer.WriteString(fmt.Sprintf("%s: %s", keyJSON, valJSON))
	}
	buffer.WriteString("}")
	return buffer.String()
}

// ==================================================================

func markVisited(str string) {
	// encode string to Base64
	b64 := base64.StdEncoding.EncodeToString([]byte(str))
	visitedSet.Set(b64, &Boolean_t{Value: true})
}

func dfsInner(node ast.Node) {
	b64 := base64.StdEncoding.EncodeToString([]byte(node.String()))
	fmt.Println("dfsInner, node.String() ", node.String())
	if visitedSet.Contains(b64) {
		return
	}

	str := node.String()
	defer markVisited(str)
	// markVisited(str)

	// Apply the visitor function to the current node
	typeVisitor(node)

	// Traverse the children
	node.Children(func(child ast.Node) ast.Node {
		dfsInner(child)
		return child
	})
}

func typeAST_norep(node ast.Node) {
	// Create a set to keep track of globalVisited nodes
	// globalVisited := make(map[ast.Node]bool)

	// Start the DFS traversal
	dfsInner(node)
}

func dfsChildrenFn(child ast.Node) ast.Node {
	dfsInner(child)
	return child
}

func typeVisitor(node ast.Node) {
	type_tag := "Unknown"

	fmt.Println("node type:", reflect.TypeOf(node))
	switch n := node.(type) {
	case *ast.Unary:
		type_tag = "Solveable"
	case *ast.Binary:
		type_tag = "Solveable"
		n.Lhs.Children(dfsChildrenFn)
		n.Rhs.Children(dfsChildrenFn)
	case *ast.Nary:
		type_tag = "Solveable"
		n.Children(dfsChildrenFn)
	case *ast.Compound:
		// compound
	case *ast.ExprStmt:
		n.Children(dfsChildrenFn)
	case *ast.Return:
		// return
	case *ast.Ident:
		type_tag = "Variable"
	case *ast.Call:
		// call
	case *ast.Function:
		type_tag = "Function"
	case *ast.Constant:
		type_tag = n.Val.Type().String()
	default:
		fmt.Println("[TODO: default] ", n, reflect.TypeOf(n))
	}

	// dont't really like this side effect
	b64 := base64.StdEncoding.EncodeToString([]byte(node.String()))
	s_t := &String_t{Value: type_tag}
	typeInfoSet.Set(b64, s_t)
}
