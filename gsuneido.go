// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package main // import "github.com/apmckinlay/gsuneido"

/*
CheckLibrary('stdlib')

TestRunner.Run(libs: #(stdlib), skipTags: #(gui, windows), quit_on_failure:);;
TestRunner.Run(TestObserverPrint(), libs: #(stdlib), skipTags: #(gui, windows));;

dlv debug -- -c ...
*/

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	_ "github.com/apmckinlay/gsuneido/aaainitfirst"
	"github.com/apmckinlay/gsuneido/builtin"
	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/database/db19"
	"github.com/apmckinlay/gsuneido/database/dbms"
	"github.com/apmckinlay/gsuneido/options"
	. "github.com/apmckinlay/gsuneido/runtime"
)

var builtDate = "Dec 16 2019" // set by: go build -ldflags "-X main.builtDate=..."

var help = `options:
	-c[lient] [ipaddress] (default 127.0.0.1)
	-l[oad] [table]
    -p[ort] # (default 3147)
    -r[epl]
    -v[ersion]`

// dbmsLocal is set if running with a local/standalone database.
var dbmsLocal IDbms
var mainThread *Thread

func main() {
	suneido := new(SuObject)
	suneido.SetConcurrent()
	Global.Builtin("Suneido", suneido)

	options.BuiltDate = builtDate
	if options.Action == "" {
		options.Action = "repl"
	}
	switch options.Action {
	case "load":
		if options.Arg == "" {
			n := db19.LoadDatabase("database.su", "suneido.db")
			println("loaded", n, "tables")
		} else {
			table := strings.TrimSuffix(options.Arg, ".su")
			n := db19.LoadTable(table, "suneido.db")
			println("loaded", n, "records to", table)
		}
		os.Exit(0)
	case "version":
		println("gSuneido " + builtDate + " (" + runtime.Version() + " " +
			runtime.GOARCH + " " + runtime.GOOS + ")")
		os.Exit(0)
	case "help":
		println(help)
		os.Exit(0)
	case "error":
		println(options.Error)
		os.Exit(1)
	case "repl", "client":
		// handled below
	default:
		println("invalid action:", options.Action)
		os.Exit(1)
	}
	Libload = libload // dependency injection
	mainThread = NewThread()
	mainThread.Poll = true
	builtin.UIThread = mainThread
	defer mainThread.Close()
	// dependency injection of GetDbms
	if options.Action == "client" {
		addr := options.Arg + ":" + options.Port
		GetDbms = func() IDbms { return dbms.NewDbmsClient(addr) }
		clientErrorLog()
	} else {
		dbmsLocal = dbms.NewDbmsLocal()
		GetDbms = func() IDbms { return dbmsLocal }
		eval("Suneido.Print = PrintStdout;;")
	}
	if options.Action == "repl" {
		repl()
	} else {
		eval("Init()")
		builtin.Run()
	}
}

func clientErrorLog() {
	dbms := mainThread.Dbms()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix(dbms.SessionId("") + " ")

	// unlike cSuneido, client error.log is still in current directory
	// this is partly because stderr has already been redirected
	f, err := os.Open(options.Errlog)
	if err != nil {
		return
	}
	defer func() {
		f.Close()
		os.Truncate(options.Errlog, 0) // can't remove since open as stderr
		if e := recover(); e != nil {
			dbms.Log("log previous errors: " + fmt.Sprint(e))
		}
	}()
	// send errors to server
	in := bufio.NewScanner(f)
	in.Buffer(nil, 1024)
	nlines := 0
	for in.Scan() {
		dbms.Log("PREVIOUS: " + in.Text())
		if nlines++; nlines > 1000 {
			dbms.Log("PREVIOUS: too many errors")
			break
		}
	}
}

// REPL -------------------------------------------------------------

var prompt = func(s string) { fmt.Fprintln(os.Stderr, s) }

func repl() {
	log.SetFlags(log.Ltime)
	log.SetPrefix("")

	if !isTerminal() {
		prompt = func(string) {}
	}

	builtin.Def()
	builtin.Concat()

	built := builtin.Built()
	if options.Action == "client" {
		built += " - client"
	}
	prompt(built)
	showOptions()
	prompt("Press Enter twice (i.e. blank line) to execute, q to quit")
	r := bufio.NewReader(os.Stdin)
	for {
		prompt("~~~")
		src := ""
		for {
			line, err := r.ReadString('\n')
			line = strings.TrimRight(line, " \t\r\n")
			if line == "q" || (err != nil && (err != io.EOF || src == "")) {
				return
			}
			if line == "" {
				break
			}
			src += line + "\n"
		}
		eval(src)
	}
}

func isTerminal() bool {
	fm, _ := os.Stdout.Stat()
	mode := fm.Mode()
	return mode&os.ModeDevice == os.ModeDevice &&
		mode&os.ModeCharDevice == os.ModeCharDevice
}

func showOptions() {
	if options.ThreadDisabled {
		prompt("- Thread disabled")
	}
	if options.TimersDisabled {
		prompt("- Timers disabled")
	}
	if options.ClearCallbackDisabled {
		prompt("- ClearCallback disabled")
	}
}

func eval(src string) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("ERROR:", e)
			if internal(e) {
				debug.PrintStack()
				fmt.Println("---")
				PrintStack(mainThread.Callstack())
			} else if se, ok := e.(*SuExcept); ok {
				PrintStack(se.Callstack)
			} else {
				PrintStack(mainThread.Callstack())
			}
		}
	}()
	src = "function () {\n" + src + "\n}"
	v, results := compile.Checked(mainThread, src)
	for _, s := range results {
		fmt.Println(s)
	}
	fn := v.(*SuFunc)
	// DisasmMixed(os.Stdout, fn, src)

	mainThread.Reset()
	result := mainThread.Start(fn, nil)
	if result != nil {
		fmt.Println(WithType(result)) // NOTE: doesn't use ToString
	}
}

type internalError interface {
	RuntimeError()
}

func internal(e interface{}) bool {
	_, ok := e.(internalError)
	return ok
}

// libload loads a name from the dbms
func libload(t *Thread, gn Gnum, name string) (result Value) {
	defer func() {
		if e := recover(); e != nil {
			// fmt.Println("INFO: error loading", name, e)
			// debug.PrintStack()
			Global.Set(gn, nil)
			panic("error loading " + name + " " + fmt.Sprint(e))
		}
	}()
	defs := t.Dbms().LibGet(name)
	if len(defs) == 0 {
		// fmt.Println("LOAD", name, "MISSING")
		return nil
	}
	for i := 0; i < len(defs); i += 2 {
		lib := defs[i]
		src := defs[i+1]
		if s, ok := LibraryOverrides[lib+":"+name]; ok {
			src = s
		}
		// want to pass the name from the start (rather than adding after)
		// so it propagates to nested Named values
		result = compile.NamedConstant(lib, name, src)
		Global.Set(gn, result) // required for overload inheritance
		// fmt.Println("LOAD", name, "SUCCEEDED")
	}
	return
}
