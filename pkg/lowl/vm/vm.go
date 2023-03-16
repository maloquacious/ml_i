// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
)

const (
	MAX_WORDS = 65_536
)

type VM struct {
	Name     string // name of the virtual machine
	PC       int
	Start    int // starting address
	BranchPC int // set by GOADD
	Core     [MAX_WORDS]Word
}

type Word struct {
	Op    op.Code
	Value int
	Text  string
}

func (m *VM) Run() error {
	m.PC = m.Start
	fmt.Printf("vm: starting %d\n", m.Start)
	var w Word
	for halt := false; !halt; {
		w, m.PC = m.Core[m.PC], m.PC+1
		switch w.Op {
		case op.DCL:
			return fmt.Errorf("%d: executing %q", m.PC-1, w.Op)
		case op.HALT:
			halt = true
		case op.MESS: // output a message
			fmt.Printf("%s", w.Text)
		default:
			panic(fmt.Sprintf("assert(op != %q != %d)", w.Op, w.Op))
		}
	}
	return fmt.Errorf("vm.Run: not implemented")
}
