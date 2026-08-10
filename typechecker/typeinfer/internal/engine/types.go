package engine

import (
	"github.com/apmckinlay/gsuneido/core"

	"github.com/apmckinlay/gsuneido/typechecker/typeinfer/typealgebra"
)

type (
	DynType   = typealgebra.DynType
	Primitive = typealgebra.Primitive
	Union     = typealgebra.Union
	Instance  = typealgebra.Instance
)

const (
	TUnknown  = typealgebra.TUnknown
	TVoid     = typealgebra.TVoid
	TBoolean  = typealgebra.TBoolean
	TFalse    = typealgebra.TFalse
	TTrue     = typealgebra.TTrue
	TNumber   = typealgebra.TNumber
	TString   = typealgebra.TString
	TDate     = typealgebra.TDate
	TFunction = typealgebra.TFunction
	TBlock    = typealgebra.TBlock
	TClass    = typealgebra.TClass
	TObject   = typealgebra.TObject
	TSequence = typealgebra.TSequence
)

func U(a, b DynType) DynType { return typealgebra.U(a, b) }

func markDirty(t DynType) DynType { return typealgebra.MarkDirty(t) }

func DynTypeOfSuValue(val core.Value) DynType { return typealgebra.DynTypeOfSuValue(val) }
