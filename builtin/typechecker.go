// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	"fmt"
	"maps"
	"slices"
	"sync"

	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/options"
	"github.com/apmckinlay/gsuneido/typechecker/typeinfer"
	"github.com/apmckinlay/gsuneido/typechecker/typeinfer/annotations"
)

type suTypeChecker struct {
	staticClass[suTypeChecker]
}

func init() {
	Global.Builtin("TypeChecker", &suTypeChecker{})
	// package level variable initializers - the builtin, method, and
	// staticMethod calls that build builtinTypeSignatures - all run before
	// any init function, so the signature table is complete by now
	// (including this class's own methods)
	loadTypeCheckerAnnotations()
}

func (*suTypeChecker) String() string {
	return "TypeChecker /* builtin class */"
}

func (tc *suTypeChecker) Equal(other any) bool {
	return tc == other
}

func (*suTypeChecker) Lookup(_ *Thread, method string) Value {
	return typecheckerMethods[method]
}

var typecheckerMethods = methods("typechecker")

// loadTypeCheckerAnnotations seeds the checker's builtin signature database.
// The checker resolves calls like Object.Size() against these, so without it
// every builtin call would be untyped.
func loadTypeCheckerAnnotations() {
	imported := make([]annotations.TypeSignature, len(builtinTypeSignatures))
	for i, ts := range builtinTypeSignatures {
		imported[i] = annotations.TypeSignature(ts)
	}
	typeinfer.LoadAnnotations(imported)
}

var _ = staticMethod(typechecker_Infer,
	"(arguments :object, references :object = #(), config :object = #()) :object")

func typechecker_Infer(arguments, references, config Value) Value {
	return runTypeChecker("TypeInfer", "Infer", arguments, references, config)
}

var _ = staticMethod(typechecker_InlayHints,
	"(arguments :object, references :object = #(), config :object = #()) :object")

func typechecker_InlayHints(arguments, references, config Value) Value {
	return runTypeChecker("TypeAnnotate", "InlayHints", arguments, references, config)
}

var _ = staticMethod(typechecker_Annotations, "() :object")

func typechecker_Annotations() Value {
	return builtinSignaturesOb()
}

var _ = staticMethod(typechecker_Members, "() :object")

// built on first call, not in a package level initializer: the staticMethod
// registrations above fill typecheckerMethods via a side effect the
// initialization order analysis cannot see, so a var initializer here would
// capture whichever methods happened to be registered by then.
func typechecker_Members() Value {
	typecheckerMembersOnce.Do(func() {
		// Members itself is introspection, not part of the API
		names := slices.Sorted(maps.Keys(typecheckerMethods))
		names = slices.DeleteFunc(names, func(s string) bool { return s == "Members" })
		typecheckerMembers = SuObjectOfStrs(names)
		typecheckerMembers.SetReadOnly()
		typecheckerMembers.SetConcurrent() // shared, but read-only so no locking
	})
	return typecheckerMembers
}

var typecheckerMembersOnce sync.Once
var typecheckerMembers *SuObject

// runTypeChecker mirrors the JSON protocol the standalone binary speaks:
// same two argument forms, same response envelope.
func runTypeChecker(method, meth string, arguments, references, config Value) Value {
	res, err := typeinfer.Process(typeinfer.Request{
		Method:     method,
		Arguments:  sourceEntries(arguments, meth, "arguments"),
		References: sourceEntries(references, meth, "references"),
		Config:     configMap(config),
	})
	if err != nil {
		panic("TypeChecker." + meth + ": " + err.Error())
	}
	ob := &SuObject{}
	ob.Put(nil, SuStr("method"), SuStr(res.Method))
	ob.Put(nil, SuStr("result"), resultsOb(res.Results, meth))
	ob.Put(nil, SuStr("diagnostics"), diagnosticsOb(res.Diagnostics))
	ob.Put(nil, SuStr("version"), SuStr(options.BuiltStr()))
	return ob
}

