// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import (
	"log"
	"slices"

	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/db19/index/ixkey"
	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/sortlist"
	"github.com/apmckinlay/gsuneido/util/str"
	"github.com/apmckinlay/gsuneido/util/tsc"
)

// TempIndex is inserted by SetApproach as required.
// It builds a sortlist of either DbRec or Row.
// Keys are not constructed for the index or Lookup/Select
// so there are no size limits.
type TempIndex struct {
	tran   QueryTran
	iter   rowIter
	st     *SuTran
	th     *Thread
	order  []string
	selOrg []string
	selEnd []string
	Query1
	rewound bool
}

var selMin []string
var selMax = []string{ixkey.Max}

const tempindexWarn = 200_000 // ???
const derivedWarn = 8_000_000 // ??? // derivedWarn is also used by Project

func NewTempIndex(src Query, order []string, tran QueryTran) *TempIndex {
	order = withoutFixed(order, src.Fixed())
	if len(order) == 0 {
		log.Println("ERROR: empty TempIndex")
	}
	ti := TempIndex{order: order, tran: tran, selOrg: selMin, selEnd: selMax}
	ti.source = src
	ti.header = src.Header().Dup() // dup because sortlist is concurrent
	ti.keys = src.Keys()
	ti.fixed = src.Fixed()
	ti.setNrows(src.Nrows())
	ti.rowSiz.Set(src.rowSize())
	ti.singleTbl.Set(src.SingleTable())
	ti.fast1.Set(src.fastSingle())
	return &ti
}

func (ti *TempIndex) String() string {
	return "tempindex" + str.Join("(,)", ti.order)
}

func (*TempIndex) Indexes() [][]string {
	panic(assert.ShouldNotReachHere())
}

func (ti *TempIndex) Transform() Query {
	return ti
}

func (*TempIndex) setApproach([]string, float64, any, QueryTran) {
	assert.ShouldNotReachHere()
}

func (*TempIndex) Simple(*Thread) []Row {
	panic(assert.ShouldNotReachHere())
}

// execution --------------------------------------------------------

func (ti *TempIndex) Rewind() {
	if ti.iter != nil {
		ti.iter.Rewind()
	}
	ti.rewound = true
}

func (ti *TempIndex) Select(cols, vals []string) {
	// similar to Where Select
	ti.nsels++
	ti.Rewind()
	if cols == nil && vals == nil { // clear select
		ti.selOrg, ti.selEnd = selMin, selMax
		return
	}
	satisfied, conflict := selectFixed(cols, vals, ti.source.Fixed())
	if conflict {
		ti.selOrg, ti.selEnd = selMax, selMin
		return
	}
	if satisfied {
		ti.selOrg, ti.selEnd = selMin, selMax
		return
	}
	ti.selOrg = ti.makeKey(cols, vals, false)
	ti.selEnd = append(ti.selOrg, ixkey.Max)
}

func (ti *TempIndex) Lookup(th *Thread, cols, vals []string) Row {
	ti.nlooks++
	if conflictFixed(cols, vals, ti.source.Fixed()) {
		return nil
	}
	ti.th = th
	defer func() { ti.th = nil }()
	if ti.iter == nil {
		ti.iter = ti.makeIndex()
	}
	key := ti.makeKey(cols, vals, true)
	row := ti.iter.Seek(key)
	if row == nil || !ti.matches(row, key) {
		return nil
	}
	return row
}

func (ti *TempIndex) makeKey(cols, vals []string, full bool) []string {
	key := make([]string, 0, len(ti.order))
	for _, col := range ti.order {
		j := slices.Index(cols, col)
		if j == -1 {
			if full {
				panic("TempIndex makeKey not full")
			}
			break
		}
		key = append(key, vals[j])
	}
	return key
}

func (ti *TempIndex) matches(row Row, key []string) bool {
	for i, col := range ti.order {
		x := row.GetRawVal(ti.header, col, nil, ti.st)
		y := key[i]
		if x != y {
			return false
		}
	}
	return true
}

func (ti *TempIndex) Get(th *Thread, dir Dir) Row {
	defer func(t uint64) { ti.tget += tsc.Read() - t }(tsc.Read())
	ti.th = th
	defer func() { ti.th = nil }()
	if ti.iter == nil {
		ti.iter = ti.makeIndex()
		ti.rewound = true
	}
	if ti.conflict() {
		return nil
	}
	var row Row
	if ti.rewound {
		if dir == Next {
			row = ti.iter.Seek(ti.selOrg)
		} else { // Prev
			row = ti.iter.Seek(ti.selEnd)
			if row == nil {
				ti.iter.Rewind()
			}
			row = ti.iter.Get(dir)
		}
		ti.rewound = false
	} else {
		row = ti.iter.Get(dir)
	}
	if row == nil || !ti.selected(row) {
		return nil
	}
	ti.ngets++
	return row
}

type rowIter interface {
	Get(Dir) Row
	Rewind()
	Seek(key []string) Row
}

func (ti *TempIndex) makeIndex() rowIter {
	ti.st = MakeSuTran(ti.tran)
	// need to copy header to avoid data race from concurrent sortlist
	ti.header = ti.source.Header().Dup()
	if ti.source.SingleTable() {
		return ti.single()
	}
	return ti.multi()
}

