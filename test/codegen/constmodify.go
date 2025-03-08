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

var (
	globalInt8Ptr  = &globalInt8
	globalInt16Ptr = &globalInt16
	globalInt32Ptr = &globalInt32
	globalInt64Ptr = &globalInt64

	globalUint8Ptr  = &globalUint8
	globalUint16Ptr = &globalUint16
	globalUint32Ptr = &globalUint32
	globalUint64Ptr = &globalUint64
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

	(*globalInt8Ptr)++   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "INCB.*(AX)"
	(*globalInt16Ptr)++  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "INCW.*(AX)"
	(*globalInt32Ptr)++  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "INCL.*(AX)"
	(*globalInt64Ptr)++  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "INCQ.*(AX)"
	(*globalUint8Ptr)++  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "INCB.*(AX)"
	(*globalUint16Ptr)++ // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "INCW.*(AX)"
	(*globalUint32Ptr)++ // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "INCL.*(AX)"
	(*globalUint64Ptr)++ // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "INCQ.*(AX)"
}

func constModifyDec() {
	globalInt8--   // amd64:"DECB.*globalInt8\\(SB\\)"
	globalInt16--  // amd64:"DECW.*globalInt16\\(SB\\)"
	globalInt32--  // amd64:"DECL.*globalInt32\\(SB\\)"
	globalInt64--  // amd64:"DECQ.*globalInt64\\(SB\\)"
	globalUint8--  // amd64:"DECB.*globalUint8\\(SB\\)"
	globalUint16-- // amd64:"DECW.*globalUint16\\(SB\\)"
	globalUint32-- // amd64:"DECL.*globalUint32\\(SB\\)"
	globalUint64-- // amd64:"DECQ.*globalUint64\\(SB\\)"

	(*globalInt8Ptr)--   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "DECB.*(AX)"
	(*globalInt16Ptr)--  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "DECW.*(AX)"
	(*globalInt32Ptr)--  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "DECL.*(AX)"
	(*globalInt64Ptr)--  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "DECQ.*(AX)"
	(*globalUint8Ptr)--  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "DECB.*(AX)"
	(*globalUint16Ptr)-- // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "DECW.*(AX)"
	(*globalUint32Ptr)-- // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "DECL.*(AX)"
	(*globalUint64Ptr)-- // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "DECQ.*(AX)"
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

	(*globalInt8Ptr) += 2   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "ADDB.*\\$2.*(AX)"
	(*globalInt16Ptr) += 2  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "ADDW.*\\$2.*(AX)"
	(*globalInt32Ptr) += 2  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "ADDL.*\\$2.*(AX)"
	(*globalInt64Ptr) += 2  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "ADDQ.*\\$2.*(AX)"
	(*globalUint8Ptr) += 2  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "ADDB.*\\$2.*(AX)"
	(*globalUint16Ptr) += 2 // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "ADDW.*\\$2.*(AX)"
	(*globalUint32Ptr) += 2 // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "ADDL.*\\$2.*(AX)"
	(*globalUint64Ptr) += 2 // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "ADDQ.*\\$2.*(AX)"
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

	(*globalInt8Ptr) &= 2   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "ANDB.*\\$2.*(AX)"
	(*globalInt16Ptr) &= 2  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "ANDW.*\\$2.*(AX)"
	(*globalInt32Ptr) &= 2  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "ANDL.*\\$2.*(AX)"
	(*globalInt64Ptr) &= 2  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "ANDQ.*\\$2.*(AX)"
	(*globalUint8Ptr) &= 2  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "ANDB.*\\$2.*(AX)"
	(*globalUint16Ptr) &= 2 // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "ANDW.*\\$2.*(AX)"
	(*globalUint32Ptr) &= 2 // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "ANDL.*\\$2.*(AX)"
	(*globalUint64Ptr) &= 2 // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "ANDQ.*\\$2.*(AX)"
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

	(*globalInt8Ptr) |= 2   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "ORB.*\\$2.*(AX)"
	(*globalInt16Ptr) |= 2  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "ORW.*\\$2.*(AX)"
	(*globalInt32Ptr) |= 2  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "ORL.*\\$2.*(AX)"
	(*globalInt64Ptr) |= 2  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "ORQ.*\\$2.*(AX)"
	(*globalUint8Ptr) |= 2  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "ORB.*\\$2.*(AX)"
	(*globalUint16Ptr) |= 2 // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "ORW.*\\$2.*(AX)"
	(*globalUint32Ptr) |= 2 // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "ORL.*\\$2.*(AX)"
	(*globalUint64Ptr) |= 2 // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "ORQ.*\\$2.*(AX)"
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

	(*globalInt8Ptr) ^= 2   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "XORB.*\\$2.*(AX)"
	(*globalInt16Ptr) ^= 2  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "XORW.*\\$2.*(AX)"
	(*globalInt32Ptr) ^= 2  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "XORL.*\\$2.*(AX)"
	(*globalInt64Ptr) ^= 2  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "XORQ.*\\$2.*(AX)"
	(*globalUint8Ptr) ^= 2  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "XORB.*\\$2.*(AX)"
	(*globalUint16Ptr) ^= 2 // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "XORW.*\\$2.*(AX)"
	(*globalUint32Ptr) ^= 2 // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "XORL.*\\$2.*(AX)"
	(*globalUint64Ptr) ^= 2 // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "XORQ.*\\$2.*(AX)"
}

