package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

type Node struct {
    Value    string
    Children []*Node
}

func main() {
    // Make an HTTP POST request to the server
    input := RequestJSON{Input: "function() { x = 1 }"}
    responseJSON, err := sendRequest(input, "http://localhost:8080/process")
    if err != nil {
        fmt.Println("Error sending request:", err)
        return
    }

    response := responseJSON.Output

    root, _ := parseResponse(response)
    printTree(root, "")
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

func parseResponse(response string) (*Node, string) {
    response = strings.TrimSpace(response)

    if response == "" {
        return nil, response
    }

    openParenIndex := strings.Index(response, "(")

    if openParenIndex == -1 {
        return nil, response
    }

    value := response[:openParenIndex]
    node := &Node{Value: value}
    response = response[openParenIndex+1:]

    for response != "" && response[0] != ')' {
        child, remaining := parseResponse(response)
        if child != nil {
            node.Children = append(node.Children, child)
        }
        response = remaining
    }

    if response != "" && response[0] == ')' {
        response = response[1:]
    }

	fmt.Println("value: ", value)
    return node, response
}

func printTree(node *Node, indent string) {
    fmt.Println(indent + node.Value)
    for _, child := range node.Children {
        printTree(child, indent+"    ")
    }
}

type RequestJSON struct {
    Input string `json:"input"`
}

type ResponseJSON struct {
    Output string `json:"output"`
}
