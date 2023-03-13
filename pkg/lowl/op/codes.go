// Package op defines the enums for opcodes.
package op

import "fmt"

// Data types
//    Character (single character), number (may be integer value or pointer).
// Variables
//    Represented by identifiers. No character variables.
// Constants
//    Numerical: decimal integer or call of OF macro.
//    Character: single character in quotes, or name.
// Registers
//    Three: A, B and C.
// Labels
//    Represented by identifiers. Enclosed in square brackets where placed.
// Subroutines
//    Names are identifiers. At most one argument.
//

// Code is an opcode in the assembler and virtual machine.
type Code byte

// enums for opcodes
const (
	PANIC Code = iota
	AAL        // add a literal to A
	AAV        // add a variable to A
	ABV        // add a variable to B
	ALIGN      // align A up to next boundary
	ANDL       // "and" A with a literal
	ANDV       // "and" A with a variable
	BMOVE      // backwards block move
	BSTK       // push register A on backwards stack
	BUMP       // increase a variable
	CAI        // compare A indirect
	CAL        // compare A with literal
	CAV        // compare A with variable
	CCI        // compare C indirect
	CCL        // compare C with literal
	CCN        // compare C with named character
	CFSTK      // push register C on forwards stack
	CLEAR      // set variable to zero
	CON        // numerical constant
	CSS        // clear subroutine stack (if any)
	DCL        // declare variable
	EQU        // equate two variables
	EXIT       // exit from subroutine
	FMOVE      // forwards block move
	FSTK       // push register A on forwards stack
	GO         // unconditional branch
	GOADD      // multi-way branch
	GOEQ       // branch if equal
	GOGE       // branch if greater than or equal
	GOGR       // branch if greater than
	GOLE       // branch if less than or equal
	GOLT       // branch if less than
	GOND       // branch if C is not a digit; otherwise put value in A
	GONE       // branch if not equal
	GOPC       // branch if C is a punctuation character
	GOSUB      // call subroutine
	IDENT      // equate name to integer
	LAA        // load A modified (variable)
	LAI        // load A indirect
	LAL        // load A with literal
	LAM        // load A modified
	LAV        // load A with variable
	LBV        // load B with variable
	LCI        // load C indirect
	LCM        // load C modified
	LCN        // load C with named character
	MESS       // output a message
	MULTL      // multiply A by a literal
	NB         // comment
	NCH        // character constant
	NOOP       // no op
	PRGEN      // end of logic
	PRGST      // start of logic
	SAL        // subtract a literal from A
	SAV        // subtract a variable from A
	SBL        // subtract a literal from B
	SBV        // subtract a variable from B
	STI        // store A indirectly in variable
	STR        // character string constant
	STV        // store A in variable
	SUBR       // declare subroutine
	UNSTK      // pop value from backwards stack
)

// String implements the Stringer interface.
// Returns the mnemonic for known op codes, and the hex code for unknown op codes.
func (op Code) String() string {
	switch op {
	case PANIC:
		return "PANIC"
	case AAL:
		return "AAL"
	case AAV:
		return "AAV"
	case ABV:
		return "ABV"
	case ALIGN:
		return "ALIGN"
	case ANDL:
		return "ANDL"
	case ANDV:
		return "ANDV"
	case BMOVE:
		return "BMOVE"
	case BSTK:
		return "BSTK"
	case BUMP:
		return "BUMP"
	case CAI:
		return "CAI"
	case CAL:
		return "CAL"
	case CAV:
		return "CAV"
	case CCI:
		return "CCI"
	case CCL:
		return "CCL"
	case CCN:
		return "CCN"
	case CFSTK:
		return "CFSTK"
	case CLEAR:
		return "CLEAR"
	case CON:
		return "CON"
	case CSS:
		return "CSS"
	case DCL:
		return "DCL"
	case EQU:
		return "EQU"
	case EXIT:
		return "EXIT"
	case FMOVE:
		return "FMOVE"
	case FSTK:
		return "FSTK"
	case GO:
		return "GO"
	case GOADD:
		return "GOADD"
	case GOEQ:
		return "GOEQ"
	case GOGE:
		return "GOGE"
	case GOGR:
		return "GOGR"
	case GOLE:
		return "GOLE"
	case GOLT:
		return "GOLT"
	case GOND:
		return "GOND"
	case GONE:
		return "GONE"
	case GOPC:
		return "GOPC"
	case GOSUB:
		return "GOSUB"
	case IDENT:
		return "IDENT"
	case LAA:
		return "LAA"
	case LAI:
		return "LAI"
	case LAL:
		return "LAL"
	case LAM:
		return "LAM"
	case LAV:
		return "LAV"
	case LBV:
		return "LBV"
	case LCI:
		return "LCI"
	case LCM:
		return "LCM"
	case LCN:
		return "LCN"
	case MESS:
		return "MESS"
	case MULTL:
		return "MULTL"
	case NB:
		return "NB"
	case NCH:
		return "NCH"
	case PRGEN:
		return "PRGEN"
	case PRGST:
		return "PRGST"
	case SAL:
		return "SAL"
	case SAV:
		return "SAV"
	case SBL:
		return "SBL"
	case SBV:
		return "SBV"
	case STI:
		return "STI"
	case STR:
		return "STR"
	case STV:
		return "STV"
	case SUBR:
		return "SUBR"
	case UNSTK:
		return "UNSTK"
	}
	// have to cast op as an int to avoid recursion into String()!
	return fmt.Sprintf("%02x", int(op))
}
