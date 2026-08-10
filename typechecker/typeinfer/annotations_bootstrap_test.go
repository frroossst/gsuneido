package typeinfer

import (
	"github.com/apmckinlay/gsuneido/builtin"

	"github.com/apmckinlay/gsuneido/typechecker/typeinfer/annotations"
)

func init() {
	sigs := builtin.BuiltinTypeSignatures()
	imported := make([]annotations.TypeSignature, len(sigs))
	for i, ts := range sigs {
		imported[i] = annotations.TypeSignature(ts)
	}
	LoadAnnotations(imported)
}
