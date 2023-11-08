// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

//go:build !windows || portable

package builtin

import (
	"os"

	. "github.com/apmckinlay/gsuneido/core"
)

func init() {
	c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	Interrupt = func() bool {
		select {
		case <-c:
			return true
		default:
			return false
		}
	}
}

func Run() {
}

var _ = builtin(GetComputerName, "()")

func GetComputerName() Value {
	name, err := os.Hostname()
	if err != nil {
		panic("GetComputerName " + err.Error())
	}
	return SuStr(name)
}

func CallbacksCount() int {
	return 0
}

func WndProcCount() int {
	return 0
}

func GetGuiResources() (int, int) {
	return 0, 0
}

func ErrlogDir() string {
	return os.TempDir()
}

func OnUIThread() bool {
	return false
}
