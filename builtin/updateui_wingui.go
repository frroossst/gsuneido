// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

//go:build windows && !portable

package builtin

import (
	"log"
	"syscall"

	"github.com/apmckinlay/gsuneido/builtin/goc"
	. "github.com/apmckinlay/gsuneido/core"
	"golang.org/x/sys/windows"
)

// rogsChan is used by other threads to Run code On the Go Side UI thread
// Need buffer so we can send to channel and then notifyCside
var rogsChan = make(chan func(), 1)

// UpdateUI runs the block on the main UI thread
var _ = builtin(UpdateUI, "(block)")

func UpdateUI(th *Thread, args []Value) Value {
	if windows.GetCurrentThreadId() == uiThreadId {
		Synchronized(th, args)
	} else {
		block := args[0]
		block.SetConcurrent()
		rogsChan <- func() { runUI(block) }
		notifyCside()
	}
	return nil
}

const notifyWparam = 0xffffffff

// notifyCside is used by UpdateUI, SetTimer, and KillTimer
func notifyCside() {
	// NOTE: this has to be the Go Syscall, not goc.Syscall
	r, _, _ := syscall.SyscallN(postMessage,
		goc.CHelperHwnd(), WM_USER, notifyWparam, 0)
	if r == 0 {
		log.Panicln("notifyCside PostMessage failed")
	}
}

// runOnGoSide is called from goc.RunOnGoSide and runtime.RunOnGoSide
func runOnGoSide() {
	for {
		select {
		case fn := <-rogsChan:
			fn()
		default: // non-blocking
			return
		}
	}
}

var updateThread *Thread

func runUI(block Value) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("ERROR in UpdateUI:", e)
			UIThread.PrintStack()
		}
	}()
	if updateThread == nil {
		updateThread = UIThread.SubThread()
	}
	updateThread.Call(block)
}
