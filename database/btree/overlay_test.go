// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package btree

import (
	"math/rand"
	"sort"
	"testing"

	. "github.com/apmckinlay/gsuneido/util/hamcrest"
	"github.com/apmckinlay/gsuneido/util/str"
)

func TestEmptyOverlay(t *testing.T) {
	var data []string
	get := func(i uint64) string { return data[i] }
	fb := CreateFbtree(nil, get, 64)
	mb := newMbtree()
	mb2 := newMbtree()
	ov := &Overlay{under: []tree{fb, mb2}, mb: mb}
	checkIter(t, data, ov)

	const n = 100
	randKey := str.UniqueRandom(3, 8)

	data = insert(data, n, randKey, mb)
	checkIter(t, data, ov)

	data = insert(data, n, randKey, mb2)
	checkIter(t, data, ov)

	for i := 0; i < n/2; i++ {
		j := rand.Intn(len(data))
		if data[j] != "" {
			ov.Delete(data[j], key2off(data[j]))
			data[j] = ""
		}
	}
	checkIter(t, data, ov)
}

type insertable interface {
	Insert(key string, off uint64)
}

func insert(data []string, n int, randKey func() string, dest insertable) []string {
	for i := 0; i < n; i++ {
		key := randKey()
		off := key2off(key)
		data = append(data, key)
		dest.Insert(key, off)
	}
	return data
}

func key2off(key string) uint64 {
	off := uint64(0)
	for _, c := range key {
		off = off<<8 + uint64(c)
	}
	return off
}

func checkIter(t *testing.T, data []string, tr tree) {
	sort.Strings(data)
	it := tr.Iter()
	for _, k := range data {
		if k == "" {
			continue
		}
		k2, o2, ok := it()
		Assert(t).True(ok)
		Assert(t).That(k2, Equals(k))
		Assert(t).That(o2, Equals(key2off(k)))
	}
	_, _, ok := it()
	Assert(t).False(ok)
}
