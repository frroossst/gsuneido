// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package commands

//go:generate stringer -type=Command

type Command byte

// command values must match jSuneido
const (
	Abort Command = iota
	Admin
	Auth
	Check
	Close
	Commit
	Connections
	Cursor
	Cursors
	Erase
	Exec
	Strategy
	Final
	Get
	GetOne
	Header
	Info
	Keys
	Kill
	LibGet
	Libraries
	Log
	Nonce
	Order
	Output
	Query
	ReadCount
	Action
	Rewind
	Run
	SessionId
	Size
	Timestamp
	Token
	Transaction
	Transactions
	Update
	WriteCount
	EndSession
	Asof
)
