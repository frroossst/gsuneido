// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import (
	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/core/types"
	"github.com/apmckinlay/gsuneido/util/dnum"
)

func NewSuQueryNode(q Query) Value {
	return SuQueryNode{q: q}
}

type SuQueryNode struct {
	ValueBase[SuQueryNode]
	q Query
}

func (SuQueryNode) Type() types.Type {
	return types.QueryNode
}

func (SuQueryNode) Equal(any) bool {
	return false
}

func (SuQueryNode) SetConcurrent() {
	// read-only so nothing to do
}

func (n SuQueryNode) Get(_ *Thread, key Value) Value {
	return n.q.ValueGet(key)
}

func qryBase(q Query, key Value) Value {
	switch key {
	case SuStr("nrows"):
		n, _ := q.Nrows()
		return IntVal(n)
	case SuStr("pop"):
		_, p := q.Nrows()
		return IntVal(p)
	case SuStr("fast1"):
		return SuBool(q.fastSingle())
	case SuStr("nchild"):
		return Zero // overridden by Query1 and Query2
	}
	return qryCost(q, key)
}

type costable interface {
	cacheCost() (float64, Cost, Cost)
}

func qryCost(q costable, key Value) Value {
	switch key {
	case SuStr("frac"):
		frac, _, _ := q.cacheCost()
		return SuDnum{Dnum: dnum.FromFloat(frac)}
	case SuStr("fixcost"):
		_, fixcost, _ := q.cacheCost()
		return IntVal(fixcost)
	case SuStr("varcost"):
		_, _, varcost := q.cacheCost()
		return IntVal(varcost)
	}
	return nil
}

func (q *Table) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("table")
	case SuStr("name"), SuStr("string"):
		return SuStr(q.name)
	case SuStr("strategy"):
		return SuStr(q.String())
	}
	return qryBase(q, key)
}

func (q *Tables) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("table")
	case SuStr("name"), SuStr("string"), SuStr("strategy"):
		return SuStr("tables")
	}
	return qryCost(q, key)
}

func (q *TablesLookup) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("table")
	case SuStr("name"):
		return SuStr(q.table)
	case SuStr("string"), SuStr("strategy"):
		return SuStr(q.String())
	}
	return qryCost(q, key)
}

func (q *Columns) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("table")
	case SuStr("name"), SuStr("string"), SuStr("strategy"):
		return SuStr("columns")
	}
	return qryCost(q, key)
}

func (q *Indexes) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("table")
	case SuStr("name"), SuStr("string"), SuStr("strategy"):
		return SuStr("indexes")
	}
	return qryCost(q, key)
}

func (q *Views) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("table")
	case SuStr("name"), SuStr("string"), SuStr("strategy"):
		return SuStr("views")
	}
	return qryCost(q, key)
}

func (q *History) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("table")
	case SuStr("name"), SuStr("string"), SuStr("strategy"):
		return SuStr("history")
	}
	return qryCost(q, key)
}

func (q *Nothing) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("nothing")
	case SuStr("string"), SuStr("strategy"):
		return SuStr("nothing(" + q.table + ")")
	}
	return qryCost(q, key)
}

func (q *ProjectNone) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"), SuStr("string"), SuStr("strategy"):
		return SuStr("projectNone")
	}
	return qryCost(q, key)
}

//-------------------------------------------------------------------

func query1(q Query, key Value) Value {
	switch key {
	case SuStr("source"):
		return NewSuQueryNode(q.(q1i).Source())
	case SuStr("nchild"):
		return One
	}
	return qryBase(q, key)
}

func (q *Extend) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("extend")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

func (q *Project) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("project")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

func (q *Rename) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("rename")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

func (q *Sort) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("sort")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

func (q *Summarize) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("summarize")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

func (q *TempIndex) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("tempindex")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

func (q *Where) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("rename")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

func (q *View) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("view")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query1(q, key)
}

//-------------------------------------------------------------------

var Two = IntVal(2)

func query2(q Query, key Value) Value {
	switch key {
	case SuStr("source1"):
		return NewSuQueryNode(q.(q2i).Source())
	case SuStr("source2"):
		return NewSuQueryNode(q.(q2i).Source2())
	case SuStr("nchild"):
		return Two
	}
	return qryBase(q, key)
}

func (q *Union) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("union")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query2(q, key)
}

func (q *Intersect) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("intersect")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query2(q, key)
}

func (q *Minus) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("minus")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query2(q, key)
}

func (q *Times) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("times")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query2(q, key)
}

func (q *Join) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("join")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query2(q, key)
}

func (q *LeftJoin) ValueGet(key Value) Value {
	switch key {
	case SuStr("type"):
		return SuStr("leftjoin")
	case SuStr("string"):
		return SuStr(format1(q))
	case SuStr("strategy"):
		return SuStr(q.stringOp())
	}
	return query2(q, key)
}
