package diagnostics_test

import (
	"testing"

	"github.com/apmckinlay/gsuneido/typechecker/diagnostics"
	. "github.com/apmckinlay/gsuneido/typechecker/typealgebra"
)

func TestScoreConfidence(t *testing.T) {
	cleanUnion := Union{Types: []DynType{TFalse, TNumber}}
	dirtyUnion := Union{Types: []DynType{TFalse, TNumber}, IsDirty: true}

	cases := []struct {
		name string
		d    diagnostics.Diagnostic
		want float64
	}{
		{"flagged style error is floored", diagnostics.Diagnostic{Severity: diagnostics.SeverityError, Flag: diagnostics.FlagStrictStringConcat, Got: []DynType{TNumber}}, 0.20},
		{"concrete error", diagnostics.Diagnostic{Severity: diagnostics.SeverityError, Got: []DynType{TString}}, 0.90},
		{"error with no operand type", diagnostics.Diagnostic{Severity: diagnostics.SeverityError}, 0.90},
		{"clean-union-arm error", diagnostics.Diagnostic{Severity: diagnostics.SeverityError, Got: []DynType{cleanUnion}}, 0.70},
		{"dirty error sinks", diagnostics.Diagnostic{Severity: diagnostics.SeverityError, Got: []DynType{dirtyUnion}}, 0.40},
		{"bare-unknown error sinks", diagnostics.Diagnostic{Severity: diagnostics.SeverityError, Got: []DynType{TUnknown}}, 0.40},
		{"plain warning", diagnostics.Diagnostic{Severity: diagnostics.SeverityWarning, Got: []DynType{TString}}, 0.35},
		{"dirty warning", diagnostics.Diagnostic{Severity: diagnostics.SeverityWarning, Got: []DynType{dirtyUnion}}, 0.25},
		{"explicit confidence override wins", diagnostics.Diagnostic{Severity: diagnostics.SeverityWarning, Confidence: 0.70}, 0.70},
		{"override beats flag floor", diagnostics.Diagnostic{Severity: diagnostics.SeverityError, Flag: diagnostics.FlagStrictStringConcat, Confidence: 0.70}, 0.70},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := diagnostics.ScoreConfidence(&tc.d); got != tc.want {
				t.Errorf("ScoreConfidence = %v, want %v", got, tc.want)
			}
		})
	}
}
