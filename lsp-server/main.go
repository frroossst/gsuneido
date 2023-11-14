package main

import (
	"fmt"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
)

type TypedNodeAST struct {
	CurrentNode  ast.TypedNode
	ChildrenNodes []ast.TypedNode
}

func (t TypedNodeAST) String() string {
	return t.CurrentNode.String()
}

func (t *TypedNodeAST) AddChild(child ast.TypedNode) {
	t.ChildrenNodes = append(t.ChildrenNodes, child)
}

func (t TypedNodeAST) GetChildren() []ast.TypedNode {
	return t.ChildrenNodes
}

func main() {
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

	ast.PropFold(f)

	// fmt.Println("type:", f.Type())
	// fmt.Println("ast:", f.String())

	// tnode := ast.AsTypedAST(f)
	// tnode.SetType("function")
	// fmt.Println(tnode.GetType())
	// fmt.Println(tnode.String())

	// dfs(f, visitor)
	t := dfs(f, typeVisitor)
	fmt.Println("typed:", t.String())
}

// dfs is a depth-first search of the AST
// it traverses the AST and applies the visitor function
// to each node
func dfs(node ast.Node, visitor func(ast.Node) ast.TypedNode) ast.TypedNode {
	// apply the visitor function to the current node
	node = visitor(node)

	node.Children(func(child ast.Node) ast.Node {
		dfs(child, visitor)
		return child
	})

	return ast.AsTypedNode(node)
}

// define a recursive function to be used in dfs with the
// function signature fn func(node Node) Node
func visitor(node ast.Node) ast.Node {
	fmt.Println("[visitor]", node.String())
	return node
}

func typeVisitor(node ast.Node) ast.TypedNode {
	tnode :=ast.AsTypedNode(node)
	fmt.Println("[visitor]", node.String(), "_type:", tnode.GetType())
	return tnode
}

	// tnode := typeWrapper(node)
	// fmt.Println("[visitor]", node.String(), "_type:", tnode.GetType())
	// return tnode
func typeWrapper(node ast.Node) ast.TypedNode {
	tnode := ast.AsTypedNode(node)
	return tnode
}

// serialize is a helper function to convert an AST node to a string
// this string is passed onto OCaml for type checking
func serialize(node ast.Node) string {
	return node.String()
}
