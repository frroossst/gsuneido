// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package core

import (
	"github.com/apmckinlay/gsuneido/core/types"
	"github.com/apmckinlay/gsuneido/util/assert"
)

// SuIter is a Value that wraps a runtime.Iter
// and provides the Suneido interator interface (Next,Dup,Infinite)
// returning itself when it reaches the end
type SuIter struct {
	ValueBase[SuIter]
	Iter
}

// Value interface --------------------------------------------------

var _ Value = (*SuIter)(nil)

// IterMethods is set by builtin/iter.go
var IterMethods Methods

func (SuIter) Lookup(_ *Thread, method string) Value {
	return IterMethods[method]
}

func (SuIter) Type() types.Type {
	return types.Iterator
}

func (it SuIter) Equal(other any) bool {
	return it == other
}

func (it SuIter) SetConcurrent() {
	it.Iter.SetConcurrent()
}

func (it SuIter) IsConcurrent() Value {
	return it.Iter.IsConcurrent()
}

//-------------------------------------------------------------------

type SuIter2 struct {
	ValueBase[SuIter2]
	iter2
}

var _ Value = (*SuIter2)(nil)

func (it SuIter2) Equal(other any) bool {
	panic(assert.ShouldNotReachHere())
}
