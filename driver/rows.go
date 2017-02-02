/*
 * go-mysqlstack
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package driver

import (
	"github.com/XeLabs/go-mysqlstack/common"
	"github.com/XeLabs/go-mysqlstack/proto"
	"github.com/pkg/errors"

	querypb "github.com/XeLabs/go-mysqlstack/sqlparser/depends/query"
	"github.com/XeLabs/go-mysqlstack/sqlparser/depends/sqltypes"
)

type Rows interface {
	Next() bool
	Close() error
	Datas() []byte
	RowsAffected() uint64
	LastInsertID() uint64
	LastError() error
	Fields() []*querypb.Field
	RowValues() ([]sqltypes.Value, error)
}

type TextRows struct {
	c            Conn
	end          bool
	err          error
	payload      []byte
	rowsAffected uint64
	insertID     uint64
	buffer       *common.Buffer
	fields       []*querypb.Field
}

func NewTextRows(c Conn) *TextRows {
	return &TextRows{
		c:      c,
		buffer: common.NewBuffer(8),
	}
}

// http://dev.mysql.com/doc/internals/en/com-query-response.html#packet-ProtocolText::ResultsetRow
func (r *TextRows) Next() bool {
	defer func() {
		if r.err != nil {
			r.c.Cleanup()
		}
	}()

	if r.end {
		return false
	}

	// if fields count is 0
	// the packet is OK-Packet without Resultset
	if len(r.fields) == 0 {
		r.end = true
		return false
	}

	if r.payload, r.err = r.c.NextPacket(); r.err != nil {
		r.end = true
		return false
	}

	switch r.payload[0] {
	case proto.EOF_PACKET:
		r.end = true
		return false

	case proto.ERR_PACKET:
		e, ierr := proto.UnPackERR(r.payload)
		if ierr != nil {
			r.err = errors.Errorf("rows.next.error:%v", e.ErrorMessage)
		}
		r.end = true
		return false
	}
	r.buffer.Reset(r.payload)

	return true
}

// Close drain the rest packets and check the error
func (r *TextRows) Close() error {
	for r.Next() {
	}
	if err := r.LastError(); err != nil {
		return err
	}

	return nil
}

// https://dev.mysql.com/doc/internals/en/com-query-response.html#packet-ProtocolText::ResultsetRow
func (r *TextRows) RowValues() ([]sqltypes.Value, error) {
	if r.fields == nil {
		return nil, errors.New("rows.fields is NIL")
	}

	colNumber := len(r.fields)
	result := make([]sqltypes.Value, colNumber)
	for i := 0; i < colNumber; i++ {
		v, err := r.buffer.ReadLenEncodeBytes()
		if err != nil {
			r.c.Cleanup()
			return nil, err
		}
		// if v is NIL, it's a NULL column
		if v != nil {
			result[i] = sqltypes.MakeTrusted(r.fields[i].Type, v)
		}
	}

	return result, nil
}

func (r *TextRows) Datas() []byte {
	return r.buffer.Datas()
}

func (r *TextRows) Fields() []*querypb.Field {
	return r.fields
}

func (r *TextRows) RowsAffected() uint64 {
	return r.rowsAffected
}

func (r *TextRows) LastInsertID() uint64 {
	return r.insertID
}

func (r *TextRows) LastError() error {
	return r.err
}
