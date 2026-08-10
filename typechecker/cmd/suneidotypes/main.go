package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/apmckinlay/gsuneido/builtin"

	"github.com/frroossst/SuneidoTypes/internal/tlog"
	"github.com/frroossst/SuneidoTypes/typeinfer"
	"github.com/frroossst/SuneidoTypes/typeinfer/annotations"
)

func loadAnnotationDatabase() {
	sigs := builtin.BuiltinTypeSignatures()
	imported := make([]annotations.TypeSignature, len(sigs))
	for i, ts := range sigs {
		imported[i] = annotations.TypeSignature(ts)
	}
	typeinfer.LoadAnnotations(imported)
}

// tlogWriter adapts tlog for net/http's ErrorLog (see runServer).
type tlogWriter struct{}

func (tlogWriter) Write(p []byte) (int, error) {
	tlog.Logf("[http] %s", bytes.TrimSpace(p))
	return len(p), nil
}

// value determined via linker
var builtDate = "dev build"

type jsonRequest struct {
	Method     string            `json:"method"`
	Arguments  []json.RawMessage `json:"arguments"`
	References []json.RawMessage `json:"references,omitempty"`
	Config     map[string]string `json:"config,omitempty"`
}

type jsonResponse struct {
	Method      string                  `json:"method"`
	Result      []any                   `json:"result"`
	Diagnostics typeinfer.DiagnosticSet `json:"diagnostics"`
	Version     string                  `json:"version"`
}

func parseEntries(raws []json.RawMessage, kind string) ([]typeinfer.SourceEntry, error) {
	entries := make([]typeinfer.SourceEntry, len(raws))
	for i, raw := range raws {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			entries[i] = typeinfer.SourceEntry{Name: fmt.Sprintf("Class%d", i), Src: s}
			continue
		}
		var obj struct {
			Name string `json:"name"`
			Src  string `json:"src"`
		}
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, fmt.Errorf("%s[%d]: must be string or {name, src}", kind, i)
		}
		if obj.Src == "" {
			return nil, fmt.Errorf("%s[%d]: missing src", kind, i)
		}
		if obj.Name == "" {
			obj.Name = fmt.Sprintf("Class%d", i)
		}
		entries[i] = typeinfer.SourceEntry{Name: obj.Name, Src: obj.Src}
	}
	return entries, nil
}

func processRequest(req jsonRequest) (jsonResponse, error) {
	args, err := parseEntries(req.Arguments, "argument")
	if err != nil {
		return jsonResponse{}, err
	}
	refs, err := parseEntries(req.References, "reference")
	if err != nil {
		return jsonResponse{}, err
	}
	res, err := typeinfer.Process(typeinfer.Request{
		Method:     req.Method,
		Arguments:  args,
		References: refs,
		Config:     req.Config,
	})
	if err != nil {
		return jsonResponse{}, err
	}
	return jsonResponse{
		Method:      res.Method,
		Result:      res.Results,
		Diagnostics: res.Diagnostics,
		Version:     builtDate,
	}, nil
}

func runJSON() error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		tlog.Logf("[ERROR] reading stdin: %v", err)
		return fmt.Errorf("reading stdin: %w", err)
	}
	var req jsonRequest
	if err := json.Unmarshal(data, &req); err != nil {
		tlog.Logf("[ERROR] parsing request (%d bytes): %v", len(data), err)
		return fmt.Errorf("parsing request: %w", err)
	}
	resp, err := processRequest(req)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(resp)
}

func runServer() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux, ErrorLog: log.New(tlogWriter{}, "", 0)}

	mux.HandleFunc("/check", handleCheck)
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
		}()
	})

	go func() {
		_, _ = io.Copy(io.Discard, os.Stdin)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	fmt.Printf("READY port=%d\n", port)
	tlog.Logf("server READY port=%d build=%s", port, builtDate)

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		tlog.Logf("[ERROR] server Serve: %v", err)
		return err
	}
	tlog.Logf("server stopped port=%d", port)
	return nil
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	defer func() {
		if rec := recover(); rec != nil {
			tlog.Logf("[ERROR] check panicked: %v", rec)
			http.Error(w, fmt.Sprintf("check panicked: %v", rec),
				http.StatusInternalServerError)
		}
	}()
	var req jsonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		tlog.Logf("[ERROR] decode request from %s: %v", r.RemoteAddr, err)
		http.Error(w, fmt.Sprintf("decode: %v", err), http.StatusBadRequest)
		return
	}
	resp, err := processRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resp); err != nil {
		tlog.Logf("[ERROR] encode response to %s: %v", r.RemoteAddr, err)
	}
}

func main() {
	serve := flag.Bool("serve", false, "run as a long-lived HTTP server on 127.0.0.1; prints 'READY port=N' to stdout on startup")
	flag.Parse()

	if err := run(*serve); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(serve bool) error {
	tlog.Setup()
	defer tlog.Close()

	loadAnnotationDatabase()

	if serve {
		return runServer()
	}
	return runJSON()
}
