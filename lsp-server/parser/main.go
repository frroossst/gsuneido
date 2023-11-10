package main

import (
    "fmt"
)

// Parser for custom type definitions
// 1. Direct type definitions
// type <name> >>= <typedef>
// 2. Union type definitions
// type <name> = <typedef> | <typedef> | ...
// 3. Object type definitions
// type <name> = #( key1: <typedef>, key2: <typedef>, ... )
// 4. Function type definitions
// <typedef> foo(<typedef>, <typedef>, a: <typedef>, b: <typedef>, ...)
// 5. Type alias
// type <name> = <typedef>
// 6. Import type definitions
// import <type_name> from <type_file>

func main() {

    src := "type foo = #( a: int, b: int )"

}

func lexer(src string) {

}

func parser(src string) {

}

func typeChecker(src string) {
    
}

func codeGen(src string) {

}

func test() {



}

