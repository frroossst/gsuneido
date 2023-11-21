package main

import (
	"bytes"
	"fmt"

	"sync"
	"time"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
	. "github.com/apmckinlay/gsuneido/core"
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

	// copying AST
	c := copyAST(f)
	fmt.Println("copied ast:", c)
	// get runtime reflect type of c
	fmt.Println("type:", c.String())
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

func (t *TypedNode) Children(f func(ast.Node) ast.Node) {
	t.node.Children(f)
}

func (t *TypedNode) Get(*Thread, Value) Value {
	panic("implement me")
}

func (t *TypedNode) SetPos(org, end int32) {
	t.node.SetPos(org, end)
}

func (t *TypedNode) astNode() { t.node.astNode() }

func (t *TypedNode) String() string {
	var buffer bytes.Buffer

	var dfs func(node *TypedNode)
	dfs = func(node *TypedNode) {
		// Write the current node type, type info, and tag to the buffer
		buffer.WriteString(fmt.Sprintf("(%s %s %T ", node.node_t, node.tag, node.node))

		// Traverse the children
		node.node.Children(func(child ast.Node) ast.Node {
			if typedChild, ok := child.(*TypedNode); ok {
				dfs(typedChild)
			}
			return child
		})

		buffer.WriteString(")")
	}

	// Start the DFS traversal
	dfs(t)

	return buffer.String()
}

// =============================================================================

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
