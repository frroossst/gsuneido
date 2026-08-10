package main

import (
	"encoding/json"
	"testing"
)

func TestProcessRequest_WireForms(t *testing.T) {
	arg, _ := json.Marshal(map[string]string{
		"name": "Adder",
		"src":  `class { Add(x = 0) { return x + 1 } }`,
	})
	bare, _ := json.Marshal(`class { Ok() { } }`)
	resp, err := processRequest(jsonRequest{
		Method:    "TypeInfer",
		Arguments: []json.RawMessage{arg, bare},
	})
	if err != nil {
		t.Fatalf("processRequest: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("want 2 results, got %d", len(resp.Result))
	}
	if resp.Method != "TypeInfer" || resp.Version == "" {
		t.Fatalf("bad envelope: %+v", resp)
	}
	if resp.Diagnostics.Errors == nil || resp.Diagnostics.Warnings == nil {
		t.Fatal("diagnostics slices must be non-nil for the wire")
	}

	// malformed entry surfaces as an error naming the slot
	if _, err := processRequest(jsonRequest{
		Method:    "TypeInfer",
		Arguments: []json.RawMessage{json.RawMessage(`{"name": "X"}`)},
	}); err == nil {
		t.Fatal("want error for entry with missing src")
	}
}
