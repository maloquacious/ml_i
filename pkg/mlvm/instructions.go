// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package mlvm

import "fmt"

/*
	Instructions are encoded in 32-bit words per the following layout:
		+---------------------------------------------------------------+
		|3 3 2 2 2 2 2 2 2|2 2 2 1 1 1 1 1 1|1 1 1 1 0 0 0 0|0 0 0 0 0 0|
		|1 0 9 8 7 6 5 4 3|2 1 0 9 8 7 6 5 4|3 2 1 0 9 8 7 6|5 4 3 2 1 0|
		+-----------+---------------+-----------------+-----------------+
		|    OP     |       A       |       B         |       C         |
		+-----------+---------------+-----------------+-----------------+
		|    OP     |       A       |       Bx                          |
		+-----------+---------------+-----------------------------------+
		|    OP     |               |      sBx                          |
		+-----------+---------------+-----------------------------------+
		+-----------------+-----------------+---------------+-----------+
		|       C         |       B         |       A       |    OP     |
		+-----------------+-----------------+---------------+-----------+
		|                         Bx        |       A       |    OP     |
		+-----------------------------------+---------------+-----------+
		|                        sBx        |       A       |    OP     |
		+-----------------------------------+---------------+-----------+
*/

type IABC struct {
	Op byte // unsigned 6-bit field
	A  byte // unsigned 8-bit field
	B  int  // unsigned 9-bit field
	C  int  // unsigned 9-bit field
}

func EncodeABC(i IABC) Word {
	return Word((((((uint32(i.Op) << 8) + uint32(i.A)) << 9) + uint32(i.B)) << 9) + uint32(i.C))
}

func DecodeABC(w Word) IABC {
	return IABC{
		Op: byte((w & 0xFC00_0000) >> 26),
		A:  byte((w & 0x03FC_0000) >> 18),
		B:  int((w & 0x0003_FE00) >> 9),
		C:  int(w & 0x0000_01FF),
	}
}

// String implements the Stringer interface.
func (i IABC) String() string {
	return fmt.Sprintf("(%x,%x,%x,%x)", i.Op, i.A, i.B, i.C)
}

type IABx struct {
	Op byte // unsigned 6-bit field
	A  byte // unsigned 8-bit field
	Bx uint // unsigned 18-bit field
}

func EncodeABx(i IABx) Word {
	return Word((((uint32(i.Op) << 8) + uint32(i.A)) << 18) + uint32(i.Bx))
}

func DecodeABx(w Word) IABx {
	return IABx{
		Op: byte((w & 0xFC00_0000) >> 26),
		A:  byte((w & 0x03FC_0000) >> 18),
		Bx: uint(w & 0x0003_FFFF),
	}
}

// String implements the Stringer interface.
func (i IABx) String() string {
	return fmt.Sprintf("(%x,%x,%x)", i.Op, i.A, i.Bx)
}

// SBx is signed and uses a bias when encoding the value.
// The bias is half the maximum integer that can be stored by Bx.
// Bx is 18 bits and has a maximum value of 0x3_FFFF, so the bias is 0x3_FFFF / 2.
// A value of -1 will be encoded as (-1 + bias), which is 0x1_FFFE.
const sBxBias = 0x0003_FFFF >> 1

type IAsBx struct {
	Op  byte // unsigned 6-bit field
	A   byte // unsigned 8-bit field
	SBx int  // signed  18-bit field
}

func EncodeAsBx(i IAsBx) Word {
	return Word((((uint32(i.Op) << 8) + uint32(i.A)) << 18) + uint32(i.SBx+sBxBias))
}

func DecodeAsBx(w Word) IAsBx {
	return IAsBx{
		Op:  byte((w & 0xFC00_0000) >> 26),
		A:   byte((w & 0x03FC_0000) >> 18),
		SBx: int(w&0x0003_FFFF) - sBxBias,
	}
}

// String implements the Stringer interface.
func (i IAsBx) String() string {
	return fmt.Sprintf("(%x,%x,%d)", i.Op, i.A, i.SBx)
}
