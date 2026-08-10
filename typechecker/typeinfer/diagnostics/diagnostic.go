package diagnostics

import (
	"github.com/frroossst/SuneidoTypes/typeinfer/typealgebra"
)

// error = some clean (non-?) arm is provably bad; evidence that is nothing but a guess caps at warning. a dirty union can still error on its clean arms.
type Severity int

const (
	SeverityWarning Severity = iota
	SeverityError
)

type Diagnostic struct {
	Severity   Severity              `json:"severity"`
	Method     string                `json:"method"`
	Pos        int                   `json:"pos"`
	Msg        string                `json:"msg"`
	Flag       Flag                  `json:"flag,omitempty"`
	Got        []typealgebra.DynType `json:"-"` // operand types, so the scorer reads dirtiness without parsing Msg
	Confidence float64               `json:"-"` // non-zero = per-rule precision, overrides the severity/dirtiness bucket
}
