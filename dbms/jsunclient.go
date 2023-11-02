// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package dbms

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"slices"

	. "github.com/apmckinlay/gsuneido/core"
	"github.com/apmckinlay/gsuneido/core/trace"
	"github.com/apmckinlay/gsuneido/dbms/commands"
	"github.com/apmckinlay/gsuneido/dbms/csio"
	"github.com/apmckinlay/gsuneido/util/ascii"
	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/str"
)

// token is to authorize the next connection
var token string

// tokenLock guards token
var tokenLock sync.Mutex

// jsunClient is the client for the jSuneido server
type jsunClient struct {
	*csio.ReadWrite
	conn net.Conn
}

func NewJsunClient(conn net.Conn) *jsunClient {
	errfn := func(err string) {
		Fatal("client:", err)
	}
	rw := csio.NewReadWrite(conn, errfn)
	c := &jsunClient{ReadWrite: rw, conn: conn}
	tokenLock.Lock()
	defer tokenLock.Unlock()
	if token != "" {
		c.auth(token)
		token = c.Token()
	}
	return c
}

// Dbms interface

var _ IDbms = (*jsunClient)(nil)

func (dc *jsunClient) Admin(admin string, _ *Sviews) {
	dc.PutCmd(commands.Admin).PutStr(admin).Request()
}

func (dc *jsunClient) Auth(_ *Thread, s string) bool {
	if !dc.auth(s) {
		return false
	}
	tokenLock.Lock()
	defer tokenLock.Unlock()
	if token == "" {
		token = dc.Token()
	}
	return true
}

func (dc *jsunClient) auth(s string) bool {
	if s == "" {
		return false
	}
	dc.PutCmd(commands.Auth).PutStr(s).Request()
	return dc.GetBool()
}

func (dc *jsunClient) Check() string {
	dc.PutCmd(commands.Check).Request()
	return dc.GetStr()
}

func (dc *jsunClient) Close() {
	// On Windows, Close() is not sufficient for a graceful close.
	// If the client exits afterwards, the server gets WSAECONNRESET.
	// Tried delays, CloseWrite, SetLinger but nothing helps.
	// Currently, csio GetCmd specifically handles WSAECONNRESET.
	err := dc.conn.Close()
	if err != nil {
		log.Println("ERROR client close:", err)
	}
}

func (dc *jsunClient) Connections() Value {
	dc.PutCmd(commands.Connections).Request()
	ob := dc.GetVal().(*SuObject)
	ob.SetReadOnly()
	return ob
}

func (dc *jsunClient) Cursor(query string, _ *Sviews) ICursor {
	dc.PutCmd(commands.Cursor).PutStr(query).Request()
	cn := dc.GetInt()
	return newClientCursor(dc, cn)
}

func (dc *jsunClient) Cursors() int {
	dc.PutCmd(commands.Cursors).Request()
	return dc.GetInt()
}

func (dc *jsunClient) DisableTrigger(string) {
	panic("DoWithoutTriggers can't be used by a client")
}
func (dc *jsunClient) EnableTrigger(string) {
	assert.ShouldNotReachHere()
}

func (dc *jsunClient) Dump(table string) string {
	dc.PutCmd(commands.Dump).PutStr(table).Request()
	return dc.GetStr()
}

func (dc *jsunClient) Exec(_ *Thread, args Value) Value {
	packed := PackValue(args) // do this first because it could panic
	if trace.ClientServer.On() {
		if len(packed) < 100 {
			trace.ClientServer.Println("    ->", args)
		}
	}
	dc.PutCmd(commands.Exec)
	dc.PutStr_(packed).Request()
	return dc.ValueResult()
}

func (dc *jsunClient) Final() int {
	dc.PutCmd(commands.Final).Request()
	return dc.GetInt()
}

func (dc *jsunClient) Get(_ *Thread, query string, dir Dir,
	_ *Sviews) (Row, *Header, string) {
	return dc.get(0, query, dir)
}

func (dc *jsunClient) get(tn int, query string, dir Dir) (Row, *Header, string) {
	dc.PutCmd(commands.GetOne).PutByte(byte(dir)).PutInt(tn).PutStr(query).Request()
	if !dc.GetBool() {
		return nil, nil, ""
	}
	off := dc.GetInt()
	hdr := dc.getHdr()
	row := dc.getRow(off)
	return row, hdr, "updateable"
}

func (dc *jsunClient) Info() Value {
	dc.PutCmd(commands.Info).Request()
	return dc.GetVal()
}

func (dc *jsunClient) Kill(sessionid string) int {
	dc.PutCmd(commands.Kill).PutStr(sessionid).Request()
	return dc.GetInt()
}

func (dc *jsunClient) Load(table string) int {
	dc.PutCmd(commands.Load).PutStr(table).Request()
	return dc.GetInt()
}

func (dc *jsunClient) Log(s string) {
	dc.PutCmd(commands.Log).PutStr(s).Request()
}

