// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"io"
	"strings"
)

func (m *VM) Step(stdout, stderr io.Writer) error {
	if m.Registers.Halted {
		return ErrHalted
	}

	w := m.Core[m.PC]
	m.PC = m.PC + 1

	switch w.Op {
	case op.AAL: // add a literal value to register A
		literalValue := w.Value
		m.A = m.A + literalValue
	case op.AAV: // add a variable to register A
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.A = m.A + variableValue
	case op.ABV: // add a variable to register B
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.B = m.B + variableValue
	case op.ANDL: // bitwise "AND" a literal value with register A
		literalValue := w.Value
		m.A = m.A & literalValue
	case op.ANDV: // bitwise AND a variable with register A
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.A = m.A & variableValue
	case op.BMOVE: // backwards block move
		// SRCPT points at the start of the source field.
		// DSTPT points to the start of the destination field.
		// Register A contains the length of the field (number of words to move)
		srcpt, dstpt := m.indirectLoad(m.Registers.SRCPT), m.indirectLoad(m.Registers.DSTPT)
		for a := m.A - 1; a >= 0; a-- {
			m.Core[dstpt+a].Value = m.Core[srcpt+a].Value
		}
	case op.BSTK: // stack A on backwards stack
		ffpt, lfpt := m.directLoad(m.Registers.FFPT), m.directLoad(m.Registers.LFPT)
		lfpt = lfpt - 1     // decrement before pushing
		if ffpt+1 >= lfpt { // ERLSO
			return fmt.Errorf("%d: FS: %w", m.PC-1, ErrStackUnderflow)
		}
		m.Stack[lfpt] = m.A
		m.directStore(m.Registers.LFPT, lfpt)
	case op.BUMP: // increase a variable by a literal value
		literalValue := w.ValueTwo
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.directStore(w.Value, literalValue+variableValue)
	case op.CAI: // compare contents of address pointed to by V to register A
		variableAddress := w.Value
		indirectValue := m.indirectLoad(variableAddress)
		m.compare(m.A, indirectValue)
	case op.CAL: // compare register A with a literal value
		literalValue := w.Value
		m.compare(m.A, literalValue)
	case op.CAV: // compare A with the value of variable
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.compare(m.A, variableValue)
	case op.CCI: // compare contents of address pointed to by V to register C
		variableAddress := w.Value
		indirectValue := m.indirectLoad(variableAddress)
		m.compare(m.C, indirectValue)
	case op.CCL: // compare register C with a literal value
		m.compare(m.C, w.Value)
	case op.CCN: // compare register C with named character
		literalValue := w.Value
		m.compare(m.C, literalValue)
	case op.CFSTK: // stack C on forwards stack
		ffpt, lfpt := m.directLoad(m.Registers.FFPT), m.directLoad(m.Registers.LFPT)
		if ffpt+1 >= lfpt { // ERLSO
			return fmt.Errorf("%d: FS: %w", m.PC-1, ErrStackOverflow)
		}
		m.Stack[ffpt], ffpt = m.C, ffpt+1
		m.directStore(m.Registers.FFPT, ffpt)
	case op.CLEAR: // set variable to zero
		variableAddress := w.Value
		m.directStore(variableAddress, 0)
	case op.CSS: // pop address of the subroutine stack
		m.RS = m.RS[:len(m.RS)-1]
	case op.EXIT: // exit from subroutine
		// pop the return address from the stack
		m.PC, m.RS = m.RS[len(m.RS)-1], m.RS[:len(m.RS)-1]
		// update the test register used by GOADD and GOBRPC
		m.Registers.JumpValue = w.Value
	case op.FMOVE: // forwards block move
		// SRCPT points at the start of the source field.
		// DSTPT points to the start of the destination field.
		// Register A contains the length of the field (number of words to move)
		srcpt, dstpt := m.indirectLoad(m.Registers.SRCPT), m.indirectLoad(m.Registers.DSTPT)
		for a := 0; a < m.A; a++ {
			m.Core[dstpt+a].Value = m.Core[srcpt+a].Value
		}
	case op.FSTK: // stack A on forwards stack
		ffpt, lfpt := m.directLoad(m.Registers.FFPT), m.directLoad(m.Registers.LFPT)
		if ffpt+1 >= lfpt { // ERLSO
			return fmt.Errorf("%d: FS: %w", m.PC-1, ErrStackOverflow)
		}
		m.Stack[ffpt], ffpt = m.A, ffpt+1
		m.directStore(m.Registers.FFPT, ffpt)
	case op.GO: // unconditional branch
		m.PC = w.Value
	case op.GOADD: // multi-way branch
		m.Registers.JumpValue = m.directLoad(w.Value)
	case op.GOEQ: // branch if equal
		if m.Registers.Cmp == IS_EQ {
			m.PC = w.Value
		}
	case op.GOGE: // branch if greater than or equal
		if m.Registers.Cmp == IS_GR || m.Registers.Cmp == IS_EQ {
			m.PC = w.Value
		}
	case op.GOGR: // branch if greater than
		if m.Registers.Cmp == IS_GR {
			m.PC = w.Value
		}
	case op.GOLE: // branch if less than or equal
		if m.Registers.Cmp == IS_LT || m.Registers.Cmp == IS_EQ {
			m.PC = w.Value
		}
	case op.GOLT: // branch if less than
		if m.Registers.Cmp == IS_LT {
			m.PC = w.Value
		}
	case op.GOND: // branch if C is not a digit; otherwise put value in A
		if !isdigit(byte(m.C)) {
			m.PC = w.Value
		} else {
			m.A = m.C - '0'
		}
	case op.GONE: // branch if not equal
		if m.Registers.Cmp != IS_EQ {
			m.PC = w.Value
		}
	case op.GOPC: // branch if C is a punctuation character
		if ispunct(byte(m.C)) {
			m.PC = w.Value
		}
	case op.GOSUB: // call subroutine
		// push return address on to the return stack
		m.RS = append(m.RS, m.PC)
		// go to the subroutine
		m.PC = w.Value
	case op.GOTBL: // jump table for exit instructions
		if w.ValueTwo == m.Registers.JumpValue {
			m.PC = w.Value
		}
	case op.HALT: // halt the machine
		// force the program counter back to this instruction
		m.PC = m.PC - 1
		// signal that we have halted
		m.Registers.Halted = true
		return ErrHalted
	case op.LAA: // load address of variable V into register A
		variableAddress := w.Value
		m.A = variableAddress
	case op.LAI: // load A with contents of the address pointed to by variable V
		variableAddress := w.Value
		indirectValue := m.indirectLoad(variableAddress)
		m.A = indirectValue
	case op.LAL: // load literal value into register A
		literalValue := w.Value
		m.A = literalValue
	case op.LAM: // load contents of address pointed to by register B + N-OF into register A
		literalValue := w.Value
		indexedValue := m.indexedLoad(literalValue)
		m.A = indexedValue
	case op.LAV: // load A with value of variable V
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.A = variableValue
	case op.LBV: // load B with value of variable B
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.B = variableValue
	case op.LCI: // load C with contents of the address pointed to by variable V
		variableAddress := w.Value
		indirectValue := m.indirectLoad(variableAddress)
		m.C = indirectValue
	case op.LCM: // load contents of address pointed to by register B + N-OF into register A
		literalValue := w.Value
		indexedValue := m.indexedLoad(literalValue)
		m.C = indexedValue
	case op.LCN: // load C with named character
		literalValue := w.Value
		m.C = literalValue
	case op.MDERCH: // copy register C to output stream
		if m.C == '$' {
			printf(stdout, "\n")
		} else {
			printf(stdout, "%s", string(byte(m.C)))
		}
	case op.MDQUIT: // graceful exit requested
		// force the program counter back to this instruction
		m.PC = m.PC - 1
		// signal that we have stopped the machine
		m.Registers.Halted = true
		return ErrQuit
	case op.MESS: // copy text to output stream
		printf(stdout, "%s", strings.ReplaceAll(w.Text, "$", "\n"))
	case op.MULTL: // multiply register A by a literal value
		literalValue := w.Value
		m.A = m.A * literalValue
	case op.NOOP: // noop
		// do nothing
	case op.SAL: // subtract a literal value from register A
		literalValue := w.Value
		m.A = m.A - literalValue
	case op.SAV: // subtract a variable from register A
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.A = m.A - variableValue
	case op.SBL: // subtract a literal value from register B
		literalValue := w.Value
		m.B = m.B - literalValue
	case op.SBV: // subtract a variable from register B
		variableAddress := w.Value
		variableValue := m.directLoad(variableAddress)
		m.B = m.B - variableValue
	case op.STI: // store register A in address pointed at by variable V
		variableAddress := w.Value
		variableValue := m.A
		m.indirectStore(variableAddress, variableValue)
	case op.STV: // store register A in variable
		variableAddress := w.Value
		variableValue := m.A
		m.directStore(variableAddress, variableValue)
	case op.UNSTK: // unstack from backwards stack
		return fmt.Errorf("%d: %s: %w\n", m.PC-1, w.Op, ErrNotImplemented)
	default:
		return fmt.Errorf("assert(op != %q != %d): %w", w.Op, w.Op, ErrInvalidOp)
	}

	return nil
}
