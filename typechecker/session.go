package typechecker

import (
	"fmt"
	"maps"

	"github.com/apmckinlay/gsuneido/typechecker/annotations"
	"github.com/apmckinlay/gsuneido/typechecker/diagnostics"
	"github.com/apmckinlay/gsuneido/typechecker/internal/annotate"
	"github.com/apmckinlay/gsuneido/typechecker/internal/engine"
	"github.com/apmckinlay/gsuneido/typechecker/typealgebra"
)

// SourceEntry is one named Suneido source.
type SourceEntry struct {
	Name string
	Src  string
}

type Request struct {
	Method     string // "TypeInfer" or "TypeAnnotate"
	Arguments  []SourceEntry
	References []SourceEntry
	Config     map[string]string
}

type TypeInfo struct {
	Methods map[string]map[string]string
	Members map[string]string
}

type ResultDiagnostic struct {
	Class  string
	Method string
	Pos    int
	Line   int
	Col    int
	Msg    string
	Flag   diagnostics.Flag
}

type DiagnosticSet struct {
	Errors   []ResultDiagnostic
	Warnings []ResultDiagnostic
}

type Result struct {
	Method      string
	Results     []any
	Diagnostics DiagnosticSet
}

func LoadAnnotations(imported []annotations.TypeSignature) {
	engine.LoadAnnotations(imported)
}

func buildConfig(raw map[string]string) (diagnostics.Config, error) {
	cfg := diagnostics.DefaultConfig()
	if v, ok := raw["strictStringConcat"]; ok {
		l, err := diagnostics.ParseLevel(v)
		if err != nil {
			return cfg, fmt.Errorf("config.strictStringConcat: %w", err)
		}
		cfg.StrictStringConcat = l
	}
	if v, ok := raw["strictCrossTypeCompares"]; ok {
		l, err := diagnostics.ParseLevel(v)
		if err != nil {
			return cfg, fmt.Errorf("config.strictCrossTypeCompares: %w", err)
		}
		cfg.StrictCrossTypeCompares = l
	}
	return cfg, nil
}

func stringifyTypes(in map[string]typealgebra.DynType) map[string]string {
	out := make(map[string]string, len(in))
	for name, ty := range in {
		out[name] = ty.String()
	}
	return out
}

func stringifyVarTypes(in map[string]map[string]typealgebra.DynType) map[string]map[string]string {
	out := make(map[string]map[string]string, len(in))
	for method, vars := range in {
		out[method] = stringifyTypes(vars)
	}
	return out
}

// converts a byte offset into 1-based line/column for the result.
func offsetToLineCol(src string, off int) (line, col int) {
	if off < 0 {
		off = 0
	}
	if off > len(src) {
		off = len(src)
	}
	line, col = 1, 1
	for i := 0; i < off; i++ {
		if src[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return
}

func parseArgument(a SourceEntry) (co *engine.ClassObject, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s: %v", a.Name, e)
		}
	}()
	return engine.NewClassObject(a.Name, engine.ParseClass(a.Src)), nil
}

func Process(req Request) (Result, error) {
	if req.Method != "TypeInfer" && req.Method != "TypeAnnotate" {
		return Result{}, fmt.Errorf("unknown method: %q (expected TypeInfer or TypeAnnotate)", req.Method)
	}

	cfg, err := buildConfig(req.Config)
	if err != nil {
		return Result{}, err
	}

	confFilter, err := parseConfidenceFilter(req.Config)
	if err != nil {
		return Result{}, err
	}

	refs := make([]engine.RefSource, len(req.References))
	for i, r := range req.References {
		refs[i] = engine.RefSource{Name: r.Name, Src: r.Src}
	}
	regs := engine.BuildReferenceRegistry(refs)

	parsed := make([]*engine.ClassObject, len(req.Arguments))
	for i, a := range req.Arguments {
		if parsed[i], err = parseArgument(a); err != nil {
			return Result{}, err
		}
	}

	env := engine.NewTypeEnv()
	regs.Seed(&env)
	pipeline := engine.DefaultPipeline()
	parentReturns := map[string]typealgebra.DynType{}
	results := make([]any, len(parsed))
	var collected []rankedDiag
	for i, c := range parsed {
		class := req.Arguments[i].Name
		pipeline.Run(c, env, parentReturns)
		switch req.Method {
		case "TypeInfer":
			results[i] = TypeInfo{
				Methods: stringifyVarTypes(engine.MethodVarTypes(c, env)),
				Members: stringifyTypes(engine.MemberTypes(c, env)),
			}
		case "TypeAnnotate":
			results[i] = annotate.AnnotateClass(req.Arguments[i].Src, env, c)
		}
		if env.Diagnostics != nil {
			filtered := diagnostics.FilterDiagnostics(*env.Diagnostics, cfg)
			for _, d := range filtered {
				line, col := offsetToLineCol(req.Arguments[i].Src, d.Pos)
				collected = append(collected, rankedDiag{
					severity:   d.Severity,
					confidence: diagnostics.ScoreConfidence(&d),
					entry: ResultDiagnostic{
						Class:  class,
						Method: d.Method,
						Pos:    d.Pos,
						Line:   line,
						Col:    col,
						Msg:    d.Msg,
						Flag:   d.Flag,
					},
				})
			}
			*env.Diagnostics = (*env.Diagnostics)[:0]
		}
		parentReturns = maps.Clone(env.Returns)
	}

	return Result{
		Method:      req.Method,
		Results:     results,
		Diagnostics: rankDiagnostics(collected, confFilter),
	}, nil
}
