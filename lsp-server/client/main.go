package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

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
)

func (s SuneidoType) String() string {
	return [...]string{
        "Unknown", "Any", "Boolean", "Number", "String", 
        "Object", "Function", "None"}[s]
}

type Node struct {
    Value    string
    Children []Node
    Type     SuneidoType
}

func Parse(input string) Node {
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
		}
	}

	return stack[0]
}

func Parse2(input string) Node {
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
		}
	}

	return stack[0]
}

func printNode(node Node, indent string) {
	fmt.Println(indent + "Node: " + node.Value)
	fmt.Println(indent + "  Type: " + node.Type.String())
	fmt.Println(indent + "  Children: ")
	for _, child := range node.Children {
		printNode(child, indent+"    ")
	}
}

func main() {
    // Make an HTTP POST request to the server
    inputStr := "function() { x = 1 }"
    inputStr = "function() { foo = 123; if (foo) { return 'hello' } else { return true } }"
    input := RequestJSON{Input: inputStr}
    responseJSON, err := sendRequest(input, "http://localhost:8080/process")
    if err != nil {
        fmt.Println("Error sending request:", err)
        return
    }

    response := responseJSON.Output
    fmt.Println("response: ", response)
    parsedAst := Parse2(response)
    printNode(parsedAst, "")
    // fmt.Println("parsed: ", printNode(Parse2(response), ""))
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
