// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

//go:build windows && !portable

package builtin

import (
	"log"
	"syscall"

	"github.com/apmckinlay/gsuneido/builtin/goc"
	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/dbg"
)

// rogsChan is used by other threads to Run code On the Go Side UI thread
// Need buffer so we can send to channel and then notifyCside
var rogsChan = make(chan func(), 1)

// UpdateUI runs the block on the main UI thread
var _ = builtin(UpdateUI, "(block)")

// UpdateUI runs the block on the main UI thread
// The block will be run in one of two ways:
// If executing in the interpreter in MainThread,
// it periodically calls runOnGoSide.
// If executing in the C message loop,
// the cside timer will trigger a call back to runOnGoSide.
func UpdateUI(th *Thread, args []Value) Value {
	block := args[0]
	if th == MainThread {
		th.Call(block)
	} else {
		block.SetConcurrent()
		rogsChan <- func() { runUI(block) }
	}
	return nil
}

const notifyMsg = WM_USER

// notifyCside is used by SetTimer, and KillTimer
// It uses PostMessage (high priority) to C side
// to handle when we're running in the message loop.
func notifyCside() {
	// NOTE: this has to be the Go Syscall, not goc.Syscall
	r, _, _ := syscall.SyscallN(postMessage,
		goc.CHelperHwnd(), notifyMsg, 0, 0)
	if r == 0 {
		log.Panicln("notifyCside PostMessage failed")
	}
}

// runOnGoSide is called by interp via runtime.RunOnGoSide
// and cside via goc.RunOnGoSide
func runOnGoSide() {
	assert.That(InRunUI == false)
	InRunUI = true
	defer func() { InRunUI = false }()
	for range 8 { // process available messages, but not forever
		select {
		case fn := <-rogsChan:
			fn()
		default: // non-blocking
			return
		}
	}
}

func runUI(block Value) {
	state := MainThread.GetState()
	defer func() {
		if e := recover(); e != nil {
			log.Println("ERROR in UpdateUI:", e)
			MainThread.PrintStack()
			dbg.PrintStack()
		}
		MainThread.RestoreState(state)
	}()
	MainThread.Call(block)
}
