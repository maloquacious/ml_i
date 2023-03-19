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
	Name      string // name of the virtual machine
	PC        int
	A, B, C   int
	Registers struct {
		Cmp         CMPRSLT
		DSTPT       int // points to the variable holding the destination field pointer (stack moves)
		FFPT        int // points to the variable holding the first free location of the forwards stack
		LCH         int // length of a character in this machine
		LFPT        int // points to the variable holding the last location in use on the backwards stack
		LNM         int // length of a number in this machine
		PARNM       int // points to the variable holding the subroutine parameter
		SRCPT       int // points to the variable holding the source field pointer (stack moves)
		Halted      bool
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
