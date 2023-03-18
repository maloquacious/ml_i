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
	case op.AAL: // add a number to register A
		m.A = m.A + w.Value
	case op.AAV: // add a variable to register A
		m.A = m.A + m.directLoad(w.Value)
	case op.ABV: // add a variable to register B
		m.B = m.B + m.directLoad(w.Value)
	case op.ALIGN: // align A up to next boundary
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.ANDL: // bitwise "AND" a number with register A
		m.A = m.A & w.Value
	case op.ANDV: // bitwise AND a variable with register A
		m.A = m.A & m.directLoad(w.Value)
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
			return fmt.Errorf("%d: FS underflow", m.PC-1)
		}
		m.Stack[lfpt] = m.A
		m.directStore(m.Registers.LFPT, lfpt)
	case op.BUMP: // increase a variable by an amount
		m.directStore(w.Value, m.directLoad(w.Value)+w.ValueTwo)
	case op.CAI: // compare A m.indirect signed integer or // compare A with m.indirect address
		m.compare(m.A, m.indirectLoad(w.Value))
	case op.CAL: // compare register A with number
		m.compare(m.A, w.Value)
	case op.CAV: // compare register A to a variable
		m.compare(m.A, m.directLoad(w.Value))
	case op.CCI: // compare register C m.indirect
		m.compare(m.C, m.indirectLoad(w.Value))
	case op.CCL: // compare register C with a number
		m.compare(m.C, w.Value)
	case op.CCN: // compare register C with named character
		m.compare(m.C, w.Value)
	case op.CFSTK: // stack C on forwards stack
		ffpt, lfpt := m.directLoad(m.Registers.FFPT), m.directLoad(m.Registers.LFPT)
		if ffpt+1 >= lfpt { // ERLSO
			return fmt.Errorf("%d: FS overflow", m.PC-1)
		}
		m.Stack[ffpt], ffpt = m.C, ffpt+1
		m.directStore(m.Registers.FFPT, ffpt)
	case op.CLEAR: // set variable to zero
		m.directStore(w.Value, 0)
	case op.CON: // numerical constant
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.CSS: // pop address of the subroutine stack
		m.RS = m.RS[:len(m.RS)-1]
	case op.DCL: // declare variable
		return fmt.Errorf("%d: executing %q", m.PC-1, w.Op)
	case op.EQU: // equate two variables
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
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
			return fmt.Errorf("%d: FS overflow", m.PC-1)
		}
		m.Stack[ffpt], ffpt = m.A, ffpt+1
		m.directStore(m.Registers.FFPT, ffpt)
	case op.GO: // unconditional branch
		// argh with all the flags
		m.PC = w.Value
	case op.GOADD: // multi-way branch
		m.Registers.JumpValue = m.directLoad(w.Value)
	case op.GOEQ: // branch if equal
		if m.Cmp == IS_EQ {
			m.PC = w.Value
		}
	case op.GOGE: // branch if greater than or equal
		if m.Cmp == IS_GR || m.Cmp == IS_EQ {
			m.PC = w.Value
		}
	case op.GOGR: // branch if greater than
		if m.Cmp == IS_GR {
			m.PC = w.Value
		}
	case op.GOLE: // branch if less than or equal
		if m.Cmp == IS_LT || m.Cmp == IS_EQ {
			m.PC = w.Value
		}
	case op.GOLT: // branch if less than
		if m.Cmp == IS_LT {
			m.PC = w.Value
		}
	case op.GOND: // branch if C is not a digit; otherwise put value in A
		if isdigit(byte(m.C)) {
			m.A = m.C
		} else {
			m.PC = w.Value
		}
	case op.GONE: // branch if not equal
		if m.Cmp != IS_EQ {
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
	case op.IDENT: // equate name to integer
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.LAA: // load address of variable V into register A
		m.A = w.Value
	case op.LAI: // load contents of variable plus register A into register A
		m.A = m.indirectLoad(w.Value)
	case op.LAL: // load number into register A
		m.A = w.Value
	case op.LAM: // load contents of address pointed to by register B + N-OF into register A
		m.A = m.indirectLoad(m.B + w.Value)
	case op.LAV: // load variable into register A
		m.A = m.directLoad(w.Value)
	case op.LBV: // load variable into register B
		m.B = m.directLoad(w.Value)
	case op.LCI: // load contents of variable plus register C into register C
		m.C = m.indirectLoad(w.Value)
	case op.LCM: // load C modified
		m.C = m.indirectLoad(m.B + w.Value)
	case op.LCN: // load C with named character
		m.C = w.Value
	case op.MDERCH: // copy register C to output stream
		if m.C == '$' {
			printf(stdout, "\n")
		} else {
			printf(stdout, "%s", string(byte(m.C)))
		}
	case op.MDLABEL:
		return fmt.Errorf("%d: internal error: %+v", m.PC-1, w.Op)
	case op.MDQUIT: // graceful exit requested
		// force the program counter back to this instruction
		m.PC = m.PC - 1
		// signal that we have stopped the machine
		m.Registers.Halted = true
		return ErrQuit
	case op.MESS: // copy text to output stream
		printf(stdout, "%s", strings.ReplaceAll(w.Text, "$", "\n"))
	case op.MULTL: // multiply register A by a number
		m.A = m.A * w.Value
	case op.NB: // comment
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.NCH: // character constant
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.NOOP: // noop
		// do nothing
	case op.PRGEN: // end of logic
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.PRGST: // start of logic
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.SAL: // subtract a number from register A
		m.A = m.A - w.Value
	case op.SAV: // subtract a variable from register A
		m.A = m.A - m.directLoad(w.Value)
	case op.SBL: // subtract a number from register B
		m.B = m.B - w.Value
	case op.SBV: // subtract a variable from register B
		m.B = m.B - m.directLoad(w.Value)
	case op.STI: // store register A in address pointed at by variable V
		m.indirectStore(w.Value, m.A)
	case op.STR: // character string constant
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.STV: // store register A in variable
		m.directStore(w.Value, m.A)
	case op.SUBR: // declare subroutine
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	case op.UNSTK: // unstack from backwards stack
		return fmt.Errorf("%d: %+v\n", m.PC-1, w)
	default:
		return fmt.Errorf("assert(op != %q != %d): %w", w.Op, w.Op, ErrInvalidOp)
	}

	return nil
}
