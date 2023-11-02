// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

//go:build windows && !portable

package builtin

import (
	"unsafe"

	"github.com/apmckinlay/gsuneido/builtin/goc"
	"github.com/apmckinlay/gsuneido/builtin/heap"
	. "github.com/apmckinlay/gsuneido/core"
)

var wtsapi32 = MustLoadDLL("wtsapi32.dll")

// dll void WTSAPI32:WTSFreeMemory(pointer adr)

var wtsFreeMemory = wtsapi32.MustFindProc("WTSFreeMemory").Addr()

func WTSFreeMemory(adr uintptr) {
	goc.Syscall1(wtsFreeMemory,
		adr)
}

// dll bool WTSAPI32:WTSQuerySessionInformation(pointer hServer, long SessionId,
// long WTSInfoClass, POINTER* ppBuffer, LONG* pBytesReturned)
var wtsQuerySessionInformation = wtsapi32.MustFindProc("WTSQuerySessionInformationA").Addr()

const WTS_CURRENT_SERVER_HANDLE = 0
const WTS_CURRENT_SESSION = uintptrMinusOne
const WTS_ClientProtocolType = 16
const WTS_SessionId = 4

func WTS_GetClientProtocolType() int {
	defer heap.FreeTo(heap.CurSize())
	pbuf := heap.Alloc(uintptrSize)
	psize := heap.Alloc(int32Size)
	rtn := goc.Syscall5(wtsQuerySessionInformation,
		WTS_CURRENT_SERVER_HANDLE,
		WTS_CURRENT_SESSION,
		WTS_ClientProtocolType,
		uintptr(pbuf),
		uintptr(psize))
	buf := *(*uintptr)(pbuf)
	size := *(*int32)(psize)
	if rtn == 0 || size != 2 || buf == 0 {
		return 0
	}
	data := *(*int16)(unsafe.Pointer(buf))
	WTSFreeMemory(buf)
	return int(data)
}

var _ = builtin(WTS_GetSessionId, "()")

func WTS_GetSessionId() Value {
	if WTS_GetClientProtocolType() == 0 {
		return Zero
	}
	defer heap.FreeTo(heap.CurSize())
	pbuf := heap.Alloc(uintptrSize)
	psize := heap.Alloc(int32Size)
	rtn := goc.Syscall5(wtsQuerySessionInformation,
		WTS_CURRENT_SERVER_HANDLE,
		WTS_CURRENT_SESSION,
		WTS_SessionId,
		uintptr(pbuf),
		uintptr(psize))
	buf := *(*uintptr)(pbuf)
	size := *(*int32)(psize)
	if rtn == 0 || size != 4 || buf == 0 {
		return Zero
	}
	data := *(*int32)(unsafe.Pointer(buf))
	WTSFreeMemory(buf)
	return IntVal(int(data))
}
