package lowl

import "github.com/maloquacious/ml_i/pkg/lowl/op"

type ADDR = uint16
type WORD = int16
type Word struct {
	op      op.Code
	data    WORD
	comment string
}

const (
	MAX_WORDS = 65_536 // 2**16 words
	MAX_STACK = 8_192  // 2**14 words

	// LCH is the number of storage units occupied by an item of character data
	LCH = 1
	// LNM is the number of storage units occupied by an item of numerical data.
	LNM = 1
	// LICH is the number of something.
	LICH = 1 / LCH
)

type VM struct {
	// a, b, and c are the registers
	a, b, c WORD
	cmpa    CMPAT

	// pc is the program counter
	pc     ADDR
	halted bool

	// core is the main block of memory
	core [MAX_WORDS]Word

	heaps struct {
		// msg holds the MESS characters
		msg  map[int]string // index is line number
		vars []WORD
	}
}

func (vm *VM) compare(r, v WORD) CMPAT {
	if r < v {
		return IS_LT
	} else if r == v {
		return IS_EQ
	}
	return IS_GR
}
