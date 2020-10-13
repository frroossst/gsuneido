// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package meta

import (
	"github.com/apmckinlay/gsuneido/db19/btree"
	"github.com/apmckinlay/gsuneido/db19/stor"
	"github.com/apmckinlay/gsuneido/util/assert"
)

// Meta is the layered info and schema metadata
// dif is per transaction, top is recently changed, old is the rest
// dif may override top which may override old
type Meta struct {
	difInfo   InfoHamt
	topInfo   InfoHamt
	oldInfo   *InfoPacked
	topSchema SchemaHamt
	oldSchema *SchemaPacked
	offInfo   uint64
	offSchema uint64
	genSchema uint32
	genInfo   uint32
}

func NewMeta(oldSchema *SchemaPacked, oldInfo *InfoPacked,
	topSchema SchemaHamt, topInfo InfoHamt) *Meta {
	return &Meta{
		oldSchema: oldSchema,
		oldInfo:   oldInfo,
		topInfo:   topInfo,
		topSchema: topSchema,
		difInfo:   InfoHamt{},
	}
}

func CreateMeta(store *stor.Stor) *Meta {
	return &Meta{
		oldInfo:   &InfoPacked{stor: store},
		oldSchema: &SchemaPacked{stor: store},
		topInfo:   InfoHamt{},
		topSchema: SchemaHamt{},
	}
}

// Mutable returns a mutable copy of a Meta
func (m *Meta) Mutable() *Meta {
	assert.That(m.difInfo.IsNil())
	ov2 := *m // copy
	ov2.difInfo = InfoHamt{}.Mutable()
	return &ov2
}

func (m *Meta) GetRoInfo(table string) *Info {
	if ti, ok := m.difInfo.Get(table); ok {
		return ti
	}
	ti, ok := m.topInfo.Get(table)
	if ok {
		if ti.isTomb() {
			return nil
		}
		var ti2 = *ti // copy
		ti = &ti2
	} else {
		ti, ok = m.oldInfo.Get(table)
		if !ok {
			return nil
		}
	}
	// set up index overlays and ixspecs
	ti.Indexes = append(ti.Indexes[:0:0], ti.Indexes...) // copy
	for i := range ti.Indexes {
		if ti.Indexes[i].GetIxspec() == nil {
			ix := *ti.Indexes[i] // copy
			ts := m.GetRoSchema(table)
			ix.SetIxspec(&ts.Indexes[i].Ixspec)
			ti.Indexes[i] = &ix
		}
	}
	if !m.difInfo.IsNil() {
		m.difInfo.Put(ti) // cache in memory
	}
	return ti
}

func (m *Meta) GetRwInfo(table string, tranNum int) *Info {
	if pti, ok := m.difInfo.Get(table); ok {
		return pti // already have mutable
	}
	var ti Info
	if pti, ok := m.topInfo.Get(table); ok {
		if ti.isTomb() {
			return nil
		}
		ti = *pti // copy
	} else if pti, ok := m.oldInfo.Get(table); ok {
		ti = *pti // copy
	} else {
		return nil
	}
	// start at 0 since these are deltas
	ti.Nrows = 0
	ti.Size = 0
	ti.mutable = true

	// set up index overlays and ixspecs
	ti.Indexes = append(ti.Indexes[:0:0], ti.Indexes...) // copy
	for i := range ti.Indexes {
		ti.Indexes[i] = ti.Indexes[i].Mutable(tranNum)
		if ti.Indexes[i].GetIxspec() == nil {
			ts := m.GetRoSchema(table)
			is := &ts.Indexes[i].Ixspec
			ti.Indexes[i].SetIxspec(is)
		}
	}

	m.difInfo.Put(&ti) // cache in memory
	return &ti
}

func (m *Meta) GetRoSchema(table string) *Schema {
	if ts, ok := m.topSchema.Get(table); ok {
		if ts.isTomb() {
			return nil
		}
		return ts
	}
	if ts, ok := m.oldSchema.Get(table); ok {
		return ts
	}
	return nil
}

// Put is used by Database.LoadedTable
func (m *Meta) Put(ts *Schema, ti *Info) *Meta {
	topSchema := m.topSchema.Mutable()
	topSchema.Put(ts)
	topInfo := m.topInfo.Mutable()
	topInfo.Put(ti)
	ov2 := *m // copy
	ov2.topSchema = topSchema.Freeze()
	ov2.topInfo = topInfo.Freeze()
	return &ov2
}

