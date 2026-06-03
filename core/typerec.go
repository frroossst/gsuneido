// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package core

import (
	"sort"
	"strings"

	"github.com/apmckinlay/gsuneido/core/types"
)


// RecordTypes is an off-switch for the always-on recording. Default on.
var RecordTypes = true

const typeRetSlot = -1

type typerec map[*SuFunc]map[int]uint32

func (th *Thread) recordType(fn *SuFunc, slot int, v Value) {
	if !RecordTypes || v == nil || fn == nil {
		return
	}
	t := v.Type()
	if t >= types.N {
		return
	}
	rec := th.typerec
	if rec == nil {
		rec = make(typerec)
		th.typerec = rec
	}
	slots := rec[fn]
	if slots == nil {
		slots = make(map[int]uint32)
		rec[fn] = slots
	}
	slots[slot] |= 1 << uint(t)
}

func (th *Thread) recordParams(fr *Frame) {
	if !RecordTypes {
		return
	}
	fn := fr.fn
	n := min(int(fn.Nparams), len(fr.locals))
	for i := range n {
		th.recordType(fn, i, fr.locals[i])
	}
}

func (th *Thread) TypeProfile(reset bool) *SuObject {
	result := &SuObject{}
	for fn, slots := range th.typerec {
		cls := fn.ClassName
		// method Name is qualified (e.g. "TpTest.Foo"); strip the class prefix
		meth := strings.TrimPrefix(fn.Name, cls+".")
		if cls == "" {
			cls = meth // free function: no enclosing class
		}
		if cls == "" {
			cls, meth = "?", "?"
		}
		methOb := childOb(childOb(result, cls), meth)
		for slot, bits := range slots {
			if name := slotName(fn, slot); name != "" {
				methOb.Set(SuStr(name), typeSetOb(bits))
			}
		}
	}
	if reset {
		th.typerec = nil
	}
	return result
}

func childOb(parent *SuObject, key string) *SuObject {
	if v := parent.GetIfPresent(nil, SuStr(key)); v != nil {
		if ob, ok := v.(*SuObject); ok {
			return ob
		}
	}
	ob := &SuObject{}
	parent.Set(SuStr(key), ob)
	return ob
}

func typeSetOb(bits uint32) *SuObject {
	ob := &SuObject{}
	names := make([]string, 0, 4)
	for t := types.Type(0); t < types.N; t++ {
		if bits&(1<<uint(t)) != 0 {
			names = append(names, t.String())
		}
	}
	sort.Strings(names)
	for _, n := range names {
		ob.Add(SuStr(n))
	}
	return ob
}

func slotName(fn *SuFunc, slot int) string {
	if slot == typeRetSlot {
		return "$return"
	}
	if slot < 0 {
		return ""
	}
	return safeVarName(fn, slot)
}

func safeVarName(fn *SuFunc, slot int) (name string) {
	defer func() {
		if recover() != nil {
			name = ""
		}
	}()
	return fn.VarName(slot)
}
