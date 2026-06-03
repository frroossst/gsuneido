// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	. "github.com/apmckinlay/gsuneido/core"
)

var _ = builtin(TypeProfile, "(reset = false)")

// TypeProfile returns the runtime types observed so far on the current
// thread as a nested object: class -> method -> variable -> set of type
// names (variable "$return" holds the return type). Pass reset:true to
// clear the buffer after reading. See core/typerec.go.
func TypeProfile(th *Thread, args []Value) Value {
	return th.TypeProfile(args[0] == True)
}
