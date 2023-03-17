// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import (
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"io"
)

const (
	MAX_WORDS = 65_536
	MAX_STACK = 8_096
)

type VM struct {
	Name    string // name of the virtual machine
	PC      int
	A, B, C int
	Cmp     CMPRSLT
	// Registers aren't registers so much as reserved addresses
	Registers struct {
		DSTPT       int // points to the variable holding the destination field pointer (stack moves)
		FFPT        int // points to the variable holding the first free location of the forwards stack
		LFPT        int // points to the variable holding the last location in use on the backwards stack
		PARNM       int // points to the variable holding the subroutine parameter
		SRCPT       int // points to the variable holding the source field pointer (stack moves)
		JumpValue   int // jump value for GOTBL
		Start, Last int // starting, last address
	}
	Streams struct {
		Stdout   io.Writer
		Messages io.Writer
	}
	Core  [MAX_WORDS]Word
	Stack [MAX_STACK]int
	RS    []int // return stack for subroutine calls
}

type Word struct {
	Op       op.Code
	Value    int
	ValueTwo int // used by GOADD and GOBRPC
	Text     string
	Source   struct {
		Line         int
		Op           op.Code
		Parameters   string
		Continuation bool
	}
}

func New() *VM {
	// when we start running the machine, the PC will be set to the first instruction in the program.
	m := &VM{}

	// first instruction should be a halt.
	m.Core[m.PC], m.PC = Word{Op: op.HALT}, m.PC+1

	// initialized the reserved addresses
	m.Registers.DSTPT, m.PC = m.PC, m.PC+1
	m.Registers.FFPT, m.PC = m.PC, m.PC+1
	m.Registers.LFPT, m.PC = m.PC, m.PC+1
	m.Registers.PARNM, m.PC = m.PC, m.PC+1
	m.Registers.SRCPT, m.PC = m.PC, m.PC+1

	return m
}
