package main

import (
	"fmt"
	"os"

	"encoding/json"

	"github.com/apmckinlay/gsuneido/compile"
)

func main() {
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

	// catching the simplest type error: `type number is not callable`
	/*
		1. Mark x as unknown (as it won't be known in the first pass)
		2. Mark num as unknown + Number (123)
		3. Evaluate x to be Number (as only then could it be added to 123)
		4. Evaluate num to be Number
		5. Throw error as Number is not callable
	*/

	src = `
			function(x, y, z)
				{
				num = x + 123
				num++
				if String?(x) and Number?(y) 
					{
					abc = x + y + z + num
					} 
				else 
					{
					num()
					}
				.qux()
				}
			`

	src = `class {
			x: 0
			msg: "hello"
		Hello(x, y) { return x + y }
		pvt_foo() { return .x }
		originalTestFunc(x, y, z)
			{
			num = x + 123
			num++
			if String?(x) and Number?(y) 
				{
				abc = x + y + z + num
				} 
			else 
				{
				num()
				}
			.qux()
			}
		pvt_bar() { return .msg }
		SetX(x) { .x = x }
		SetMsg(msg) { .msg = msg }
		Get() { return Object(numx: .x, strmsg: .msg) }
		AddBreak() { return x + "123" }
		}`
	/*
	 */

	fmt.Println("src:", src)
	fmt.Println()
	fmt.Println("compiled:", compile.AstParser(src).Const())
	fmt.Println()

	p := compile.AstParser(src)

	cl := p.TypeClass()
	fmt.Println("=== Class ===")
	fmt.Println("class ", cl.Name, " from ", cl.Base)
	fmt.Println("\tAttributes:")
	for name, attr := range cl.Attributes {
		a := attr[0]
		fmt.Println("\t", name, ":")
		fmt.Println("\t\t", a.Value)
		fmt.Println("\t\t", a.Tag)
		fmt.Println("\t\t", a.Type_t)
		fmt.Println()
	}
	fmt.Println("\tMethods:")
	for _, method := range cl.Methods {
		m := method[0]
		fmt.Println("\t", m.Name, "(", m.Parameters, ")")
		for _, stmt := range m.Body {
			fmt.Println("\t\t", stmt)
		}
		fmt.Println()
	}

	// convert to json
	jsonData, err := json.Marshal(cl)
	if err != nil {
		panic(err)
	}

	// delete file if it exists
	if _, err := os.Stat("output.json"); err == nil {
		err = os.Remove("output.json")
		if err != nil {
			panic(err)
		}
	}
	// write json data to file
	fobj, err := os.OpenFile("output.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer fobj.Close()

	_, err = fobj.WriteString(string(jsonData))
	if err != nil {
		panic(err)
	}

}
