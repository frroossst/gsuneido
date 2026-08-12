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
	"github.com/apmckinlay/gsuneido/typechecker"
	"github.com/apmckinlay/gsuneido/typechecker/annotations"
)

type suTypeChecker struct {
	staticClass[suTypeChecker]
}

func init() {
	Global.Builtin("TypeChecker", &suTypeChecker{})
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

func loadTypeCheckerAnnotations() {
	imported := make([]annotations.TypeSignature, len(builtinTypeSignatures))
	for i, ts := range builtinTypeSignatures {
		imported[i] = annotations.TypeSignature(ts)
	}
	typechecker.LoadAnnotations(imported)
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

func typechecker_Members() Value {
	typecheckerMembersOnce.Do(func() {
		names := slices.Sorted(maps.Keys(typecheckerMethods))
		names = slices.DeleteFunc(names, func(s string) bool { return s == "Members" })
		typecheckerMembers = SuObjectOfStrs(names)
		typecheckerMembers.SetReadOnly()
		typecheckerMembers.SetConcurrent() // ok since read-only
	})
	return typecheckerMembers
}

var typecheckerMembersOnce sync.Once
var typecheckerMembers *SuObject

func runTypeChecker(method, meth string, arguments, references, config Value) Value {
	res, err := typechecker.Process(typechecker.Request{
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

// each element is a source string, or an object with src and optional name
func sourceEntries(v Value, meth, kind string) []typechecker.SourceEntry {
	ob := ToContainer(v)
	entries := make([]typechecker.SourceEntry, ob.ListSize())
	for i := range entries {
		el := ob.ListGet(i)
		name := fmt.Sprintf("Class%d", i)
		if src, ok := el.ToStr(); ok {
			entries[i] = typechecker.SourceEntry{Name: name, Src: src}
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
		entries[i] = typechecker.SourceEntry{Name: name, Src: ToStr(src)}
	}
	return entries
}

// bad keys and values are reported by the checker, not here
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

func resultsOb(results []any, meth string) Value {
	ob := &SuObject{}
	for _, r := range results {
		switch r := r.(type) {
		case typechecker.TypeInfo:
			ob.Add(typeInfoOb(r))
		case string:
			ob.Add(SuStr(r))
		default:
			panic(fmt.Sprintf("TypeChecker.%s: unexpected result type %T", meth, r))
		}
	}
	return ob
}

func typeInfoOb(ti typechecker.TypeInfo) Value {
	meths := &SuObject{}
	for _, name := range slices.Sorted(maps.Keys(ti.Methods)) {
		meths.Put(nil, SuStr(name), typeMapOb(ti.Methods[name]))
	}
	ob := &SuObject{}
	ob.Put(nil, SuStr("methods"), meths)
	ob.Put(nil, SuStr("members"), typeMapOb(ti.Members))
	return ob
}

// sorted for idempotency
func typeMapOb(m map[string]string) Value {
	ob := &SuObject{}
	for _, k := range slices.Sorted(maps.Keys(m)) {
		ob.Put(nil, SuStr(k), SuStr(m[k]))
	}
	return ob
}

func diagnosticsOb(ds typechecker.DiagnosticSet) Value {
	ob := &SuObject{}
	ob.Put(nil, SuStr("errors"), diagListOb(ds.Errors))
	ob.Put(nil, SuStr("warnings"), diagListOb(ds.Warnings))
	return ob
}

func diagListOb(ds []typechecker.ResultDiagnostic) Value {
	ob := &SuObject{}
	for _, d := range ds {
		e := &SuObject{}
		e.Put(nil, SuStr("class"), SuStr(d.Class))
		e.Put(nil, SuStr("method"), SuStr(d.Method))
		e.Put(nil, SuStr("pos"), IntVal(d.Pos))
		e.Put(nil, SuStr("line"), IntVal(d.Line))
		e.Put(nil, SuStr("col"), IntVal(d.Col))
		e.Put(nil, SuStr("msg"), SuStr(d.Msg))
		if flag := d.Flag.String(); flag != "" { // omitted when there is none
			e.Put(nil, SuStr("flag"), SuStr(flag))
		}
		ob.Add(e)
	}
	return ob
}

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
		builtinSignatures.SetConcurrent() // ok since read-only
	})
	return builtinSignatures
}