// sourceEntries accepts either wire form for each list element:
// a bare source string, or an object with src and optional name.
// Unnamed entries get Class0, Class1, ... to match the JSON protocol.
func sourceEntries(v Value, meth, kind string) []typeinfer.SourceEntry {
	ob := ToContainer(v)
	entries := make([]typeinfer.SourceEntry, ob.ListSize())
	for i := range entries {
		el := ob.ListGet(i)
		name := fmt.Sprintf("Class%d", i)
		if src, ok := el.ToStr(); ok {
			entries[i] = typeinfer.SourceEntry{Name: name, Src: src}
			continue
		}
		e, ok := el.ToContainer()
		if !ok {
			panic(fmt.Sprintf("TypeChecker.%s: %s[%d]: must be a string or an object with src",
				meth, kind, i))
		}
		src := e.GetIfPresent(nil, SuStr("src"))
		if src == nil {
			panic(fmt.Sprintf("TypeChecker.%s: %s[%d]: missing src", meth, kind, i))
		}
		if nm := e.GetIfPresent(nil, SuStr("name")); nm != nil && ToStr(nm) != "" {
			name = ToStr(nm)
		}
		entries[i] = typeinfer.SourceEntry{Name: name, Src: ToStr(src)}
	}
	return entries
}

// configMap reads the named members as strings. Unrecognized keys are ignored
// by the checker, and a bad value is reported by it, not here.
func configMap(v Value) map[string]string {
	ob := ToContainer(v)
	if ob.NamedSize() == 0 {
		return nil
	}
	cfg := make(map[string]string, ob.NamedSize())
	iter := ob.Iter2(false, true)
	for k, val := iter(); k != nil; k, val = iter() {
		cfg[ToStrOrString(k)] = ToStrOrString(val)
	}
	return cfg
}

// resultsOb converts one result per argument: TypeInfer yields inferred types,
// TypeAnnotate yields the source with annotations spliced in.
func resultsOb(results []any, meth string) Value {
	ob := &SuObject{}
	for _, r := range results {
		switch r := r.(type) {
		case typeinfer.TypeInfo:
			ob.Add(typeInfoOb(r))
		case string:
			ob.Add(SuStr(r))
		default:
			panic(fmt.Sprintf("TypeChecker.%s: unexpected result type %T", meth, r))
		}
	}
	return ob
}

func typeInfoOb(ti typeinfer.TypeInfo) Value {
	meths := &SuObject{}
	for _, name := range slices.Sorted(maps.Keys(ti.Methods)) {
		meths.Put(nil, SuStr(name), typeMapOb(ti.Methods[name]))
	}
	ob := &SuObject{}
	ob.Put(nil, SuStr("methods"), meths)
	ob.Put(nil, SuStr("members"), typeMapOb(ti.Members))
	return ob
}

// typeMapOb builds name -> type. Keys are sorted so repeated calls on the
// same source produce identical objects (Go map order is randomized).
func typeMapOb(m map[string]string) Value {
	ob := &SuObject{}
	for _, k := range slices.Sorted(maps.Keys(m)) {
		ob.Put(nil, SuStr(k), SuStr(m[k]))
	}
	return ob
}

func diagnosticsOb(ds typeinfer.DiagnosticSet) Value {
	ob := &SuObject{}
	ob.Put(nil, SuStr("errors"), diagListOb(ds.Errors))
	ob.Put(nil, SuStr("warnings"), diagListOb(ds.Warnings))
	return ob
}

func diagListOb(ds []typeinfer.ResultDiagnostic) Value {
	ob := &SuObject{}
	for _, d := range ds {
		e := &SuObject{}
		e.Put(nil, SuStr("class"), SuStr(d.Class))
		e.Put(nil, SuStr("method"), SuStr(d.Method))
		e.Put(nil, SuStr("pos"), IntVal(d.Pos))
		e.Put(nil, SuStr("line"), IntVal(d.Line))
		e.Put(nil, SuStr("col"), IntVal(d.Col))
		e.Put(nil, SuStr("msg"), SuStr(d.Msg))
		// omitted when there is no flag, as in the JSON protocol
		if flag := d.Flag.String(); flag != "" {
			e.Put(nil, SuStr("flag"), SuStr(flag))
		}
		ob.Add(e)
	}
	return ob
}

// the signature table never changes after startup, so build the object once
var builtinSignaturesOnce sync.Once
var builtinSignatures *SuObject

func builtinSignaturesOb() Value {
	builtinSignaturesOnce.Do(func() {
		builtinSignatures = &SuObject{}
		for _, ts := range builtinTypeSignatures {
			e := &SuObject{}
			e.Put(nil, SuStr("kind"), SuStr(ts.Kind))
			e.Put(nil, SuStr("prefix"), SuStr(ts.Prefix))
			e.Put(nil, SuStr("name"), SuStr(ts.Name))
			e.Put(nil, SuStr("sig"), SuStr(ts.Sig))
			builtinSignatures.Add(e)
		}
		builtinSignatures.SetReadOnly()
		builtinSignatures.SetConcurrent() // shared, but read-only so no locking
	})
	return builtinSignatures
}
