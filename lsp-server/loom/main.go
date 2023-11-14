package main

import (
	"fmt"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
)

func main() {
	src := `
		function(x) {
			return x + 1
		}
	`

	p := compile.NewParser(src)
	f := p.Function()
	ast.Blocks(f)

	fmt.Println(f.Type())

	fmt.Println(f.String())
}
