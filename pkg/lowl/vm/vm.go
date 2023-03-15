/*
 * Copyright (c) 2023 Michael D Henderson. All rights reserved.
 */

package vm

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
)

const (
	MAX_WORDS = 65_536
)

type VM struct {
	PC   int
	Core [64]Word
}

type Word struct {
	Op      op.Code
	Address int
}

func (vm *VM) Run() error {
	return fmt.Errorf("vm.Run: not implemented")
}
