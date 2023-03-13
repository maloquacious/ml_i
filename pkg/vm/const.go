package vm

const (
	MAX_WORDS = 65_536 // 2**16 words
	MAX_STACK = 8_192  // 2**14 words

	// LCH is the number of storage units occupied by an item of character data
	LCH = 1
	// LNM is the number of storage units occupied by an item of numerical data.
	LNM = 1
	// LICH is the number of something.
	LICH = 1 / LCH

	// these values are set after comparing a register to a value
	IS_LT = -1
	IS_EQ = 0
	IS_GT = 1
)