func (dc *jsunClient) LibGet(name string) []string {
	dc.PutCmd(commands.LibGet).PutStr(name).Request()
	n := dc.GetSize()
	v := make([]string, 2*n)
	sizes := make([]int, n)
	for i := 0; i < 2*n; i += 2 {
		v[i] = dc.GetStr() // library
		sizes[i/2] = dc.GetSize()
	}
	for i := 1; i < 2*n; i += 2 {
		v[i] = dc.GetN(sizes[i/2]) // text
	}
	return v
}

func (dc *jsunClient) Libraries() []string {
	dc.PutCmd(commands.Libraries).Request()
	return dc.GetStrs()
}

func (dc *jsunClient) Nonce(*Thread) string {
	dc.PutCmd(commands.Nonce).Request()
	return dc.GetStr()
}

func (dc *jsunClient) Run(_ *Thread, code string) Value {
	dc.PutCmd(commands.Run).PutStr(code).Request()
	return dc.ValueResult()
}

func (dc *jsunClient) Schema(string) string {
	panic("Schema only available standalone")
}

func (dc *jsunClient) SessionId(th *Thread, id string) string {
	if s := th.Session(); s != "" && id == "" {
		return s // use cached value
	}
	dc.PutCmd(commands.SessionId).PutStr(id).Request()
	s := dc.GetStr()
	th.SetSession(s)
	return s
}

func (dc *jsunClient) Size() uint64 {
	dc.PutCmd(commands.Size).Request()
	return uint64(dc.GetInt64())
}

func (dc *jsunClient) Timestamp() SuDate {
	dc.PutCmd(commands.Timestamp).Request()
	return dc.GetVal().(SuDate)
}

func (dc *jsunClient) Token() string {
	dc.PutCmd(commands.Token).Request()
	return dc.GetStr()
}

func (dc *jsunClient) Transaction(update bool) ITran {
	dc.PutCmd(commands.Transaction).PutBool(update).Request()
	tn := dc.GetInt()
	return &TranClient{dc: dc, tn: tn}
}

func (dc *jsunClient) Transactions() *SuObject {
	dc.PutCmd(commands.Transactions).Request()
	ob := &SuObject{}
	for n := dc.GetInt(); n > 0; n-- {
		ob.Add(IntVal(dc.GetInt()))
	}
	return ob
}

func (dc *jsunClient) Unuse(lib string) bool {
	panic("can't Unuse('" + lib + "')\n" +
		"When client-server, only the server can Unuse")
}

func (dc *jsunClient) Use(lib string) bool {
	if slices.Contains(dc.Libraries(), lib) {
		return false
	}
	panic("can't Use('" + lib + "')\n" +
		"When client-server, only the server can Use")
}

func (dc *jsunClient) Unwrap() IDbms {
	return dc
}

func (dc *jsunClient) getHdr() *Header {
	n := dc.GetInt()
	fields := make([]string, 0, n)
	columns := make([]string, 0, n)
	for i := 0; i < n; i++ {
		s := dc.GetStr()
		if ascii.IsUpper(s[0]) {
			s = str.UnCapitalize(s)
		} else if !strings.HasSuffix(s, "_lower!") {
			fields = append(fields, s)
		}
		if s != "-" {
			columns = append(columns, s)
		}
	}
	return NewHeader([][]string{fields}, columns)
}

func (dc *jsunClient) getRow(off int) Row {
	return Row([]DbRec{{Record: dc.GetRec(), Off: uint64(off)}})
}

// ------------------------------------------------------------------

type TranClient struct {
	dc       *jsunClient
	conflict string
	tn       int
	ended    bool
}

var _ ITran = (*TranClient)(nil)

func (tc *TranClient) Abort() string {
	tc.ended = true
	tc.dc.PutCmd(commands.Abort).PutInt(tc.tn).Request()
	return ""
}

func (tc *TranClient) Asof(int64) int64 {
	return 0 // jSuneido doesn't support Asof
}

func (tc *TranClient) Complete() string {
	tc.ended = true
	tc.dc.PutCmd(commands.Commit).PutInt(tc.tn).Request()
	if tc.dc.GetBool() {
		return ""
	}
	tc.conflict = tc.dc.GetStr()
	return tc.conflict
}

func (tc *TranClient) Conflict() string {
	return tc.conflict
}

func (tc *TranClient) Ended() bool {
	return tc.ended
}

func (tc *TranClient) Delete(_ *Thread, _ string, off uint64) {
	tc.dc.PutCmd(commands.Erase).PutInt(tc.tn).PutInt(int(off)).Request()
}

func (tc *TranClient) Get(_ *Thread, query string, dir Dir,
	_ *Sviews) (Row, *Header, string) {
	return tc.dc.get(tc.tn, query, dir)
}

