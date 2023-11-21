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
			if not notLogic or notLogic {
				return x * 2
			}
			return x + 1 + b
		}
	`
	fmt.Println("src:", src)
	fmt.Println("compiled:", compile.AstParser(src).Const())

	p := compile.NewParser(src)
	fmt.Println("going to parse")
	f := p.Function()

	ast.PropFold(f)

	fmt.Println("folded:", f)

	fmt.Println("export = ", f.String())

	t := dfs(f, propFoldVisitor)
	fmt.Println("typed ast:", t)
}

// dfs is a depth-first search of the AST
// it traverses the AST and applies the visitor function
// to each node
func dfs(node ast.Node, visitorFn func(ast.Node) ast.Node) ast.Node {
	// apply the visitor function to the current node
	currNode := visitorFn(node)

	currNode.Children(func(child ast.Node) ast.Node {
		dfs(child, visitorFn)
		return child
	})

	return currNode
}

// define a recursive function to be used in dfs with the
// function signature fn func(node Node) Node
func visitor(node ast.Node) ast.Node {
	return node
}

func propFoldVisitor(node ast.Node) ast.Node {
	if stmt, ok := node.(ast.Statement); ok {
		stmt.Position() // for error reporting
	}
	// fmt.Println("node:", node.String())
	propFoldChildren(node) // RECURSE

	return node
}

func propFoldChildren(node ast.Node) {
	switch n := node.(type) {
	case *ast.Binary:
		fmt.Println("binary:", n.Tok)
	default:
		fmt.Println("default:", n)
	}

	node.Children(propFoldVisitor)
}

func typeVisitor(node ast.Node) ast.TypedNode {
	return nil
}

func serialise(root ast.TypedNode) string {
	return ""
}
