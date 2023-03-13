package lowl

type CMPAT int

// enums for CMPAT
const (
	IS_LT CMPAT = -1 // register is less than value
	IS_EQ CMPAT = 0  // register is equal to value
	IS_GR CMPAT = 1  // register is greater than value
)
