// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package core

import (
	"testing"

	"github.com/apmckinlay/gsuneido/util/assert"
)

func TestArgSpecString(t *testing.T) {
	test := func(as *ArgSpec, expected string) {
		t.Helper()
		assert.T(t).This(as.String()).Is(expected)
	}
	test(&ArgSpec0, "ArgSpec()")
	test(&ArgSpec3, "ArgSpec(?, ?, ?)")
	test(&ArgSpecEach0, "ArgSpec(@)")
	test(&ArgSpecEach1, "ArgSpec(@+1)")
	test(&ArgSpecBlock, "ArgSpec(block:)")
	test(&ArgSpec{Nargs: 0, Spec: []byte{2, 0, 1}, Names: vals("a", "b", "c")},
		"ArgSpec(c:, a:, b:)")
	test(&ArgSpec{Nargs: 4, Spec: []byte{2, 1}, Names: vals("a", "b", "c", "d")},
		"ArgSpec(?, ?, c:, b:)")
}

func TestArgSpecEqual(t *testing.T) {
	as := []*ArgSpec{
		&ArgSpec0,
		&ArgSpec4,
		&ArgSpecEach0,
		&ArgSpecEach1,
		&ArgSpecBlock,
		{Nargs: 2, Spec: []byte{0, 1}, Names: []Value{SuStr("foo"), SuStr("bar")}},
		{Nargs: 2, Spec: []byte{0, 1}, Names: []Value{SuStr("foo"), SuStr("baz")}},
	}
	for i, x := range as {
		for j, y := range as {
			assert.T(t).This(x.Equal(y)).Is(i == j)
		}
	}
}