func (tc *TranClient) Query(query string, _ *Sviews) IQuery {
	tc.dc.PutCmd(commands.Query).PutInt(tc.tn).PutStr(query).Request()
	qn := tc.dc.GetInt()
	return newClientQuery(tc.dc, qn)
}

func (tc *TranClient) ReadCount() int {
	tc.dc.PutCmd(commands.ReadCount).PutInt(tc.tn).Request()
	return tc.dc.GetInt()
}

func (tc *TranClient) Action(_ *Thread, action string, _ *Sviews) int {
	tc.dc.PutCmd(commands.Action).PutInt(tc.tn).PutStr(action).Request()
	return tc.dc.GetInt()
}

func (tc *TranClient) Update(_ *Thread, _ string, off uint64, rec Record) uint64 {
	tc.dc.PutCmd(commands.Update).
		PutInt(tc.tn).PutInt(int(off)).PutRec(rec).Request()
	return uint64(tc.dc.GetInt())
}

func (tc *TranClient) WriteCount() int {
	tc.dc.PutCmd(commands.WriteCount).PutInt(tc.tn).Request()
	return tc.dc.GetInt()
}

func (tc *TranClient) String() string {
	return "Transaction" + strconv.Itoa(tc.tn)
}

// ------------------------------------------------------------------

// clientQueryCursor is the common stuff for clientQuery and clientCursor
type clientQueryCursor struct {
	dc   *jsunClient
	hdr  *Header
	keys []string // cache
	id   int
	qc   qcType
}

type qcType byte

const (
	query  qcType = 'q'
	cursor qcType = 'c'
)

func (qc *clientQueryCursor) Close() {
	qc.dc.PutCmd(commands.Close).PutInt(qc.id).PutByte(byte(qc.qc)).Request()
}

func (qc *clientQueryCursor) Header() *Header {
	if qc.hdr == nil { // cached
		qc.dc.PutCmd(commands.Header).PutInt(qc.id).PutByte(byte(qc.qc)).Request()
		qc.hdr = qc.dc.getHdr()
	}
	return qc.hdr
}

func (qc *clientQueryCursor) Keys() []string {
	if qc.keys == nil { // cached
		qc.dc.PutCmd(commands.Keys).PutInt(qc.id).PutByte(byte(qc.qc)).Request()
		nk := qc.dc.GetInt()
		qc.keys = make([]string, nk)
		for i := range qc.keys {
			cb := str.CommaBuilder{}
			n := qc.dc.GetInt()
			for ; n > 0; n-- {
				cb.Add(qc.dc.GetStr())
			}
			qc.keys[i] = cb.String()
		}
	}
	return qc.keys
}

func (qc *clientQueryCursor) Order() []string {
	qc.dc.PutCmd(commands.Order).PutInt(qc.id).PutByte(byte(qc.qc)).Request()
	return qc.dc.GetStrs()
}

func (qc *clientQueryCursor) Rewind() {
	qc.dc.PutCmd(commands.Rewind).PutInt(qc.id).PutByte(byte(qc.qc)).Request()
}

func (qc *clientQueryCursor) Strategy(_ bool) string {
	qc.dc.PutCmd(commands.Strategy).PutInt(qc.id).PutByte(byte(qc.qc)).Request()
	return qc.dc.GetStr()
}

// clientQuery implements IQuery ------------------------------------
type clientQuery struct {
	clientQueryCursor
}

func newClientQuery(dc *jsunClient, qn int) *clientQuery {
	return &clientQuery{clientQueryCursor{dc: dc, id: qn, qc: query}}
}

var _ IQuery = (*clientQuery)(nil)

func (q *clientQuery) Get(_ *Thread, dir Dir) (Row, string) {
	q.dc.PutCmd(commands.Get).PutByte(byte(dir)).PutInt(0).PutInt(q.id).Request()
	if !q.dc.GetBool() {
		return nil, ""
	}
	off := q.dc.GetInt()
	row := q.dc.getRow(off)
	return row, "updateable"
}

func (q *clientQuery) Output(_ *Thread, rec Record) {
	q.dc.PutCmd(commands.Output).PutInt(q.id).PutRec(rec).Request()
}

// clientCursor implements IQuery ------------------------------------
type clientCursor struct {
	clientQueryCursor
}

func newClientCursor(dc *jsunClient, cn int) *clientCursor {
	return &clientCursor{clientQueryCursor{dc: dc, id: cn, qc: cursor}}
}

var _ ICursor = (*clientCursor)(nil)

func (q *clientCursor) Get(_ *Thread, tran ITran, dir Dir) (Row, string) {
	t := tran.(*TranClient)
	q.dc.PutCmd(commands.Get).PutByte(byte(dir)).PutInt(t.tn).PutInt(q.id).Request()
	if !q.dc.GetBool() {
		return nil, ""
	}
	off := q.dc.GetInt()
	row := q.dc.getRow(off)
	return row, "updateable"
}
