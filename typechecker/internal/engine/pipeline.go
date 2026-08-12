package engine

import (
	"fmt"
	"sort"
	"strings"
)

type Pipeline struct {
	Annotations AnnotationSet
}

const maxFixpointPasses = 64

func envTypeSignature(env TypeEnv) string {
	sig := func(t DynType) string {
		if t == nil {
			return "-"
		}
		return t.String()
	}
	parts := make([]string, 0, len(env.Nodes)+len(env.Params)+len(env.Returns)+len(env.Members))
	for n, ty := range env.Nodes {
		parts = append(parts, fmt.Sprintf("n%p=%s", n, sig(ty)))
	}
	for p, ty := range env.Params {
		parts = append(parts, fmt.Sprintf("a%p=%s", p, sig(ty)))
	}
	for k, v := range env.Returns {
		parts = append(parts, "r:"+k+"="+sig(v))
	}
	for k, v := range env.Members {
		parts = append(parts, "m:"+k+"="+sig(v))
	}
	for k, v := range env.PostCtorMembers {
		parts = append(parts, "o:"+k+"="+sig(v))
	}
	for k, v := range env.PreCtorReturns {
		parts = append(parts, "q:"+k+"="+sig(v))
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

func iterateToFixpoint(env TypeEnv, body func()) {
	prev := envTypeSignature(env)
	for range maxFixpointPasses {
		body()
		cur := envTypeSignature(env)
		if cur == prev {
			return
		}
		prev = cur
	}
}

func DefaultPipeline() Pipeline {
	return Pipeline{Annotations: Annotations()}
}

func (p Pipeline) Run(cls *ClassObject, env TypeEnv, parentReturns map[string]DynType) {
	if parentReturns == nil {
		parentReturns = map[string]DynType{}
	}

	// sigs must be bound before the first CallsiteResolutionPass
	env = env.WithClass(cls, buildMethodSigs(cls))

	DateNarrowingPass(cls, env)

	LocalInference(cls, env)
	NameResolutionPass(cls, env)
	MemberAssignmentPass(cls, env)
	ReturnUnionPass(cls, env)

	iterateToFixpoint(env, func() {
		CallsiteResolutionPass(cls, env, p.Annotations)
		SuperCallsiteResolutionPass(cls, env, parentReturns)
		NameResolutionPass(cls, env)
		MemberAssignmentPass(cls, env)
		ReturnUnionPass(cls, env)
	})

	MemberDirtyPass(cls, env)
	// MemberAssignmentPass is deliberately NOT re-run after MemberDirtyPass
	NameResolutionPass(cls, env)
	ReturnUnionPass(cls, env)

	// after member unions have settled, before the narrowing phase; re-stamp only if it pinned something
	if AssertMemberPass(cls, env) {
		NameResolutionPass(cls, env)
		ReturnUnionPass(cls, env)
	}

	if ConstructorExecPass(cls, env) {
		env.CaptureSeedReturns()
		CallsiteResolutionPass(cls, env, p.Annotations)
		ReturnUnionPass(cls, env)
	}

	// NameResolutionPass is intentionally omitted from this loop (it would undo the narrowing)
	iterateToFixpoint(env, func() {
		FlowNarrowingPass(cls, env)
		RefreshTrinaryTypes(cls, env)
		CallsiteResolutionPass(cls, env, p.Annotations)
		SuperCallsiteResolutionPass(cls, env, parentReturns)
		clearReturnDiagnostics(env) // scrub warnings from earlier iterations that later ones invalidated
		ReturnUnionPass(cls, env)
	})

	// check passes from here down: order-independent among themselves, strictly after the narrowing loop

	// taint locals fed by guesses so the checks below can downgrade
	GuessTaintPass(cls, env)

	// after taint (so guessed sigs are excluded), before the checks that
	// consume the inferred param requirements
	RequirementPass(cls, env)

	// after the narrowing loop so RHS stamps inside guards are narrowed.
	AssertAssignmentCheckPass(cls, env)

	TypeCheckPass(cls, env)

	// after FlowNarrowing (above) so narrowed condition types are visible.
	BooleanConditionCheckPass(cls, env)

	CallsiteCheckPass(cls, env, p.Annotations)

	ArityCheckPass(cls, env)

	StaticMemberCheckPass(cls, env)

	CapabilityCheckPass(cls, env)

	ComputeSummaries(cls, env)
}

// keep in sync with the "return type ..." diagnostics ReturnUnionPass emits
func clearReturnDiagnostics(env TypeEnv) {
	if env.Diagnostics == nil {
		return
	}
	kept := (*env.Diagnostics)[:0]
	for _, d := range *env.Diagnostics {
		if !strings.HasPrefix(d.Msg, "return type ") {
			kept = append(kept, d)
		}
	}
	*env.Diagnostics = kept
}

func extensionThisType(className string) DynType {
	switch className {
	case "Strings":
		return TString
	case "Numbers":
		return TNumber
	case "Dates":
		return TDate
	case "Objects", "Records":
		return TObject
	}
	return nil
}
