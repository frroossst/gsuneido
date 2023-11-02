// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package compile

import (
	"strings"

	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/util/regex"
)

// EvalString executes string containing Suneido code
// i.e. string.Eval()
func EvalString(th *Thread, s string) Value {
	s = strings.Trim(s, " \t\r\n")
	if isGlobal(s) {
		// optimize if just a global name
		return Global.GetName(th, s)
	}
	s = "function () {\n" + s + "\n}"
	fn := NamedConstant("", "eval", s, nil).(*SuFunc)
	return th.Call(fn)
}

// benchmark shows Suneido regex is faster than Go regexp for this
var rxGlobal = regex.Compile(`\A[A-Z][_a-zA-Z0-9]*?[!?]?\Z`)

func isGlobal(s string) bool {
	return rxGlobal.Matches(s)
}
