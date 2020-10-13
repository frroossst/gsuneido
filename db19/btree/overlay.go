// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package btree

import (
	"github.com/apmckinlay/gsuneido/db19/ixspec"
	"github.com/apmckinlay/gsuneido/db19/stor"
	"github.com/apmckinlay/gsuneido/util/assert"
)

type treeIter = func() (string, uint64, bool)

type tree interface {
	Iter(check bool) treeIter
}

// Overlay is an immutable fbtree plus one or more mbtrees.
// Update transactions have a mutable mbtree at the top to store updates.
type Overlay struct {
	// under are the underlying fbtree and mbtree's
	under []tree
	// mb is the mutable top mbtree, nil if read-only
	mb *mbtree
}

func NewOverlay(store *stor.Stor, is *ixspec.T) *Overlay {
	assert.That(is != nil)
	return &Overlay{under: []tree{CreateFbtree(store, is)}}
}

// var Under [8]int64

// Mutable returns a modifiable copy of an Overlay
func (ov *Overlay) Mutable(tranNum int) *Overlay {
	assert.That(ov.mb == nil)
	under := append([]tree(nil), ov.under...) // copy
	// atomic.AddInt64(&Under[len(under)], 1)
	return &Overlay{under: under, mb: newMbtree(tranNum)}
}

func (ov *Overlay) GetIxspec() *ixspec.T {
	return ov.base().ixspec
}

func (ov *Overlay) SetIxspec(is *ixspec.T) {
	fb := *ov.base() // copy
	fb.ixspec = is
	ov.under[0] = &fb
}

// Insert inserts into the mutable top mbtree
func (ov *Overlay) Insert(key string, off uint64) {
	ov.mb.Insert(key, off)
}

const tombstone = 1 << 63

// Delete either deletes the key/offset from the mutable mbtree
// or inserts a tombstone into the mutable mbtree.
func (ov *Overlay) Delete(key string, off uint64) {
	if !ov.mb.Delete(key, off) {
		// key not present
		ov.mb.Insert(key, off|tombstone)
	}
}

func (ov *Overlay) Check(fn func(uint64)) int {
	n, _, _ := ov.base().check(fn)
	return n
}

func (ov *Overlay) QuickCheck() {
	ov.base().quickCheck()
}

func (ov *Overlay) base() *fbtree {
	return ov.under[0].(*fbtree)
}

// iter -------------------------------------------------------------

type ovsrc struct {
	iter treeIter
	key  string
	off  uint64
	ok   bool
}

// Iter returns a treeIter function
func (ov *Overlay) Iter(check bool) treeIter {
	if ov.mb == nil && len(ov.under) == 1 {
		// only fbtree, no merge needed
		return ov.under[0].Iter(check)
	}
	srcs := make([]ovsrc, 0, len(ov.under)+1)
	if ov.mb != nil {
		srcs = append(srcs, ovsrc{iter: ov.mb.Iter(check)})
	}
	for i := range ov.under {
		srcs = append(srcs, ovsrc{iter: ov.under[i].Iter(check)})
	}
	for i := range srcs {
		srcs[i].next()
	}
	return func() (string, uint64, bool) {
		i := ovsrcNext(srcs)
		key, off, ok := srcs[i].key, srcs[i].off, srcs[i].ok
		srcs[i].next()
		return key, off >> 1, ok
	}
}

func (src *ovsrc) next() {
	src.key, src.off, src.ok = src.iter()
	// adjust offset so tombstone comes first
	src.off = (src.off << 1) | ((src.off >> 63) ^ 1)
}

// ovsrcNext returns the index of the next element
func ovsrcNext(srcs []ovsrc) int {
	min := 0
	for {
		for i := 1; i < len(srcs); i++ {
			if ovsrcLess(&srcs[i], &srcs[min]) {
				min = i
			}
		}
		if !isTombstone(srcs[min].off) {
			return min
		}
		// skip over the insert matching the tombstone
		for i := range srcs {
			if i != min &&
				srcs[i].key == srcs[min].key && srcs[i].off&^1 == srcs[min].off {
				srcs[i].next()
			}
		}
		srcs[min].next() // skip the tombstone itself
	}
}

func isTombstone(off uint64) bool {
	return (off & 1) == 0
}

func ovsrcLess(x, y *ovsrc) bool {
	if !x.ok {
		return false
	}
	return !y.ok || x.key < y.key || (x.key == y.key && x.off < y.off)
}

//-------------------------------------------------------------------

func (ov *Overlay) StorSize() int {
	return 5 + 1 + 5
}

func (ov *Overlay) Write(w *stor.Writer) {
	fb := ov.base()
	w.Put5(fb.root).Put1(fb.treeLevels).Put5(fb.redirsOff)
}

// ReadOverlay reads an Overlay from storage BUT without ixspec
func ReadOverlay(st *stor.Stor, r *stor.Reader) *Overlay {
	root := r.Get5()
	treeLevels := r.Get1()
	redirsOff := r.Get5()
	return &Overlay{under: []tree{OpenFbtree(st, root, treeLevels, redirsOff)}}
}

//-------------------------------------------------------------------

// UpdateWith takes the mbtree updates from ov2 and adds them as a new layer to ov
func (ov *Overlay) UpdateWith(latest *Overlay) {
	// reuse the new slice and overwrite ov.under with the latest
	ov.under = append(ov.under[:0], latest.under...)
	// add mbtree updates
	ov.Freeze()
}

func (ov *Overlay) Freeze() {
	ov.under = append(ov.under, ov.mb)
	ov.mb = nil
}

//-------------------------------------------------------------------

// Merge merges the mbtree for tranNum (if there is one) into the fbtree
func (ov *Overlay) Merge(tranNum int) *Overlay {
	assert.That(ov.mb == nil)
	if len(ov.under) == 1 {
		panic("merge: missing overlay")
	}
	mb := ov.under[1].(*mbtree)
	if mb.tranNum != tranNum {
		panic("merge: wrong tranNum")
	}
	fb := ov.base()
	fb = Merge(fb, mb)
	return &Overlay{under: []tree{fb}}
}

func (ov *Overlay) WithMerged(ov2 *Overlay) *Overlay {
	// ov2.under[0] is the new fbtree from Merge
	// ov2.under[1] is the mbtree that we merged in
	ov2.under = append(ov2.under, ov.under[2:]...)
	return ov2
}

//-------------------------------------------------------------------

// Save writes the Overlay's base fbtree to storage
// and returns the new fbtree (in an Overlay) to later pass to With
func (ov *Overlay) Save(flatten bool) *Overlay {
	assert.That(ov.mb == nil)
	fb := ov.base()
	fb = fb.Save(flatten)
	return &Overlay{under: []tree{fb}}
}

// WithSaved returns a new Overlay,
// combining the current state (ov) with the updated fbtree (in ov2)
func (ov *Overlay) WithSaved(ov2 *Overlay) *Overlay {
	// ov2.under[0] is the new fbtree from Save
	ov2.under = append(ov2.under, ov.under[1:]...)
	return ov2
}