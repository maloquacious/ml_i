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
	BranchPC int // set by GOADD
	Core     [MAX_WORDS]Word
}

type Word struct {
	Op    op.Code
	Value int
	Text  string
}

func (m *VM) Run() error {
	return fmt.Errorf("vm.Run: not implemented")
}
