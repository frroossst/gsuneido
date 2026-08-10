package engine

import (
	"github.com/apmckinlay/gsuneido/typechecker/typeinfer/diagnostics"
)

type (
	Diagnostic = diagnostics.Diagnostic
	Severity   = diagnostics.Severity
	Level      = diagnostics.Level
	Flag       = diagnostics.Flag
	Config     = diagnostics.Config
)

const (
	SeverityWarning = diagnostics.SeverityWarning
	SeverityError   = diagnostics.SeverityError

	LevelOff   = diagnostics.LevelOff
	LevelWarn  = diagnostics.LevelWarn
	LevelError = diagnostics.LevelError

	FlagNone                    = diagnostics.FlagNone
	FlagStrictStringConcat      = diagnostics.FlagStrictStringConcat
	FlagStrictCrossTypeCompares = diagnostics.FlagStrictCrossTypeCompares
)

func ParseLevel(s string) (Level, error) { return diagnostics.ParseLevel(s) }

func DefaultConfig() Config { return diagnostics.DefaultConfig() }

func FilterDiagnostics(diags []Diagnostic, cfg Config) []Diagnostic {
	return diagnostics.FilterDiagnostics(diags, cfg)
}

func ScoreConfidence(d *Diagnostic) float64 { return diagnostics.ScoreConfidence(d) }
