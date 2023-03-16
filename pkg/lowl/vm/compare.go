// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

type CMPRSLT int

// enums for CMPAT
const (
	IS_LT CMPRSLT = -1 // register is less than value
	IS_EQ CMPRSLT = 0  // register is equal to value
	IS_GR CMPRSLT = 1  // register is greater than value
)

func (m *VM) compareA(v int) {
	if m.A < v {
		m.ACmp = IS_LT
	} else if m.A == v {
		m.ACmp = IS_EQ
	} else {
		m.ACmp = IS_GR
	}
}
