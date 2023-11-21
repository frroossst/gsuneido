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

type TypedNode struct {
	node ast.Node
	// "expr" or "stmt"
	tag    string
	node_t string
}

func NewTypedNode(node ast.Node, tag string) TypedNode {
	return TypedNode{node, tag, ""}
}

func (t TypedNode) String() string {
	return fmt.Sprintf("%s %s", t.tag, t.node.String())
}

func (t *TypedNode) GetType() string {
	return t.node_t
}

func (t *TypedNode) SetType(node_t string) {
	t.node_t = node_t
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
		fmt.Println("lhs:", n.Lhs.String())
		fmt.Println("rhs:", n.Rhs.String())
		return
	case *ast.Nary:
		fmt.Println("nary:", n.Tok)
		fmt.Println("exprs:", n.Exprs)
		t := NewTypedNode(n, "expr")
		fmt.Println("typed node:", t)
		return
	case *ast.Return:
		fmt.Println("return:", n.ValueBase.Type())
		return
	default:
		fmt.Println("default:", n)
	}

	node.Children(propFoldVisitor)
}
