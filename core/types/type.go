// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package types

//go:generate stringer -type=Type

type Type int

// must match Ord up to Object
const (
	Boolean Type = iota
	Number
	String
	Date
	Object
	Record
	Function
	Block
	BuiltinFunction
	Class
	Method
	Except
	Instance
	Iterator
	Transaction
	Query
	Cursor
	File
	AstNode
	BuiltinClass
	LruCache
	N
)
