// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import "fmt"

var (
	ErrCycles         = fmt.Errorf("too many cycles")
	ErrHalted         = fmt.Errorf("halted")
	ErrInvalidOp      = fmt.Errorf("invalid op")
	ErrQuit           = fmt.Errorf("quit")
	ErrStackOverflow  = fmt.Errorf("stack overflow")
	ErrStackUnderflow = fmt.Errorf("stack underflow")
)
