package lowl

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
)

func (vm *VM) Run() error {
	for vm.step() {
		//
	}
	return nil
}

func (vm *VM) step() bool {
	const ok = true
	if vm.halted {
		return !ok
	}

	c := vm.core[vm.pc]
	fmt.Printf("step: %06d: %s\n", vm.pc, c.op.String())
	vm.pc++

	switch c.op {
	case op.PANIC:
		panic(c.op.String()) // Code = iota
	case op.AAL: // add a literal value to A
		vm.a = vm.a + c.data
		return ok
	case op.AAV: // add a variable to A
		vm.a = vm.a + vm.heaps.vars[c.data]
		return ok
	case op.ABV:
		panic(c.op.String()) // add a variable to B
	case op.ALIGN:
		panic(c.op.String()) // align A up to next boundary
	case op.ANDL:
		panic(c.op.String()) // "and" A with a literal
	case op.ANDV:
		panic(c.op.String()) // "and" A with a variable
	case op.BMOVE:
		panic(c.op.String()) // backwards block move
	case op.BSTK:
		panic(c.op.String()) // push register A on backwards stack
	case op.BUMP:
		panic(c.op.String()) // increase a variable
	case op.CAI:
		panic(c.op.String()) // compare A indirect
	case op.CAL: // compare A with literal
		vm.cmpa = vm.compare(vm.a, c.data)
		return ok
	case op.CAV: // compare A with variable
		vm.cmpa = vm.compare(vm.a, vm.heaps.vars[c.data])
		return ok
	case op.CCI:
		panic(c.op.String()) // compare C indirect
	case op.CCL:
		panic(c.op.String()) // compare C with literal
	case op.CCN:
		panic(c.op.String()) // compare C with named character
	case op.CFSTK:
		panic(c.op.String()) // push register C on forwards stack
	case op.CLEAR:
		panic(c.op.String()) // set variable to zero
	case op.CON: // defines a numerical constant
		panic(c.op.String())
	case op.CSS:
		panic(c.op.String()) // clear subroutine stack (if any)
	case op.DCL:
		panic(c.op.String()) // declare variable
	case op.EQU:
		panic(c.op.String()) // equate two variables
	case op.EXIT:
		panic(c.op.String()) // exit from subroutine
	case op.FMOVE:
		panic(c.op.String()) // forwards block move
	case op.FSTK:
		panic(c.op.String()) // push register A on forwards stack
	case op.GO: // unconditional branch
		vm.pc = ADDR(c.data)
		return ok
	case op.GOADD:
		panic(c.op.String()) // multi-way branch
	case op.GOEQ: // branch if equal
		if vm.cmpa == IS_EQ {
			vm.pc = ADDR(c.data)
		}
		return ok
	case op.GOGE: // branch if greater than or equal
		if !(vm.cmpa == IS_LT) {
			vm.pc = ADDR(c.data)
		}
		return ok
	case op.GOGR: // branch if greater than
		if vm.cmpa == IS_GR {
			vm.pc = ADDR(c.data)
		}
		return ok
	case op.GOLE:
		panic(c.op.String()) // branch if less than or equal
	case op.GOLT:
		panic(c.op.String()) // branch if less than
	case op.GOND:
		panic(c.op.String()) // branch if C is not a digit; otherwise put value in A
	case op.GONE: // branch if not equal
		if !(vm.cmpa == IS_EQ) {
			vm.pc = ADDR(c.data)
		}
		return ok
	case op.GOPC:
		panic(c.op.String()) // branch if C is a punctuation character
	case op.GOSUB:
		panic(c.op.String()) // call subroutine
	case op.IDENT:
		panic(c.op.String()) // equate name to integer
	case op.LAA:
		panic(c.op.String()) // load A modified (variable)
	case op.LAI:
		panic(c.op.String()) // load A indirect
	case op.LAL: // load A with literal
		vm.a = c.data
		return ok
	case op.LAM:
		panic(c.op.String()) // load A modified
	case op.LAV: // load A with variable
		vm.a = vm.heaps.vars[int(c.data)]
		return ok
	case op.LBV:
		panic(c.op.String()) // load B with variable
	case op.LCI:
		panic(c.op.String()) // load C indirect
	case op.LCM:
		panic(c.op.String()) // load C modified
	case op.LCN:
		panic(c.op.String()) // load C with named character
	case op.MESS: // output a message to a stream
		fmt.Print(vm.heaps.msg[int(c.data)])
		return ok
	case op.MULTL:
		panic(c.op.String()) // multiply A by a literal
	case op.NB: // comment
		// fmt.Printf("comment(%q)\n", c.comment)
		return ok
	case op.NCH:
		panic(c.op.String()) // character constant
	case op.NOOP: // no op
		return ok
	case op.PRGEN: // end of logic
		vm.halted = true
		return ok
	case op.PRGST: // start of logic
		panic("PRGST is not executable")
	case op.SAL: // subtract a literal from A
		vm.a = vm.a - c.data
		return ok
	case op.SAV: // subtract a variable from A
		vm.a = vm.a - vm.heaps.vars[c.data]
		return ok
	case op.SBL:
		panic(c.op.String()) // subtract a literal from B
	case op.SBV: // subtract a variable from B
		panic(c.op.String())
	case op.STI:
		panic(c.op.String()) // store A indirectly in variable
	case op.STR:
		panic(c.op.String()) // character string constant
	case op.STV: // store A in variable
		vm.heaps.vars[int(c.data)] = vm.a
		return ok
	case op.SUBR:
		panic(c.op.String()) // declare subroutine
	case op.UNSTK:
		panic(c.op.String()) // pop value from backwards stack
	default:
		panic(fmt.Sprintf("assert(op != %d)", c.op))
	}
}
