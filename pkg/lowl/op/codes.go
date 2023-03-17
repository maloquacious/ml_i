// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

// Package op defines the enums for opcodes.
package op

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
	HALT   Code = iota
	AAL         // add a literal to A
	AAV         // add a variable to A
	ABV         // add a variable to B
	ALIGN       // align A up to next boundary
	ANDL        // "and" A with a literal
	ANDV        // "and" A with a variable
	BMOVE       // backwards block move
	BSTK        // push register A on backwards stack
	BUMP        // increase a variable
	CAI         // compare A indirect
	CAL         // compare A with literal
	CAV         // compare A with variable
	CCI         // compare C indirect
	CCL         // compare C with literal
	CCN         // compare C with named character
	CFSTK       // push register C on forwards stack
	CLEAR       // set variable to zero
	CON         // numerical constant
	CSS         // clear subroutine stack (if any)
	DCL         // declare variable
	EQU         // equate two variables
	EXIT        // exit from subroutine
	EXITEQ      // branch if equal
	EXITGE      // branch if greater than or equal
	EXITGR      // branch if greater than
	EXITLE      // branch if less than or equal
	EXITLT      // branch if less than
	EXITND      // branch if C is not a digit; otherwise put value in A
	EXITNE      // branch if not equal
	EXITPC      // branch if C is a punctuation character
	FMOVE       // forwards block move
	FSTK        // push register A on forwards stack
	GO          // unconditional branch
	GOADD       // multi-way branch
	GOEQ        // branch if equal
	GOGE        // branch if greater than or equal
	GOGR        // branch if greater than
	GOLE        // branch if less than or equal
	GOLT        // branch if less than
	GOND        // branch if C is not a digit; otherwise put value in A
	GONE        // branch if not equal
	GOPC        // branch if C is a punctuation character
	GOSUB       // call subroutine
	GOTBL       // jump table for exit instructions
	IDENT       // equate name to integer
	LAA         // load A modified (variable)
	LAI         // load A indirect
	LAL         // load A with literal
	LAM         // load A modified
	LAV         // load A with variable
	LBV         // load B with variable
	LCI         // load C indirect
	LCM         // load C modified
	LCN         // load C with named character
	MESS        // output a message
	MULTL       // multiply A by a literal
	NB          // comment
	NCH         // character constant
	NOOP        // no op
	PRGEN       // end of logic
	PRGST       // start of logic
	SAL         // subtract a literal from A
	SAV         // subtract a variable from A
	SBL         // subtract a literal from B
	SBV         // subtract a variable from B
	STI         // store A indirectly in variable
	STR         // character string constant
	STV         // store A in variable
	SUBR        // declare subroutine
	UNSTK       // pop value from backwards stack
	// implementation dependent op codes
	MDERCH  // MDERCH - emit character in register C
	MDLABEL // declare a label
	MDQUIT  // MDQUIT - exit the program
	UNKNOWN // not really an opcode
)