func (m *Meta) DropTable(table string) *Meta {
	assert.That(m.difInfo.IsNil())
	_, old := m.oldSchema.Get(table)
	topSchema := m.topSchema.Mutable()
	if old {
		// can't update old packed, so output tombstone
		topSchema.Put(newSchemaTomb(table))
	} else {
		if !topSchema.Delete(table) {
			return nil
		}
	}
	_, old = m.oldInfo.Get(table)
	topInfo := m.topInfo.Mutable()
	if old {
		// can't update old packed, so output tombstone
		topInfo.Put(newInfoTomb(table))
	} else {
		topInfo.Delete(table)
	}
	ov2 := *m // copy
	ov2.topSchema = topSchema.Freeze()
	ov2.topInfo = topInfo.Freeze()
	return &ov2
}

func (m *Meta) ForEachSchema(fn func(*Schema)) {
	m.topSchema.ForEach(fn)
	m.oldSchema.ForEach(func(sc *Schema) {
		// skip the ones already processed from roSchema
		if _, ok := m.topSchema.Get(sc.Table); !ok {
			fn(sc)
		}
	})
}

func (p SchemaPacked) ForEach(fn func(*Schema)) {
	if p.buf == nil {
		return
	}
	r := stor.NewReader(p.buf)
	r.Get3() // size
	nitems := r.Get2()
	if nitems == 0 {
		return
	}
	nfingers := 1 + nitems/perFingerSchema
	for i := 0; i < nfingers; i++ {
		r.Get3() // skip the fingers
	}
	for ; nitems > 0; nitems-- {
		fn(ReadSchema(p.stor, r))
	}
}

//-------------------------------------------------------------------

// LayeredOnto layers the mutable mbtree's from a transaction
// onto the latest/current state and returns a new state.
// Also, the nrows and size deltas are applied.
// Note: this does not merge the btrees, that is done later by merge.
// Nor does it save the changes to disk, that is done later by persist.
func (m *Meta) LayeredOnto(latest *Meta) *Meta {
	// start with a copy of the latest hash table because it may have more
	assert.That(latest.difInfo.IsNil())
	topInfo2 := latest.topInfo.Mutable()
	m.difInfo.ForEach(func(ti *Info) {
		if ti.mutable {
			lti, ok := topInfo2.Get(ti.Table)
			if !ok {
				lti, ok = latest.oldInfo.Get(ti.Table)
			}
			if !ok || lti.isTomb() {
				return
			}
			ti.Nrows += lti.Nrows
			ti.Size += lti.Size
			for i := range ti.Indexes {
				ti.Indexes[i].UpdateWith(lti.Indexes[i])
			}
			topInfo2.Put(ti)
		}
	})
	result := *latest // copy
	result.topInfo = topInfo2.Freeze()
	return &result
}

//-------------------------------------------------------------------

const Noffsets = 4

type offsets = [Noffsets]uint64

func (m *Meta) Write(st *stor.Stor) offsets {
	assert.That(m.difInfo.IsNil())
	offs := offsets{
		m.oldSchema.Offset(),
		m.oldInfo.Offset(),
		m.offSchema,
		m.offInfo,
	}
	if m.topSchema.generation != m.genSchema {
		offs[2] = m.topSchema.Write(st)
		m.genSchema = m.topSchema.generation
		m.offSchema = offs[2]
	}
	if m.topInfo.generation != m.genInfo {
		offs[3] = m.topInfo.Write(st)
		m.genInfo = m.topInfo.generation
		m.offInfo = offs[3]
	}
	return offs
}

func ReadMeta(st *stor.Stor, offs offsets) *Meta {
	m := Meta{
		oldSchema: NewSchemaPacked(st, offs[0]),
		oldInfo:   NewInfoPacked(st, offs[1]),
		topSchema: ReadSchemaHamt(st, offs[2]),
		topInfo:   ReadInfoHamt(st, offs[3]),
	}
	m.offSchema = offs[2]
	m.genSchema = m.topSchema.generation
	m.offInfo = offs[3]
	m.genInfo = m.topInfo.generation
	return &m
}

//-------------------------------------------------------------------

// Merge is called by state.Merge
// to merge the mbtree's for tranNum into the fbtree's.
// It collect updates which are then applied by ApplyMerge
func (m *Meta) Merge(tranNum int) []update {
	return m.topInfo.process(func(bto btOver) btOver {
		return bto.Merge(tranNum)
	})
}

// ApplyMerge applies the updates collected by meta.Merge
func (m *Meta) ApplyMerge(updates []update) {
	m.topInfo = m.topInfo.withUpdates(updates, btOver.WithMerged)
}

//-------------------------------------------------------------------

// Persist is called by state.Persist to write the state to the database.
// It collects the new fbtree roots which are then applied ApplyPersist.
func (m *Meta) Persist(flatten bool) []update {
	return m.topInfo.process(func(ov *btree.Overlay) *btree.Overlay {
		return ov.Save(flatten)
	})
}

// ApplyPersist takes the new fbtree roots from meta.Persist
// and updates the state with them.
func (m *Meta) ApplyPersist(updates []update) {
	m.topInfo = m.topInfo.withUpdates(updates, btOver.WithSaved)
}