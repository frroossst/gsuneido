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
			return x + 1
		}
	`
	fmt.Println("src:", src)
	// remove whitespace

	fmt.Println("compiled:", compile.AstParser(src).Const())

	p := compile.NewParser(src)
	f := p.Function()

	fmt.Println(p.Lxr.GetSI())

	ast.PropFold(f)

	// fmt.Println("type:", f.Type())
	// fmt.Println("ast:", f.String())

	// tnode := ast.AsTypedAST(f)
	// tnode.SetType("function")
	// fmt.Println(tnode.GetType())
	// fmt.Println(tnode.String())

	fmt.Println("item: ", p.Item, "token: ", p.Token, p.Lxr.GetSI())

	// dfs(f, visitor)
	t := dfs(f, typeVisitor)
	fmt.Println("typed:", t.String())
}

// tree to represent a copy of the AST,
// these nodes are analogous to the AST nodes
// but they do not need to impl the Node interface
// as they are not used in the compiler
type TypedTree struct {
	root		*TypedNode
}

func (t* TypedTree) SetRoot(r TypedNode) {
	t.root = &r
}

func (t* TypedTree) GetRoot() TypedNode {
	return *t.root
}

type TypedNode struct {
	uid 		uint64
	node_t 		string	
	children 	[]*TypedNode
}

func TypedNodeConstructor() *TypedNode {
	return &TypedNode{
		uid: uid.Next(),
		node_t: "undetermined",
		children: []*TypedNode{},
	}
}

func (n *TypedNode) AddChild(c *TypedNode) {
	n.children = append(n.children, c)
}

func (n *TypedNode) GetChilden() []*TypedNode {
	return n.children
}

// recursively descend using depth first search and continually build a copy with TypedNodes
func copyAST(node ast.Node, visitorFn func(ast.Node) ast.Node, newCopy ast.TypedNode) ast.TypedNode {

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
