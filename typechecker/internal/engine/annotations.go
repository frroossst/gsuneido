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

func Annotations() AnnotationSet {
	m := make(AnnotationSet, len(builtinAnnotations))
	for name, sigs := range builtinAnnotations {
		m[name] = append([]Signature(nil), sigs...)
	}
	return m
}

func ParseTypeAnnotation(s string) (DynType, error) {
	return typealgebra.ParseAnnotation(s)
}

func signatureFromAst(fn *ast.Function) (Signature, error) {
	return annotations.SignatureFromAst(fn)
}
