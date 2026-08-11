// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	"strings"
	"testing"

	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/util/assert"
)

// callTypeChecker goes through Global and Lookup, the path Suneido code takes
func callTypeChecker(th *Thread, method string, args ...Value) Value {
	tc := Global.GetName(th, "TypeChecker")
	f := tc.Lookup(th, method)
	if f == nil {
		panic("no such method: TypeChecker." + method)
	}
	for _, a := range args {
		th.Push(a)
	}
	specs := []*ArgSpec{&ArgSpec0, &ArgSpec1, &ArgSpec2, &ArgSpec3}
	return f.Call(th, tc, specs[len(args)])
}

func namedSrc(name, src string) Value {
	e := &SuObject{}
	e.Put(nil, SuStr("name"), SuStr(name))
	e.Put(nil, SuStr("src"), SuStr(src))
	return SuObjectOf(e)
}

func get(v Value, member string) Value {
	return ToContainer(v).GetIfPresent(nil, SuStr(member))
}

const adder = `class { Add(x = 0) { y = "abc".Size(); return x + y } }`

func TestTypeChecker_Infer(t *testing.T) {
	assert := assert.T(t)
	res := callTypeChecker(&Thread{}, "Infer", namedSrc("Adder", adder))

	assert.This(get(res, "method")).Is(SuStr("TypeInfer"))
	result := ToContainer(get(res, "result"))
	assert.This(result.ListSize()).Is(1)

	// y comes from "abc".Size(), so this also proves the builtin signature
	// database was seeded - without it Size() would be untyped
	add := get(get(result.ListGet(0), "methods"), "Add")
	assert.This(get(add, "x")).Is(SuStr("number"))
	assert.This(get(add, "y")).Is(SuStr("number"))

	assert.This(ToContainer(get(get(res, "diagnostics"), "errors")).ListSize()).Is(0)
}

func TestTypeChecker_InlayHints(t *testing.T) {
	assert := assert.T(t)
	res := callTypeChecker(&Thread{}, "InlayHints", namedSrc("Adder", adder))

	assert.This(get(res, "method")).Is(SuStr("TypeAnnotate"))
	result := ToContainer(get(res, "result"))
	assert.This(result.ListSize()).Is(1)

	annotated := ToStr(result.ListGet(0))
	assert.Msg(annotated).That(strings.Contains(annotated, "Add(x :number = 0) :number"))
	assert.Msg(annotated).That(strings.Contains(annotated, "y /* number */"))
}

// a bare string is the other wire form; it gets a generated name
func TestTypeChecker_BareStringArgument(t *testing.T) {
	assert := assert.T(t)
	res := callTypeChecker(&Thread{}, "Infer", SuObjectOf(SuStr(adder)))

	result := ToContainer(get(res, "result"))
	assert.This(result.ListSize()).Is(1)
	assert.This(get(get(get(result.ListGet(0), "methods"), "Add"), "x")).Is(SuStr("number"))
}

func TestTypeChecker_Diagnostics(t *testing.T) {
	assert := assert.T(t)
	res := callTypeChecker(&Thread{}, "Infer",
		namedSrc("Bad", `class { F() { return 1 + Object() } }`))

	errs := ToContainer(get(get(res, "diagnostics"), "errors"))
	assert.This(errs.ListSize()).Is(1)
	d := errs.ListGet(0)
	assert.This(get(d, "class")).Is(SuStr("Bad"))
	assert.This(get(d, "method")).Is(SuStr("F"))
	assert.This(get(d, "line")).Is(IntVal(1))
	assert.Msg(ToStr(get(d, "msg"))).
		That(strings.Contains(ToStr(get(d, "msg")), `operator "+" expects number, got object`))
	// no strict-mode flag on this one, so the member is absent
	assert.This(get(d, "flag")).Is(nil)
}

func TestTypeChecker_Annotations(t *testing.T) {
	assert := assert.T(t)
	ann := ToContainer(callTypeChecker(&Thread{}, "Annotations"))
	assert.That(ann.ListSize() > 100)

	// TypeChecker's own methods must be present: they are registered by
	// package level var initializers, which all run before the init that
	// hands this table to the checker
	var sawInfer, sawZlibCompress bool
	for i := range ann.ListSize() {
		e := ann.ListGet(i)
		kind, prefix := ToStr(get(e, "kind")), ToStr(get(e, "prefix"))
		name, sig := ToStr(get(e, "name")), ToStr(get(e, "sig"))
		if prefix == "typechecker" && name == "Infer" {
			sawInfer = true
			assert.This(kind).Is("static")
			assert.That(strings.HasPrefix(sig, "(arguments :object"))
		}
		if prefix == "zlib" && name == "Compress" {
			sawZlibCompress = true
			assert.This(sig).Is("(string :string) :string")
		}
	}
	assert.Msg("TypeChecker.Infer missing from Annotations").That(sawInfer)
	assert.Msg("Zlib.Compress missing from Annotations").That(sawZlibCompress)
}

// regression: a var initializer would capture only the methods registered
// before it, which is how Ftsearch, OpenPGP, and PdfEncrypt lose entries.
// Members itself is introspection, so it is not listed.
func TestTypeChecker_Members(t *testing.T) {
	assert := assert.T(t)
	members := callTypeChecker(&Thread{}, "Members")
	assert.This(members.String()).Is(`#("Annotations", "Infer", "InlayHints")`)
	assert.This(ToContainer(members).ListSize()).Is(len(typecheckerMethods) - 1)
}

func TestTypeChecker_ConfigAndReferences(t *testing.T) {
	assert := assert.T(t)
	cfg := &SuObject{}
	cfg.Put(nil, SuStr("strictStringConcat"), SuStr("off"))
	cfg.Put(nil, SuStr("strictCrossTypeCompares"), SuStr("off"))
	refs := SuObjectOf(SuStr(`class { Helper() { return 1 } }`))

	res := callTypeChecker(&Thread{}, "Infer",
		namedSrc("Adder", adder), refs, cfg)
	assert.This(get(res, "method")).Is(SuStr("TypeInfer"))
	assert.This(ToContainer(get(res, "result")).ListSize()).Is(1)
}

func TestTypeChecker_BadArguments(t *testing.T) {
	assert.T(t).This(func() {
		callTypeChecker(&Thread{}, "Infer", SuObjectOf(SuObjectOf(SuStr("no src member"))))
	}).Panics("missing src")
}