func constModifySHL() {
	globalInt8 <<= 2   // amd64:"SHLB.*\\$2.*globalInt8\\(SB\\)"
	globalInt16 <<= 2  // amd64:"SHLW.*\\$2.*globalInt16\\(SB\\)"
	globalInt32 <<= 2  // amd64:"SHLL.*\\$2.*globalInt32\\(SB\\)"
	globalInt64 <<= 2  // amd64:"SHLQ.*\\$2.*globalInt64\\(SB\\)"
	globalUint8 <<= 2  // amd64:"SHLB.*\\$2.*globalUint8\\(SB\\)"
	globalUint16 <<= 2 // amd64:"SHLW.*\\$2.*globalUint16\\(SB\\)"
	globalUint32 <<= 2 // amd64:"SHLL.*\\$2.*globalUint32\\(SB\\)"
	globalUint64 <<= 2 // amd64:"SHLQ.*\\$2.*globalUint64\\(SB\\)"

	(*globalInt8Ptr) <<= 2   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "SHLB.*\\$2.*(AX)"
	(*globalInt16Ptr) <<= 2  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "SHLW.*\\$2.*(AX)"
	(*globalInt32Ptr) <<= 2  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "SHLL.*\\$2.*(AX)"
	(*globalInt64Ptr) <<= 2  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "SHLQ.*\\$2.*(AX)"
	(*globalUint8Ptr) <<= 2  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "SHLB.*\\$2.*(AX)"
	(*globalUint16Ptr) <<= 2 // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "SHLW.*\\$2.*(AX)"
	(*globalUint32Ptr) <<= 2 // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "SHLL.*\\$2.*(AX)"
	(*globalUint64Ptr) <<= 2 // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "SHLQ.*\\$2.*(AX)"
}

func constModifySHR() {
	globalInt8 >>= 2   // amd64:"SARB.*\\$2.*globalInt8\\(SB\\)"
	globalInt16 >>= 2  // amd64:"SARW.*\\$2.*globalInt16\\(SB\\)"
	globalInt32 >>= 2  // amd64:"SARL.*\\$2.*globalInt32\\(SB\\)"
	globalInt64 >>= 2  // amd64:"SARQ.*\\$2.*globalInt64\\(SB\\)"
	globalUint8 >>= 2  // amd64:"SHRB.*\\$2.*globalUint8\\(SB\\)"
	globalUint16 >>= 2 // amd64:"SHRW.*\\$2.*globalUint16\\(SB\\)"
	globalUint32 >>= 2 // amd64:"SHRL.*\\$2.*globalUint32\\(SB\\)"
	globalUint64 >>= 2 // amd64:"SHRQ.*\\$2.*globalUint64\\(SB\\)"

	(*globalInt8Ptr) >>= 2   // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "SARB.*\\$2.*(AX)"
	(*globalInt16Ptr) >>= 2  // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "SARW.*\\$2.*(AX)"
	(*globalInt32Ptr) >>= 2  // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "SARL.*\\$2.*(AX)"
	(*globalInt64Ptr) >>= 2  // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "SARQ.*\\$2.*(AX)"
	(*globalUint8Ptr) >>= 2  // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "SHRB.*\\$2.*(AX)"
	(*globalUint16Ptr) >>= 2 // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "SHRW.*\\$2.*(AX)"
	(*globalUint32Ptr) >>= 2 // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "SHRL.*\\$2.*(AX)"
	(*globalUint64Ptr) >>= 2 // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "SHRQ.*\\$2.*(AX)"
}

