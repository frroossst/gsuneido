package main

import (
	"fmt"

	"sync"
	"time"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
)

type UID struct {
	counter uint64
	mu      sync.Mutex
}

func (u *UID) Next() uint64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.counter++

	id := uint64(time.Now().UnixMicro())

	return id
}

// define globally acessibly UID
var uid UID

func main() {
	uid = UID{}
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
	fmt.Println("src:", src)
	fmt.Println("compiled:", compile.AstParser(src).Const())

	p := compile.NewParser(src)
	f := p.Function()

	ast.PropFold(f)

	t := dfs(f, propFoldVisitor)
	fmt.Println("typed ast:", t)

	// typing AST
	typeAST_norep(f)
	fmt.Println("=== type maps ===")
	for k, v := range typeInfoMap {
		fmt.Println("key:", (*k).String(), "value:", v)
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
}

func (t *TypeInfo) String() string {
	return fmt.Sprintf("Tag= %s, Node_t= %s\n", t.Tag, t.Node_t)
}

// ! Global, remove later
// Create a map to store the type information for each node
var typeInfoMap = make(map[*ast.Node]*TypeInfo)

// Function to set the type information for a node
func SetTypeInfo(node *ast.Node, tag string, node_t string) {
	typeInfoMap[node] = &TypeInfo{Tag: tag, Node_t: node_t}
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

func typeAST_norep(node ast.Node) {
	// Create a set to keep track of visited nodes
	visited := make(map[ast.Node]bool)

	var dfsInner func(node ast.Node)
	dfsInner = func(node ast.Node) {
		if visited[node] {
			// Skip this node if it has already been visited
			return
		}

		// Mark this node as visited
		visited[node] = true

		// Apply the visitor function to the current node
		typeVisitor(node)

		// Traverse the children
		node.Children(func(child ast.Node) ast.Node {
			dfsInner(child)
			return child
		})
	}

	// Start the DFS traversal
	dfsInner(node)
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
	switch node.(type) {
	case *ast.Binary:
		tag = "binary"
		node_t = "expr"
	case *ast.Nary:
		tag = "nary"
		node_t = "expr"
	case *ast.Unary:
		tag = "unary"
		node_t = "expr"
	case *ast.Call:
		tag = "call"
		node_t = "expr"
	case *ast.Function:
		tag = "function"
		node_t = "stmt"
	default:
		tag = "unknown"
		node_t = "unknown"
	}

	SetTypeInfo(&node, tag, node_t)
}

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

// dfs is a depth-first search of the AST
// it traverses the AST and applies the visitor function
// to each node
func dfs(node ast.Node, visitorFn func(ast.Node) ast.Node) ast.Node {
	// apply the visitor function to the current node
	currNode := visitorFn(node)

	// currNode.Children(func(child ast.Node) ast.Node {
	// 	dfs(child, visitorFn)
	// 	return child
	// })

	return currNode
}

func propFoldVisitor(node ast.Node) ast.Node {
	if stmt, ok := node.(ast.Statement); ok {
		stmt.Position() // for error reporting
	}
	propFoldChildren(node) // RECURSE

	return node
}

func propFoldChildren(node ast.Node) {
	switch n := node.(type) {
	case *ast.Binary:
		fmt.Println("binary:", n.Tok)
		return
	case *ast.Nary:
		fmt.Println("nary:", n.Tok)
		return
	case *ast.Return:
		return
	default:
		fmt.Println("default:", n)
	}

	node.Children(propFoldVisitor) // RECURSE
}
