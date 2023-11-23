package main

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
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
			num = x + 123
			num()
			}`

	fmt.Println("src:", src)
	fmt.Println("compiled:", compile.AstParser(src).Const())

	p := compile.NewParser(src)
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

	fmt.Println("=== KV Set ===")
	for _, k := range visitedSet.Keys {
		// decode string from Base64
		b, err := base64.StdEncoding.DecodeString(k)
		if err != nil {
			panic(err)
		}
		fmt.Println("key:", string(b))
	}

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

type TypeInfo struct {
	Tag    string
	Node_t string
	Token  string
}

func (t *TypeInfo) String() string {
	return fmt.Sprintf("Tag= %s, Node_t= %s, Token=%s\n", t.Tag, t.Node_t, t.Token)
}

// TODO: Instead of using a global map, use a map that is passed around
// var typeInfoMap = make(map[*ast.Node]*TypeInfo)
var typeInfoSet = KeyValueSet[*TypeInfo]{}

// var globalVisited = make(map[string]bool)
var visitedSet = KeyValueSet[bool]{}

type KeyValueSet[T any] struct {
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

/*
func SetTypeInfo(node *ast.Node, tag string, node_t string, tok string) {
	typeInfoMap[node] = &TypeInfo{Tag: tag, Node_t: node_t, Token: tok}
}

func GetTypeInfo(node *ast.Node) *TypeInfo {
	return typeInfoMap[node]
}
*/

// ==================================================================

func markVisited(str string) {
	// encode string to Base64
	b64 := base64.StdEncoding.EncodeToString([]byte(str))
	visitedSet.Set(b64, true)
}

func dfsInner(node ast.Node) {
	b64 := base64.StdEncoding.EncodeToString([]byte(node.String()))
	if visitedSet.Contains(b64) {
		return
	}

	str := node.String()
	defer markVisited(str)
	// markVisited(str)

	// Mark this node as globalVisited
	// globalVisited[node.String()] = true

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

func typeVisitor(node ast.Node) {
	var tag, node_t string
	tok := "unimpl!"

	// why doesn't node.(type) match with ast.Binart?
	// print typeof using reflect
	fmt.Println("node type:", reflect.TypeOf(node))
	switch n := node.(type) {
	case *ast.Unary:
		tag = "expr"
		node_t = "unknown"
	case *ast.Binary:
		fmt.Println("hit binary")
		tag = "expr"
		node_t = "unknown"
	case *ast.Nary:
		tag = "expr"
		node_t = "unknown"
	case *ast.Compound:
		tag = "expr"
		node_t = "unknown"
	case *ast.ExprStmt:
		tag = "stmt"
		node_t = "unknown"
		node.Children(func(child ast.Node) ast.Node {
			dfsInner(child)
			return child
		})
	case *ast.Return:
		tag = "stmt"
		node_t = "unknown"
	case *ast.Ident:
		tag = "expr"
		node_t = "unknown"
	case *ast.Call:
		tag = "expr"
		node_t = "unknown"
	case *ast.Function:
		tag = "expr"
		node_t = "unknown"
	default:
		fmt.Println("n=", n)
		tag = "unknown"
		node_t = "unknown"
	}
	// dont't really like this side effect
	// SetTypeInfo(&node, tag, node_t, tok)
	// convert the pointer &node to a string
	b64 := base64.StdEncoding.EncodeToString([]byte(node.String()))
	typeInfoSet.Set(b64, &TypeInfo{Tag: tag, Node_t: node_t, Token: tok})
}
