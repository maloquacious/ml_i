// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"io"
)

const (
	MAX_WORDS = 65_536
)

type VM struct {
	Name     string // name of the virtual machine
	PC       int
	A, B, C  int
	ACmp     CMPRSLT
	Start    int // starting address
	BranchPC int // set by GOADD
	Core     [MAX_WORDS]Word
}

type Word struct {
	Op    op.Code
	Value int
	Text  string
}

func (m *VM) Run(fp io.Writer) error {
	m.PC = m.Start
	fmt.Printf("vm: starting %d\n", m.Start)
	var w Word
	for counter, running := 10_000, true; running && counter > 0; counter-- {
		w, m.PC = m.Core[m.PC], m.PC+1
		switch w.Op {
		case op.AAL: // add a number to register A
			m.A = m.A + w.Value
		case op.AAV: // add a variable to register A
			m.A = m.A + m.Core[w.Value].Value
		case op.ABV: // add to B a variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.ALIGN: // align A up to next boundary
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.ANDL: // "and" A with a number
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.ANDV: // "and" A with a variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.BMOVE: // backwards block move
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.BSTK: // stack A on backwards stack
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.BUMP: // increase a variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CAI: // compare A indirect signed integer or // compare A with indirect address
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CAL: // compare A with number
			m.compareA(w.Value)
		case op.CAV: // compare register A to a variable
			m.compareA(m.Core[w.Value].Value)
		case op.CCI: // compare C indirect
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CCL: // compare C with number
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CCN: // compare C with named character
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CFSTK: // stack C on forwards stack
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CLEAR: // set variable to zero
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CON: // numerical constant
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CSS: // clear subroutine stack (if any)
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.DCL: // declare variable
			return fmt.Errorf("%d: executing %q", m.PC-1, w.Op)
		case op.EQU: // equate two variables
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.EXIT: // exit from subroutine
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.FMOVE: // forwards block move
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.FSTK: // stack A on forwards stack
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.GO: // unconditional branch
			// argh with all the flags
			m.PC = w.Value
		case op.GOADD: // multi-way branch
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.GOEQ: // branch if equal
			if m.ACmp == IS_EQ {
				m.PC = w.Value
			}
		case op.GOGE: // branch if greater than or equal
			if m.ACmp == IS_GR || m.ACmp == IS_EQ {
				m.PC = w.Value
			}
		case op.GOGR: // branch if greater than
			if m.ACmp == IS_GR {
				m.PC = w.Value
			}
		case op.GOLE: // branch if less than or equal
			if m.ACmp == IS_LT || m.ACmp == IS_EQ {
				m.PC = w.Value
			}
		case op.GOLT: // branch if less than
			if m.ACmp == IS_LT {
				m.PC = w.Value
			}
		case op.GOND: // branch if C is not a digit; otherwise put value in A
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.GONE: // branch if not equal
			if m.ACmp != IS_EQ {
				m.PC = w.Value
			}
		case op.GOPC: // branch if C is a punctuation character
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.GOSUB: // call subroutine
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.HALT:
			running = false
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.IDENT: // equate name to integer
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.LAA: // load A modified (variable) // load A modified (table item)
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.LAI: // load A indirect
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.LAL: // load A with number
			m.A = w.Value
		case op.LAM: // load A modified
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.LAV: // load variable into register A
			m.A = m.Core[w.Value].Value
		case op.LBV: // load B with variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.LCI: // load C indirect
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.LCM: // load C modified
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.LCN: // load C with named character
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.MDLABEL:
			return fmt.Errorf("%d: internal error: %q", w.Op)
		case op.MESS: // copy text to output stream
			_, _ = fmt.Fprintf(fp, "%s", w.Text)
		case op.MULTL: // multiply register A by a number
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.NB: // comment
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.NCH: // character constant
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.PRGEN: // end of logic
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.PRGST: // start of logic
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.SAL: // subtract a number from register A
			m.A = m.A - w.Value
		case op.SAV: // subtract a variable from register A
			m.A = m.A - m.Core[w.Value].Value
		case op.SBL: // subtract from B a number
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.SBV: // subtract from B a variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.STI: // store A indirectly in variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.STR: // character string constant
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.STV: // store register A in variable
			m.Core[w.Value].Value = m.A
		case op.SUBR: // declare subroutine
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.UNSTK: // unstack from backwards stack
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		default:
			return fmt.Errorf("assert(op != %q != %d)", w.Op, w.Op)
		}
	}
	return nil
}
