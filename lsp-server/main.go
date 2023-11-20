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

	src := `
		function(x) {
			// test folding
			notLogic = not true
			constF = 1 + 2 + 3 + 4 + 5
			return x + 1
		}
	`
	fmt.Println("src:", src)
	fmt.Println("compiled:", compile.AstParser(src).Const())

	p := compile.NewParser(src)
	fmt.Println("going to parse")
	f := p.Function()

	ast.PropFold(f)

	fmt.Println("folded:", f)

	// t := dfs(f, typeVisitor)
	// fmt.Println("typed ast:", t)
}

// tree to represent a copy of the AST,
// these nodes are analogous to the AST nodes
// but they do not need to impl the Node interface
// as they are not used in the compiler
type TypedTree struct {
	root *TypedNode
}

func (t *TypedTree) SetRoot(r TypedNode) {
	t.root = &r
}

func (t *TypedTree) GetRoot() TypedNode {
	return *t.root
}

type TypedNode struct {
	uid      uint64
	node_t   string
	data     string
	meta     string
	children []*TypedNode
}

func TypedNodeConstructor() *TypedNode {
	return &TypedNode{
		uid:      0,
		node_t:   "undetermined",
		children: []*TypedNode{},
	}
}

func (n *TypedNode) GetUID() uint64 {
	return n.uid
}

func (n *TypedNode) SetUID(u uint64) {
	n.uid = u
}

func (n *TypedNode) GetType() string {
	return n.node_t
}

func (n *TypedNode) AddChild(c *TypedNode) {
	n.children = append(n.children, c)
}

func (n *TypedNode) GetChilden() []*TypedNode {
	return n.children
}


// dfs is a depth-first search of the AST
// it traverses the AST and applies the visitor function
// to each node
func dfs(node ast.Node, visitorFn func(ast.Node) ast.TypedNode) ast.TypedNode {
	// apply the visitor function to the current node
	currNode := visitorFn(node)

	currNode.Children(func(child ast.Node) ast.Node {
		dfs(child, visitorFn)
		return child
	})

	return ast.AsTypedNode(currNode)
}

// define a recursive function to be used in dfs with the
// function signature fn func(node Node) Node
func visitor(node ast.Node) ast.Node {
	return node
}

func typeVisitor(node ast.Node) ast.TypedNode {
	tnode := ast.AsTypedNode(node)
	tnode.SetUID(uid.Next())
	return tnode
}

func serialise(root ast.TypedNode) string {
	return ""
}
