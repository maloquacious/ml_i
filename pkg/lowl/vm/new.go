// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import "github.com/maloquacious/ml_i/pkg/lowl/op"

// New - yes
func New() *VM {
	// when we start running the machine, the PC will be set to the first instruction in the program.
	m := &VM{PC: 0}
	m.Registers.LCH = 1
	m.Registers.LNM = 1

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

func (m *VM) SetWord(pc int, w Word) {
	m.Core[pc] = w
}
