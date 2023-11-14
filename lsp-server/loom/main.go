package main

import (
	"fmt"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
)

	/*
	p := compile.NewParser(src)
	f := p.Function()
	ast.Blocks(f)

	fmt.Println(f.Type())
	fmt.Println(f.String())

	for _, stmt := range f.Body {
		if stmt != nil {
			fmt.Println(stmt.String())
		}
	}
	*/

func main() {
	src := `
		function(x) {
			return x + 1
		}
	`
	fmt.Println("src:", src)

	p := compile.NewParser(src)
	f := p.Function()

	ast.PropFold(f)

	fmt.Println("type:", f.Type())

	fmt.Println("ast:", f.String())

	ast.DepthFirstSearch(f)
}
