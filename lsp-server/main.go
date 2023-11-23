package main

import (
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
	for k, v := range typeInfoMap {
		fmt.Println("pointer: ", k, "key:", (*k).String(), "value:", v)
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

// =============================================================================

type TypedNode struct {
	node ast.Node
	// "expr" or "stmt"
	tag    string
	node_t string
}

func NewTypedNode(node ast.Node, tag string) *TypedNode {
	return &TypedNode{node, tag, ""}
}

func (t *TypedNode) GetType() string {
	return t.node_t
}

func (t *TypedNode) SetType(node_t string) {
	t.node_t = node_t
}

func (t *TypedNode) GetNode() ast.Node {
	return t.node
}

// =============================================================================

type TypeStoreDB struct {
	db map[ast.Node]string
}

type TypeInfo struct {
	Tag    string
	Node_t string
	Token  string
}

func (t *TypeInfo) String() string {
	return fmt.Sprintf("Tag= %s, Node_t= %s, Token=%s\n", t.Tag, t.Node_t, t.Token)
}

// ! Global, remove later
// Create a map to store the type information for each node
var typeInfoMap = make(map[*ast.Node]*TypeInfo)
var globalVisited = make(map[string]bool)

// Function to set the type information for a node
func SetTypeInfo(node *ast.Node, tag string, node_t string, tok string) {
	typeInfoMap[node] = &TypeInfo{Tag: tag, Node_t: node_t, Token: tok}
}

// Function to get the type information for a node
func GetTypeInfo(node *ast.Node) *TypeInfo {
	return typeInfoMap[node]
}

func NewTypeStoreDB() *TypeStoreDB {
	return &TypeStoreDB{make(map[ast.Node]string)}
}

func (t *TypeStoreDB) Get(node ast.Node) string {
	return t.db[node]
}

func (t *TypeStoreDB) Set(node ast.Node, node_t string) {
	t.db[node] = node_t
}

func dfsInner(node ast.Node, globalVisited map[string]bool) {
	if globalVisited[node.String()] {
		// Skip this node if it has already been globalVisited
		return
	}

	// Mark this node as globalVisited
	globalVisited[node.String()] = true

	// Apply the visitor function to the current node
	typeVisitor(node)

	// Traverse the children
	node.Children(func(child ast.Node) ast.Node {
		dfsInner(child, globalVisited)
		return child
	})
}

func typeAST_norep(node ast.Node) {
	// Create a set to keep track of globalVisited nodes
	// globalVisited := make(map[ast.Node]bool)

	// Start the DFS traversal
	dfsInner(node, globalVisited)
}

func typeAST(node ast.Node) {
	typeVisitor(node)

	node.Children(func(child ast.Node) ast.Node {
		typeAST(child)
		return child
	})
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
		globalVisited := make(map[string]bool)
		node.Children(func(child ast.Node) ast.Node {
			dfsInner(child, globalVisited)
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
	// return (tag), (node_t)

	SetTypeInfo(&node, tag, node_t, tok)
}

/*
func copyAST(node ast.Node) *TypedNode {
	currNode := copyVisitor(node)
	return currNode
}

func copyVisitor(node ast.Node) *TypedNode {
	var tag string
	switch node.(type) {
	case *ast.Binary:
		tag = "binary"
	case *ast.Nary:
		tag = "nary"
	case *ast.Unary:
		tag = "unary"
	case *ast.Call:
		tag = "call"
	case *ast.Function:
		tag = "function"
	default:
		tag = "unknown"
	}
	if stmt, ok := node.(ast.Statement); ok {
		stmt.Position()
	}
	copyChildren(node) // RECURSE

	return NewTypedNode(node, tag)
}

func copyChildren(node ast.Node) {
	node.Children(func(child ast.Node) ast.Node {
		typedChild := NewTypedNode(child, "expr")
		child = typedChild.GetNode()
		copyVisitor(child)
		return child
	})
}
*/
