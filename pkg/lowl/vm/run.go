// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import (
	"errors"
	"io"
)

func (m *VM) Run(fp, msg io.Writer) error {
	m.PC = m.Registers.Start
	m.Streams.Stdout = fp
	m.Streams.Messages = msg

	ffpt, lfpt := 0, len(m.Stack)
	if m.Registers.FFPT != 0 {
		m.directStore(m.Registers.FFPT, ffpt)
	}
	if m.Registers.LFPT != 0 {
		m.directStore(m.Registers.LFPT, lfpt)
	}

	printf(m.Streams.Messages, "vm: starting %d\n", m.Registers.Start)
	m.Registers.Halted = false
	for counter := 10_000; !m.Registers.Halted && counter > 0; counter-- {
		if err := m.Step(fp, msg); err != nil {
			if !errors.Is(err, ErrQuit) {
				return err
			}
			// graceful exit; cleanup and return happy
			return nil
		}
	}
	return ErrCycles
}
