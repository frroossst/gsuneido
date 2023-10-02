// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package runtime

import (
	"fmt"
	"log"

	"github.com/apmckinlay/gsuneido/options"
)

func Warning(args ...any) {
	s := fmt.Sprintln(args...)
	if options.WarningsThrow.Matches(s) {
		panic(s)
	} else {
		log.Print("WARNING ", s)
	}
}
