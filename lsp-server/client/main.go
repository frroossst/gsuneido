package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "strings"
    "net/http"

	"github.com/apmckinlay/gsuneido/compile"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
//	"github.com/apmckinlay/gsuneido/compile/lexer"
)

type TypeError int
const (
    TypeMismatch TypeError = iota
    TypeNotKnown
    TypeNotInferred
    TypeNotInferredFromChildren
    TypeNotInferredFromParent
    TypeNotInferredFromSibling
)

type TypeErrorThrown struct {
    ErrorType TypeError
    Message string
}

type SuneidoType int
const (
    Unknown SuneidoType = iota
    Any                       
    Boolean                   
    Number                    
    String                    
    Object                    
    Function                  
    None                      
    Operator
)

func (s SuneidoType) String() string {
	return [...]string{
        "Unknown", "Any", "Boolean", "Number", "String", 
        "Object", "Function", "None", "Operator"}[s]
}

type EqualTypeOperators int
const (
    Eq EqualTypeOperators = iota
    Cat
)
type OperatorType struct {
    Operator EqualTypeOperators
    Left     SuneidoType
    Right    SuneidoType
}

type Node struct {
    Value    string
    Children []Node
    Type     SuneidoType
}

func generateAST(input string) Node {
	stack := []Node{}
	word := ""

	for _, c := range input {
		switch c {
		case '(':
			if word != "" {
				stack = append(stack, Node{Value: word})
				word = ""
			}
			stack = append(stack, Node{Value: "("})
		case ')':
			if word != "" {
				stack = append(stack, Node{Value: word})
				word = ""
			}
			temp := []Node{}
			for len(stack) > 0 && stack[len(stack)-1].Value != "(" {
				temp = append(temp, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1] // pop "("
			stack[len(stack)-1].Children = temp
		case ' ':
			if word != "" {
				stack = append(stack, Node{Value: word})
				word = ""
			}
		default:
			word += string(c)
            word = strings.TrimSpace(word)
		}
	}

	return stack[0]
}

/*
func reduceTree(node Node) Node {
    newNode := Node{
        Value: node.Value,
        Type:  node.Type,
    }

    for _, child := range node.Children {
        newNode.Children = append(newNode.Children, reduceTree(child))
    }

    if newNode.Value == "Binary" || newNode.Value == "Unary" || newNode.Value == "Trinary" {
        if len(newNode.Children) > 0 {
            newNode.Value = newNode.Children[len(newNode.Children)-1].Value
            newNode.Type = Operator
        }
    }

    return newNode
}
*/

func reduceTree(node Node) Node {
    newNode := Node{
        Value: node.Value,
        Type:  node.Type,
    }

    for _, child := range node.Children {
        newNode.Children = append(newNode.Children, reduceTree(child))
    }

    tempType := Operator
    if newNode.Value == "Call" {
        tempType = Function
    }

    if newNode.Value == "Binary" || newNode.Value == "Unary" || 
        newNode.Value == "Trinary" || newNode.Value == "Call" {
        if len(newNode.Children) > 0 {
            newNode.Value = newNode.Children[len(newNode.Children)-1].Value
            newNode.Type = tempType
            newNode.Children = newNode.Children[:len(newNode.Children)-1]
        }
    }

    return newNode
}

// first pass to determine obvious syntactically
// guaranteed types
func directTypeInference(node Node, parent Node) Node {

    return node

}


func main() {
    // Make an HTTP POST request to the server
    inputStr := "function() { x = 1 }"
    // inputStr = "function(a, b) { return a is b  }"
    // inputStr = "function() { return 1 + 2 + 3 + 4 + 5 + 6 }"
    inputStr = 
    `function(x) 
        {
        foo = "123"
        bar = x $ foo
        baz = foo is bar
        return baz
        }
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
    // expr := p.Expression()
    // lxr := p.Lxr
    // for ;; {
    //     item := lxr.Next()
    //     if item.Token == tok.Eof {
    //         break
    //     }
    //     if !skipPrintMap[item.Token] {
    //         fmt.Println(item)
    //     }
    // }

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

    input := RequestJSON{Input: inputStr}
    responseJSON, err := sendRequest(input, "http://localhost:8080/process")
    if err != nil {
        fmt.Println("Error sending request:", err)
        return
    }

    response := responseJSON.Output
    fmt.Println("response: ", response)
    parsedAst := generateAST(response)
    newRoot := reduceTree(parsedAst)
    fmt.Println()
    printNode(newRoot, "\t")
}

func printNode(node Node, indent string) error {
	fmt.Println(indent + "Node: " + node.Value)
	fmt.Println(indent + "  Type: " + node.Type.String())
    if len(node.Children) == 0 {
        fmt.Println(indent + "  Children: []")
        return nil
    }
	fmt.Println(indent + "  Children: ")
	for _, child := range node.Children {
		printNode(child, indent+"    ")
	}
    return nil
}

func sendRequest(input RequestJSON, url string) (ResponseJSON, error) {
    requestBody, err := json.Marshal(input)
    if err != nil {
        return ResponseJSON{}, err
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
    if err != nil {
        return ResponseJSON{}, err
    }
    defer resp.Body.Close()

    var response ResponseJSON
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return ResponseJSON{}, err
    }

    return response, nil
}

type RequestJSON struct {
    Input string `json:"input"`
}

type ResponseJSON struct {
    Output string `json:"output"`
}

