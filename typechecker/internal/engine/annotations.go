package engine

import (
	"github.com/apmckinlay/gsuneido/compile/ast"

	"github.com/apmckinlay/gsuneido/typechecker/annotations"
	"github.com/apmckinlay/gsuneido/typechecker/typealgebra"
)

type (
	AnnotationSet = annotations.Set
	Signature     = annotations.Signature
	Param         = annotations.Param
)

var builtinAnnotations = AnnotationSet{}

func LoadAnnotations(imported []annotations.TypeSignature) {
	builtinAnnotations, _ = annotations.Load(imported)
}

// shared, not copied - passes only read it, which
// TestPropSharedAnnotationsNotMutated enforces. Written once by LoadAnnotations
// at startup, so concurrent runs are reading a table nobody writes.
func Annotations() AnnotationSet {
	return builtinAnnotations
}

func ParseTypeAnnotation(s string) (DynType, error) {
	return typealgebra.ParseAnnotation(s)
}

func signatureFromAst(fn *ast.Function) (Signature, error) {
	return annotations.SignatureFromAst(fn)
}