func (ti *TempIndex) selected(row Row) bool {
	if ti.satisfied() {
		return true
	}
	for i, sel := range ti.selOrg {
		col := ti.order[i]
		x := row.GetRawVal(ti.header, col, ti.th, ti.st)
		if x != sel {
			return false
		}
	}
	return true
}

func (ti *TempIndex) satisfied() bool {
	return len(ti.selOrg) == 1 && ti.selOrg[0] == ixkey.Min &&
		len(ti.selEnd) == 1 && ti.selEnd[0] == ixkey.Max
}

func (ti *TempIndex) conflict() bool {
	return len(ti.selOrg) == 1 && ti.selOrg[0] == ixkey.Max &&
		len(ti.selEnd) == 1 && ti.selEnd[0] == ixkey.Min
}

//-------------------------------------------------------------------
// single is for ti.source.SingleTable.
// Since there is only one table,
// we can store the single record directly in the sortlist.

type singleIter struct {
	iter *sortlist.Iter[DbRec]
}

func (ti *TempIndex) single() rowIter {
	var th2 Thread // separate thread because sortlist runs in the background
	b := sortlist.NewSorting(
		func(x DbRec) bool { return x == DbRec{} },
		func(x, y DbRec) bool { return ti.less(&th2, Row{x}, Row{y}) })
	nrows := 0
	warned := false
	for {
		row := ti.source.Get(ti.th, Next)
		if row == nil {
			break
		}
		b.Add(row[0])
		nrows++
		if nrows > tempindexWarn && !warned {
			Warning("temp index large >", tempindexWarn)
			warned = true
		}
	}
	if nrows > 2*tempindexWarn {
		log.Println("temp index large =", nrows)
	}
	// NOTE: the closure captures ti not ti.th
	lt := func(rec DbRec, key []string) bool {
		return ti.less2(ti.th, Row{rec}, key)
	}
	return singleIter{b.Finish().Iter(lt)}
}

func (ti *TempIndex) less(th *Thread, xrow, yrow Row) bool {
	for _, col := range ti.order {
		x := xrow.GetRawVal(ti.header, col, th, ti.st)
		y := yrow.GetRawVal(ti.header, col, th, ti.st)
		if x != y {
			return x < y
		}
	}
	return false
}

// less2 is used for Seek
func (ti *TempIndex) less2(th *Thread, row Row, key []string) bool {
	n := max(len(ti.order), len(key))
	for i := range n {
		if i >= len(key) {
			return false
		}
		if i >= len(ti.order) {
			return true
		}
		x := row.GetRawVal(ti.header, ti.order[i], th, ti.st)
		y := key[i]
		if x != y {
			return x < y
		}
	}
	return false
}

func (it singleIter) Seek(key []string) Row {
	it.iter.Seek(key)
	return it.get()
}

func (it singleIter) Rewind() {
	it.iter.Rewind()
}

func (it singleIter) Get(dir Dir) Row {
	if dir == Next {
		it.iter.Next()
	} else {
		it.iter.Prev()
	}
	return it.get()
}

func (it singleIter) get() Row {
	if it.iter.Eof() {
		return nil
	}
	return Row{it.iter.Cur()}
}

//-------------------------------------------------------------------
// multi is when we have multiple records (not ti.source.SingleTable).
// So we store the Row's in the sortlist.

type multiIter struct {
	iter *sortlist.Iter[Row]
}

func (ti *TempIndex) multi() rowIter {
	var th2 Thread // separate thread because sortlist runs in the background
	b := sortlist.NewSorting(
		func(row Row) bool { return row == nil },
		func(xrow, yrow Row) bool {
			return ti.less(&th2, xrow, yrow)
		})
	nrows := 0
	warned := false
	derived := 0
	derivedWarned := false
	for {
		row := ti.source.Get(ti.th, Next)
		if row == nil {
			break
		}
		nrows++
		if nrows > tempindexWarn && !warned {
			Warning("temp index large >", tempindexWarn)
			warned = true
		}
		derived += row.Derived()
		if derived > derivedWarn && !derivedWarned {
			Warning("temp index derived large >", derivedWarn,
				"average", derived/nrows)
			derivedWarned = true
		}
		b.Add(row)
	}
	if nrows > 2*tempindexWarn {
		log.Println("temp index large =", nrows)
	}
	if derived > 2*derivedWarn {
		log.Println("temp index derived large =",
			derived, "average", derived/nrows)
	}
	// NOTE: the closure captures ti not ti.th
	lt := func(row Row, key []string) bool {
		return ti.less2(ti.th, row, key)
	}
	return multiIter{b.Finish().Iter(lt)}
}

func (it multiIter) Seek(key []string) Row {
	it.iter.Seek(key)
	return it.get()
}

func (it multiIter) Rewind() {
	it.iter.Rewind()
}

func (it multiIter) Get(dir Dir) Row {
	if dir == Next {
		it.iter.Next()
	} else {
		it.iter.Prev()
	}
	return it.get()
}

func (it multiIter) get() Row {
	if it.iter.Eof() {
		return nil
	}
	return it.iter.Cur()
}
