package typeinfer

import (
	"encoding/json"
	"fmt"
	"maps"
	"sync/atomic"
	"time"

	"github.com/frroossst/SuneidoTypes/internal/tlog"
	"github.com/frroossst/SuneidoTypes/typeinfer/annotations"
	"github.com/frroossst/SuneidoTypes/typeinfer/diagnostics"
	"github.com/frroossst/SuneidoTypes/typeinfer/internal/annotate"
	"github.com/frroossst/SuneidoTypes/typeinfer/internal/engine"
	"github.com/frroossst/SuneidoTypes/typeinfer/typealgebra"
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
	Methods map[string]map[string]string `json:"methods"`
	Members map[string]string            `json:"members"`
}

// ResultDiagnostic is one reported finding, positioned in its class's source.
type ResultDiagnostic struct {
	Class  string           `json:"class"`
	Method string           `json:"method"`
	Pos    int              `json:"pos"`
	Line   int              `json:"line"`
	Col    int              `json:"col"`
	Msg    string           `json:"msg"`
	Flag   diagnostics.Flag `json:"flag,omitempty"`
}

type DiagnosticSet struct {
	Errors   []ResultDiagnostic `json:"errors"`
	Warnings []ResultDiagnostic `json:"warnings"`
}

type Result struct {
	Method      string
	Results     []any
	Diagnostics DiagnosticSet
}

func LoadAnnotations(imported []annotations.TypeSignature) {
	engine.LoadAnnotations(imported)
}

var reqSeq atomic.Uint64

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

// stringifyTypes reports types in annotation spelling, the same language
// TypeAnnotate splices into the source.
func stringifyTypes(in map[string]typealgebra.DynType) map[string]string {
	out := make(map[string]string, len(in))
	for name, ty := range in {
		out[name] = ty.String()
	}
	return out
}

// stringifyVarTypes is stringifyTypes for MethodVarTypes' nested per-method map.
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

func Process(req Request) (res Result, err error) {
	reqID := reqSeq.Add(1)
	start := time.Now()
	curClass := ""

	if reqJSON, mErr := json.Marshal(req); mErr == nil {
		tlog.Logf("[req %d] request method=%q args=%d refs=%d: %s",
			reqID, req.Method, len(req.Arguments), len(req.References), reqJSON)
	}

	defer func() {
		if err != nil {
			where := ""
			if curClass != "" {
				where = fmt.Sprintf(" class=%q", curClass)
			}
			tlog.Logf("[ERROR] [req %d]%s failed after %s: %v", reqID, where, time.Since(start), err)
			return
		}
		tlog.Logf("[req %d] response method=%q classes=%d errors=%d warnings=%d total=%s",
			reqID, res.Method, len(res.Results),
			len(res.Diagnostics.Errors), len(res.Diagnostics.Warnings), time.Since(start))
	}()

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
	regs := engine.BuildReferenceRegistry(refs, func(format string, args ...any) {
		tlog.Logf("[ERROR] [req %d] "+format, append([]any{reqID}, args...)...)
	})

	parsed := make([]*engine.ClassObject, len(req.Arguments))
	for i, a := range req.Arguments {
		parsed[i] = engine.NewClassObject(a.Name, engine.ParseClass(a.Src))
	}
	env := engine.NewTypeEnv()
	regs.Seed(&env)
	pipeline := engine.DefaultPipeline()
	parentReturns := map[string]typealgebra.DynType{}
	results := make([]any, len(parsed))
	var collected []rankedDiag
	for i, c := range parsed {
		curClass = req.Arguments[i].Name
		t := time.Now()
		pipeline.Run(c, env, parentReturns)
		tlog.Logf("[req %d] class[%d/%d]=%q pipeline.Run bytes=%d took=%s",
			reqID, i+1, len(parsed), curClass, len(req.Arguments[i].Src), time.Since(t))
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
						Class:  curClass,
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
	curClass = "" // past the per-class work; nothing below is class-specific

	return Result{
		Method:      req.Method,
		Results:     results,
		Diagnostics: rankDiagnostics(collected, confFilter),
	}, nil
}
