package engine

import (
	"github.com/apmckinlay/gsuneido/typechecker/typeinfer/diagnostics"
)

// Aliases so the passes can talk about diagnostics without each importing the
// diagnostics package. Only what the passes actually use lives here; session.go
// calls diagnostics directly for the config and filtering it needs.
type (
	Diagnostic = diagnostics.Diagnostic
	Severity   = diagnostics.Severity
	Flag       = diagnostics.Flag
	Config     = diagnostics.Config
)

const (
	SeverityWarning = diagnostics.SeverityWarning
	SeverityError   = diagnostics.SeverityError

	FlagNone                    = diagnostics.FlagNone
	FlagStrictStringConcat      = diagnostics.FlagStrictStringConcat
	FlagStrictCrossTypeCompares = diagnostics.FlagStrictCrossTypeCompares
)

func ScoreConfidence(d *Diagnostic) float64 { return diagnostics.ScoreConfidence(d) }
