/*
 * go-mysqlstack
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package proto

import (
	"github.com/XeLabs/go-mysqlstack/common"
	"github.com/pkg/errors"
)

const (
	OK_PACKET byte = 0x00
)

type OK struct {
	Header       byte // 0x00
	AffectedRows uint64
	LastInsertID uint64
	StatusFlags  uint16
	Warnings     uint16
}

// https://dev.mysql.com/doc/internals/en/packet-OK_Packet.html
func UnPackOK(data []byte) (o *OK, err error) {
	o = &OK{}
	buf := common.ReadBuffer(data)

	// header
	if o.Header, err = buf.ReadU8(); err != nil {
		return
	}
	if o.Header != OK_PACKET {
		err = errors.Errorf("packet.header[%v]!=OK_PACKET", o.Header)
		return
	}

	// AffectedRows
	if o.AffectedRows, err = buf.ReadLenEncode(); err != nil {
		return
	}

	// LastInsertID
	if o.LastInsertID, err = buf.ReadLenEncode(); err != nil {
		return
	}

	// Status
	if o.StatusFlags, err = buf.ReadU16(); err != nil {
		return
	}

	// Warnings
	if o.Warnings, err = buf.ReadU16(); err != nil {
		return
	}

	return
}

func PackOK(o *OK) []byte {
	buf := common.NewBuffer(64)

	// OK
	buf.WriteU8(OK_PACKET)

	// affected rows
	buf.WriteLenEncode(o.AffectedRows)

	// last insert id
	buf.WriteLenEncode(o.LastInsertID)

	// status
	buf.WriteU16(o.StatusFlags)

	// warnings
	buf.WriteU16(o.Warnings)

	return buf.Datas()
}
