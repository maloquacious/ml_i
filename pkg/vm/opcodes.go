// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

//	Data types
//		Character (single character), number (may be integer value or pointer).
//	Variables
//		Represented by identifiers. No character variables.
//	Constants
//		Numerical: decimal integer or call of OF macro.
//		Character: single character in quotes, or name.
//	Registers
//		Three: A, B and C.
//	Labels
//		Represented by identifiers. Enclosed in square brackets where placed.
//	Subroutines
//		Names are identifiers. At most one argument.
//

type OPCODE uint16

// enums for opcodes
const (
	PANIC OPCODE = iota
	//		AAL		N-OF						add to A a literal.
	AAL
	//		AAV		V							add to A a variable.
	//		ABV		V							add to B a variable.
	//		ALIGN								align A up to next boundary.
	//		ANDL	N							"and" A with a literal.
	//		ANDV	V							"and" A with a variable.
	//	*	BMOVE								backwards block move.
	//	*	BSTK								stack A on backwards stack.
	BSTK
	//	*	BUMP	V,N-OF						increase a variable.
	//	*	CAI		V,(X)						compare A indirect.
	//	*	CAI		V,(A)
	//	*	CAL		N-OF						compare A with literal.
	//	*	CAV		V,(X)						compare A with variable.
	//	*	CAV		V,(A)
	CAV
	//	*	CCI		V							compare C indirect.
	//	*	CCL		'character'					compare C with literal.
	//	*	CCN		charname					compare C with named character.
	//	*	CFSTK								stack C on forwards stack.
	CFSTK
	//	*	CLEAR	V							set variable to zero.
	//		CON		N-OF						numerical constant.
	//	*	CSS									clear subroutine stack (if any).
	//		DCL		V							declare variable.
	//		EQU		V,V							equate two variables.
	//	*	EXIT	N,subroutine name			exit from subroutine.
	//	*	FMOVE								forwards block move.
	//	*	FSTK								stack A on forwards stack.
	FSTK
	//		GO		label spec					unconditional branch.
	//	*	GOADD	V							multi-way branch.
	//	*	GOEQ	label spec					branch if equal.
	//	*	GOGE	label spec					branch if greater than or equal.
	GOGE
	//	*	GOGR	label spec					branch if greater than.
	//	*	GOLE	label spec					branch if less than or equal.
	//	*	GOLT	label spec					branch if less than.
	//	*	GOND	label spec					branch if C is not a digit; otherwise put value in A.
	//	*	GONE	label spec					branch if not equal.
	//	*	GOPC	label spec					branch if C is a punctuation character.
	//		GOSUB	subroutine name,(distance)	call subroutine.
	//		GOSUB	subroutine name,(X)
	//		IDENT	V,decimal integer			equate name to integer.
	//		LAA		V,D							load A modified (variable).
	//		LAA		table label,C				load A modified (table item).
	//		LAI		V,(R)						load A indirect.
	//		LAI		V,(X)
	//		LAL		N-OF						load A with literal.
	//		LAM		N-OF						load A modified.
	//		LAV		V,(R)						load A with variable.
	//		LAV		V,(X)
	LAV
	//		LBV		V							load B with variable.
	//		LCI		V,(R)						load C indirect.
	//		LCI		V,(X)
	//		LCM		N-OF						load C modified.
	//		LCN		charname					load C with named character.
	//	*	MESS	'characters'				output a message.
	//		MULTL	N-OF						multiply A by a literal.
	//		NB		'characters'				comment.
	//		NCH		charname					character constant.
	NOOP
	//		PRGEN								end of logic.
	//		PRGST	'characters'				start of logic.
	//		SAL		N-OF						subtract from A a literal.
	//		SAV		V							subtract from A a variable.
	//		SBL		N-OF						subtract from B a literal.
	//		SBV		V							subtract from B a variable.
	//		STI		V,(P)						store A indirectly in variable.
	//		STI		V,(X)
	//		STR		'characters'				character string constant.
	//		STV		V,(P)						store A in variable.
	//		STV		V,(X)
	STV
	//		SUBR	subroutine name,(PARNM),N	declare subroutine.
	//		SUBR	subroutine name,(X)
	//	*	UNSTK	V							unstack from backwards stack.
	UNSTK
)
