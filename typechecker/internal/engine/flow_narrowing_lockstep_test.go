package engine

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"

	"github.com/apmckinlay/gsuneido/typechecker/internal/synth"
)

// enabled for the whole test binary: any test that runs the narrowing pass
// also enforces the lockstep invariant.
func init() {
	lockstepCheck = func(sc narrowScope, site string) {
		for k, in := range sc.InGuard {
			if !in {
				continue
			}
			if _, ok := sc.Types[k]; !ok {
				panic(fmt.Sprintf("narrowScope lockstep broken at %s: InGuard[%q] set without Types entry", site, k))
			}
		}
		for k, in := range sc.MemberInGuard {
			if !in {
				continue
			}
			if _, ok := sc.Members[k]; !ok {
				panic(fmt.Sprintf("narrowScope lockstep broken at %s: MemberInGuard[%q] set without Members entry", site, k))
			}
		}
	}
}

// a lockstep violation panics and surfaces through safeRun as a pipeline panic
func TestPropNarrowScopeLockstep(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		src := genSource(rt, synth.Config{})
		runGenerated(rt, src)
	})
}
