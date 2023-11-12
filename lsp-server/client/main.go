package main

import (
    "fmt"

	"github.com/apmckinlay/gsuneido/compile"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
//	"github.com/apmckinlay/gsuneido/compile/lexer"
)

func main() {
    // Make an HTTP POST request to the server
    inputStr := "function() { x = 1 }"
    // inputStr = "function(a, b) { return a is b  }"
    // inputStr = "function() { return 1 + 2 + 3 + 4 + 5 + 6 }"
    inputStr = 
    `foo = function(x) 
        {
        foo = "123"
        bar = x $ foo
        baz = foo is bar
        return baz
        }
    a = foo()
    Print(:a)
    `
    // inputStr = "function() { x = 'foo'\n if String?(foo) { return 'str' } }"
    // inputStr = "function() { x = Object(a: 1, b: 's') }"
    // inputStr = "function() { x++ }"
    // inputStr = "function() { foo = 123; if (foo) { return 'hello' } else { return true } }"
    
    // ! testing official compiler
    // define an arrray of tokens to skip
    skipPrint := []tok.Token{tok.Whitespace, tok.Newline, tok.Comment, tok.LCurly, tok.RCurly, tok.LParen, tok.RParen}
    skipPrintMap := make(map[tok.Token]bool)
    for _, token := range skipPrint {
        skipPrintMap[token] = true
    }

	p := compile.AstParser(inputStr)
    p.InitFuncInfo()
    lxr := p.Lxr
    for ;; {
        item := lxr.Next()
        if item.Token == tok.Eof {
            break
        }
        if !skipPrintMap[item.Token] {
            fmt.Println(item)
        }
    }

    //  var mainThread Thread
    //	mainThread.Name = "main"
    //	mainThread.UIThread = true
    //	v, results := compile.Checked(&mainThread, inputStr)
    //    fmt.Println("v: ", v)
    //	for _, s := range results {
    //		fmt.Println("(" + s + ")")
    //	}
	// p := compile.AstParser(inputStr)
    if true {
        panic("lsp-server/client/main.go: 211 [DEBUG]")
    }

}
