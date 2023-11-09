package main

import (
    "encoding/json"
    "fmt"
    "time"
    "net/http"

	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/tokens"
)

// RequestJSON is a struct to hold the request data.
type RequestJSON struct {
    Input string `json:"input"`
}

// ResponseJSON is a struct to hold the response data.
type ResponseJSON struct {
    Output string `json:"output"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var req RequestJSON
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Perform some processing on the input string.
	p := compile.AstParser(req.Input)
    
    currentTime := time.Now()
    formattedTime := currentTime.Format(time.RFC3339)
    fmt.Println("[", formattedTime, "]")
    fmt.Println("input: ", req.Input)
	ast := p.Const()
    fmt.Println("output: ", ast)
    fmt.Println("=========================================")
	if p.Token != tokens.Eof {
        // TODO: return error
		p.Error("did not parse all input")
	}
    output := ast.String()

    response := ResponseJSON{Output: output}
    jsonResponse, err := json.Marshal(response)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonResponse)
}

func main() {
    http.HandleFunc("/process", handler)

    // Start the server on port 8080 without SSL
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Error starting server: ", err)
        return
    }
}