// TODO: check rules check 0-A is NEG A

func constModifyNOT() {
	globalInt8 = ^globalInt8     // amd64:"NOTB.*globalInt8\\(SB\\)"
	globalInt16 = ^globalInt16   // amd64:"NOTW.*globalInt16\\(SB\\)"
	globalInt32 = ^globalInt32   // amd64:"NOTL.*globalInt32\\(SB\\)"
	globalInt64 = ^globalInt64   // amd64:"NOTQ.*globalInt64\\(SB\\)"
	globalUint8 = ^globalUint8   // amd64:"NOTB.*globalUint8\\(SB\\)"
	globalUint16 = ^globalUint16 // amd64:"NOTW.*globalUint16\\(SB\\)"
	globalUint32 = ^globalUint32 // amd64:"NOTL.*globalUint32\\(SB\\)"
	globalUint64 = ^globalUint64 // amd64:"NOTQ.*globalUint64\\(SB\\)"

	(*globalInt8Ptr) = ^(*globalInt8Ptr)     // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "NOTB.*(AX)"
	(*globalInt16Ptr) = ^(*globalInt16Ptr)   // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "NOTW.*(AX)"
	(*globalInt32Ptr) = ^(*globalInt32Ptr)   // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "NOTL.*(AX)"
	(*globalInt64Ptr) = ^(*globalInt64Ptr)   // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "NOTQ.*(AX)"
	(*globalUint8Ptr) = ^(*globalUint8Ptr)   // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "NOTB.*(AX)"
	(*globalUint16Ptr) = ^(*globalUint16Ptr) // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "NOTW.*(AX)"
	(*globalUint32Ptr) = ^(*globalUint32Ptr) // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "NOTL.*(AX)"
	(*globalUint64Ptr) = ^(*globalUint64Ptr) // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "NOTQ.*(AX)"
}

func constModifyNEG() {
	globalInt8 = -globalInt8     // amd64:"NEGB.*globalInt8\\(SB\\)"
	globalInt16 = -globalInt16   // amd64:"NEGW.*globalInt16\\(SB\\)"
	globalInt32 = -globalInt32   // amd64:"NEGL.*globalInt32\\(SB\\)"
	globalInt64 = -globalInt64   // amd64:"NEGQ.*globalInt64\\(SB\\)"
	globalUint8 = -globalUint8   // amd64:"NEGB.*globalUint8\\(SB\\)"
	globalUint16 = -globalUint16 // amd64:"NEGW.*globalUint16\\(SB\\)"
	globalUint32 = -globalUint32 // amd64:"NEGL.*globalUint32\\(SB\\)"
	globalUint64 = -globalUint64 // amd64:"NEGQ.*globalUint64\\(SB\\)"

	(*globalInt8Ptr) = -(*globalInt8Ptr)     // amd64:"MOVQ.*globalInt8Ptr\\(SB\\), AX", "NEGB.*(AX)"
	(*globalInt16Ptr) = -(*globalInt16Ptr)   // amd64:"MOVQ.*globalInt16Ptr\\(SB\\), AX", "NEGW.*(AX)"
	(*globalInt32Ptr) = -(*globalInt32Ptr)   // amd64:"MOVQ.*globalInt32Ptr\\(SB\\), AX", "NEGL.*(AX)"
	(*globalInt64Ptr) = -(*globalInt64Ptr)   // amd64:"MOVQ.*globalInt64Ptr\\(SB\\), AX", "NEGQ.*(AX)"
	(*globalUint8Ptr) = -(*globalUint8Ptr)   // amd64:"MOVQ.*globalUint8Ptr\\(SB\\), AX", "NEGB.*(AX)"
	(*globalUint16Ptr) = -(*globalUint16Ptr) // amd64:"MOVQ.*globalUint16Ptr\\(SB\\), AX", "NEGW.*(AX)"
	(*globalUint32Ptr) = -(*globalUint32Ptr) // amd64:"MOVQ.*globalUint32Ptr\\(SB\\), AX", "NEGL.*(AX)"
	(*globalUint64Ptr) = -(*globalUint64Ptr) // amd64:"MOVQ.*globalUint64Ptr\\(SB\\), AX", "NEGQ.*(AX)"
}
