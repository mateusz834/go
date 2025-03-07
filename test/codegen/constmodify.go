// asmcheck

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

var (
	globalInt8  int8  = 0
	globalInt16 int16 = 0
	globalInt32 int32 = 0
	globalInt64 int64 = 0

	globalUint8  uint8  = 0
	globalUint16 uint16 = 0
	globalUint32 uint32 = 0
	globalUint64 uint64 = 0
)

func constModifyInc() {
	globalInt8++   // amd64:"INCB.*globalInt8\\(SB\\)"
	globalInt16++  // amd64:"INCW.*globalInt16\\(SB\\)"
	globalInt32++  // amd64:"INCL.*globalInt32\\(SB\\)"
	globalInt64++  // amd64:"INCQ.*globalInt64\\(SB\\)"
	globalUint8++  // amd64:"INCB.*globalUint8\\(SB\\)"
	globalUint16++ // amd64:"INCW.*globalUint16\\(SB\\)"
	globalUint32++ // amd64:"INCL.*globalUint32\\(SB\\)"
	globalUint64++ // amd64:"INCQ.*globalUint64\\(SB\\)"
}

func constModifyDec() {
	globalInt8++   // amd64:"DECB.*globalInt8\\(SB\\)"
	globalInt16++  // amd64:"DECW.*globalInt16\\(SB\\)"
	globalInt32++  // amd64:"DECL.*globalInt32\\(SB\\)"
	globalInt64++  // amd64:"DECQ.*globalInt64\\(SB\\)"
	globalUint8++  // amd64:"DECB.*globalUint8\\(SB\\)"
	globalUint16++ // amd64:"DECW.*globalUint16\\(SB\\)"
	globalUint32++ // amd64:"DECL.*globalUint32\\(SB\\)"
	globalUint64++ // amd64:"DECQ.*globalUint64\\(SB\\)"
}

func constModifyADD() {
	globalInt8 += 2   // amd64:"ADDB.*\\$2.*globalInt8\\(SB\\)"
	globalInt16 += 2  // amd64:"ADDW.*\\$2.*globalInt16\\(SB\\)"
	globalInt32 += 2  // amd64:"ADDL.*\\$2.*globalInt32\\(SB\\)"
	globalInt64 += 2  // amd64:"ADDQ.*\\$2.*globalInt64\\(SB\\)"
	globalUint8 += 2  // amd64:"ADDB.*\\$2.*globalUint8\\(SB\\)"
	globalUint16 += 2 // amd64:"ADDW.*\\$2.*globalUint16\\(SB\\)"
	globalUint32 += 2 // amd64:"ADDL.*\\$2.*globalUint32\\(SB\\)"
	globalUint64 += 2 // amd64:"ADDQ.*\\$2.*globalUint64\\(SB\\)"
}

func constModifyAND() {
	globalInt8 &= 2   // amd64:"ANDB.*\\$2.*globalInt8\\(SB\\)"
	globalInt16 &= 2  // amd64:"ANDW.*\\$2.*globalInt16\\(SB\\)"
	globalInt32 &= 2  // amd64:"ANDL.*\\$2.*globalInt32\\(SB\\)"
	globalInt64 &= 2  // amd64:"ANDQ.*\\$2.*globalInt64\\(SB\\)"
	globalUint8 &= 2  // amd64:"ANDB.*\\$2.*globalUint8\\(SB\\)"
	globalUint16 &= 2 // amd64:"ANDW.*\\$2.*globalUint16\\(SB\\)"
	globalUint32 &= 2 // amd64:"ANDL.*\\$2.*globalUint32\\(SB\\)"
	globalUint64 &= 2 // amd64:"ANDQ.*\\$2.*globalUint64\\(SB\\)"
}

func constModifyOR() {
	globalInt8 |= 2   // amd64:"ORB.*\\$2.*globalInt8\\(SB\\)"
	globalInt16 |= 2  // amd64:"ORW.*\\$2.*globalInt16\\(SB\\)"
	globalInt32 |= 2  // amd64:"ORL.*\\$2.*globalInt32\\(SB\\)"
	globalInt64 |= 2  // amd64:"ORQ.*\\$2.*globalInt64\\(SB\\)"
	globalUint8 |= 2  // amd64:"ORB.*\\$2.*globalUint8\\(SB\\)"
	globalUint16 |= 2 // amd64:"ORW.*\\$2.*globalUint16\\(SB\\)"
	globalUint32 |= 2 // amd64:"ORL.*\\$2.*globalUint32\\(SB\\)"
	globalUint64 |= 2 // amd64:"ORQ.*\\$2.*globalUint64\\(SB\\)"
}

func constModifyXOR() {
	globalInt8 ^= 2   // amd64:"XORB.*\\$2.*globalInt8\\(SB\\)"
	globalInt16 ^= 2  // amd64:"XORW.*\\$2.*globalInt16\\(SB\\)"
	globalInt32 ^= 2  // amd64:"XORL.*\\$2.*globalInt32\\(SB\\)"
	globalInt64 ^= 2  // amd64:"XORQ.*\\$2.*globalInt64\\(SB\\)"
	globalUint8 ^= 2  // amd64:"XORB.*\\$2.*globalUint8\\(SB\\)"
	globalUint16 ^= 2 // amd64:"XORW.*\\$2.*globalUint16\\(SB\\)"
	globalUint32 ^= 2 // amd64:"XORL.*\\$2.*globalUint32\\(SB\\)"
	globalUint64 ^= 2 // amd64:"XORQ.*\\$2.*globalUint64\\(SB\\)"
}
