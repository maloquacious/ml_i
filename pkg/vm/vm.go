package vm

import "fmt"

// todo: the words in the VM should hold runes, not 16 bit integers.

type VM struct {
	// Almost all statements in LOWL involve at most one storage address,
	// and all assignments, comparisons and arithmetic operations are done
	// via registers. There are notionally three registers, as follows:
	Registers struct {
		// A is the numerical accumulator.
		// All statements that use the A register have numerical operands.
		A uint16

		// B is the index register.
		// All statements that use the A register have numerical operands.
		B uint16

		// C is the character register.
		// All statements that use the C register have a single character as operand.
		C uint16

		// CMP is set after comparing the A register.
		// It is not changed until the next comparison.
		CMP int
	}

	// The VM uses two stacks.
	Stacks struct {
		// FS is the forward stack
		FS [MAX_STACK]WORD

		// BS is the backwards stack
		BS [MAX_STACK]WORD

		// FFPT points at the first free location in the forward stack.
		// It starts at zero and increases as items are pushed on to the stack.
		FFPT int

		// LFPT points at the last location in use on the backwards stack.
		// It starts at MAX_STACK and decreases as items are pushed on to the stack.
		LFPT int

		// The stack diagram looks like this:
		//         FS +------- +
		//            | used   |
		//            | used   |
		//  FFPT ---> | unused |
		//            | unused |
		//            | unused |
		//            +------- +
		//
		//         BS +------- +
		//            | unused |
		//            | unused |
		//  LFPT ---> | used   |
		//            | used   |
		//            | used   |
		//            +------- +
	}

	Memory [MAX_WORDS]WORD

	// PC is the program counter.
	// It always points at the next word of memory to execute.
	PC int
}

type CHAR uint16
type INTEGER uint16
type WORD []uint16

// Exec executes the instruction at the current PC.
func (vm *VM) Exec() {
	// grab the instruction from the current PC location
	instruction := vm.Memory[vm.PC]
	op := OPCODE(instruction[0])

	// increment the program counter
	vm.PC++

	// execute the instruction
	switch op {
	//		AAL		N-OF						add to A a literal.
	case AAL: // add a literal value to register A
		// nOF is the first argument of the instruction and represents a literal value.
		nOF := instruction[1]
		vm.Registers.A = vm.Registers.A + nOF
	//	*	BSTK								stack A on backwards stack.
	case BSTK:
		// [BSTK] preserve A
		//        LAV  LFPT, X
		//        SAL  OF(LNM)
		//        STV  LFPT, X
		//        restore A
		//        STI  LFPT, X
		//        LAV  FFPT, X
		//        CAV  LFPT, A
		//        GOGE ERLSO, ...
		vm.Stacks.LFPT--
		vm.Stacks.BS[vm.Stacks.LFPT] = WORD{vm.Registers.A}
	//	*	CAV		V,(X)						compare A with variable.
	//	*	CAV		V,(A)
	case CAV: // compare A with variable V
		// V is the first argument of the instruction and represents the offset of a variable in memory.
		v := instruction[1]
		vm.Registers.CMP = vm.compare(vm.Registers.A, v)

	//	*	CFSTK								stack C on forwards stack.
	case CFSTK:
		// CFSTK is essentially the same as FSTK
		vm.Stacks.FS[vm.Stacks.FFPT] = WORD{vm.Registers.C}
		vm.Stacks.FFPT++
	//	*	FSTK								stack A on forwards stack.
	case FSTK:
		// [FSTK] STI  FFPT, X
		//        LAV  FFPT, X
		//        AAL  OF(LNM)
		//        STV  FFPT, P
		//        CAV  LFPT, A
		//        GOGE ERLSO, ...
		vm.Stacks.FS[vm.Stacks.FFPT] = WORD{vm.Registers.A}
		vm.Stacks.FFPT++
	//	*	GOGE	label spec					branch if greater than or equal.
	case GOGE:
		//Argument 1: Name of designated label.
		//Argument 2: distance of designated label (as for second argument to GOSUB).
		//Argument 3: E if the branch goes out of a subroutine; X otherwise.
		//Argument 4: C if the branch is an exit following a GOSUB statement;
		if vm.Registers.CMP == IS_GT || vm.Registers.CMP == IS_EQ {
			distance := instruction[2] // todo: can distance be negative?
			vm.PC = vm.PC + int(distance)
		}
	//		LAV		V,(R)						load A with variable.
	//		LAV		V,(X)
	case LAV:
		// Load A with value of V
		// V is the first argument of the instruction and represents the offset of a variable in memory.
		v := instruction[1]
		vm.Registers.A = vm.getVariable(v)
	//		STV		V,(P)						store A in variable.
	//		STV		V,(X)
	case STV: // store
		// V is the first argument of the instruction and represents the offset of a variable in memory.
		v := instruction[1]
		vm.Registers.A = vm.getVariable(v)
	//	*	UNSTK	V							unstack from backwards stack.
	case UNSTK:
		// [UNSTK] LAI LFP,T X
		//         STV V,x
		//         BUMP LF, POF T(LN M)

		// V is the first argument of the instruction and represents the offset of a variable in memory.
		v := instruction[1]
		word := vm.Stacks.BS[vm.Stacks.LFPT]
		vm.Stacks.LFPT--
		// move word to variable V, which should be an offset into memory, right?
		vm.Memory[v] = word
	default:
		panic(fmt.Sprintf("assert(op != %d)", op))
	}
}

func (vm *VM) compare(r, v uint16) int {
	if r < v {
		return IS_LT
	} else if r == v {
		return IS_EQ
	}
	return IS_GT
}

func (vm *VM) getVariable(v uint16) uint16 {
	return vm.Memory[v][0]
}
