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

func (m *VM) Run() error {
	m.PC = m.Start
	fmt.Printf("vm: starting %d\n", m.Start)
	var w Word
	for counter, running := 10_000, true; running && counter > 0; counter-- {
		w, m.PC = m.Core[m.PC], m.PC+1
		switch w.Op {
		case op.AAL: // add a literal to register A
			m.A = m.A + w.Value
		case op.AAV: // add to A a variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.ABV: // add to B a variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.ALIGN: // align A up to next boundary
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.ANDL: // "and" A with a literal
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
		case op.CAL: // compare A with literal
			m.compareA(w.Value)
		case op.CAV: // compare A with variable signed integer // compare A with address
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CCI: // compare C indirect
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.CCL: // compare C with literal
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
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.GOGR: // branch if greater than
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.GOLE: // branch if less than or equal
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.GOLT: // branch if less than
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
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
		case op.LAL: // load A with literal
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
			fmt.Printf("%s", w.Text)
		case op.MULTL: // multiply register A by a literal
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.NB: // comment
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.NCH: // character constant
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.PRGEN: // end of logic
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.PRGST: // start of logic
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.SAL: // subtract from A a literal
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.SAV: // subtract from A a variable
			return fmt.Errorf("%d: %+v\n", m.PC-1, w)
		case op.SBL: // subtract from B a literal
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
