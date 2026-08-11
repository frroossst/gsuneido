package diagnostics

import (
	"github.com/apmckinlay/gsuneido/typechecker/typeinfer/typealgebra"
)

// error = some clean (non-?) arm is provably bad; evidence that is nothing but a guess caps at warning. a dirty union can still error on its clean arms.
type Severity int

const (
	SeverityWarning Severity = iota
	SeverityError
)

type Diagnostic struct {
	Severity Severity
	Method   string
	Pos      int
	Msg      string
	Flag     Flag
	// operand types, so the scorer reads dirtiness without parsing Msg
	Got []typealgebra.DynType
	// non-zero = per-rule precision, overrides the severity/dirtiness bucket
	Confidence float64
}
