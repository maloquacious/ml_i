// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm_test

import (
	"bytes"
	"errors"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"github.com/maloquacious/ml_i/pkg/lowl/vm"
	"strings"
	"testing"
)

// TestVM_Run tests basic run and halt.
func TestVM_Run(t *testing.T) {
	opc := op.HALT
	expectPC := 0
	m := &vm.VM{}
	if m.PC != expectPC || m.Core[0].Op != opc {
		t.Fatalf("setw: want pc %d op HALT: got %d %q\n", expectPC, m.PC, m.Core[m.PC].Op)
	}
	if err := m.Run(nil, nil); err != nil {
		if !errors.Is(err, vm.ErrHalted) {
			t.Fatalf("run: want ErrHalt: got %v\n", err)
		}
	}
	if m.PC != expectPC {
		t.Errorf("pc: wants %d: got %d\n", expectPC, m.PC)
	}
}

// TestVM_DirectLoads tests all direct loads into registers.
func TestVM_DirectLoads(t *testing.T) {
	opc := op.LAL
	input, expect := 3, 3
	m := &vm.VM{}
	m.SetWord(0, vm.Word{Op: opc, Value: input})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	} else if m.A != expect {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expect, m.A)
	}
}

// TestVM_Math tests register math.
func TestVM_Math(t *testing.T) {
	opc := op.AAL
	inputA, inputB, inputC, inputV := 3, 4, 5, 88
	expectPC, expectA, expectB, expectC := 1, inputA+inputV, inputB, inputC
	m := &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}

	opc = op.AAV
	inputA, inputB, inputC, addrV, inputV := 3, 4, 5, 1, 88
	expectPC, expectA, expectB, expectC, expectV := 1, inputA+inputV, inputB, inputC, inputV
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: addrV})
	m.SetWord(addrV, vm.Word{Op: op.CON, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}
	if valOfV := m.Core[addrV].Value; valOfV != expectV {
		t.Errorf("%s: *v: want %d: got %d\n", opc, inputV, expectV)
	}

	opc = op.ABV
	inputA, inputB, inputC, addrV, inputV = 3, 4, 5, 1, 88
	expectPC, expectA, expectB, expectC, expectV = 1, inputA, inputB+inputV, inputC, inputV
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: addrV})
	m.SetWord(addrV, vm.Word{Op: op.CON, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}
	if valOfV := m.Core[addrV].Value; valOfV != expectV {
		t.Errorf("%s: *v: want %d: got %d\n", opc, inputV, expectV)
	}

	opc = op.ANDL
	inputA, inputB, inputC, inputV = 3, 4, 5, 11
	expectPC, expectA, expectB, expectC = 1, inputA&inputV, inputB, inputC
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}

	opc = op.ANDV
	inputA, inputB, inputC, addrV, inputV = 3, 4, 5, 1, 11
	expectPC, expectA, expectB, expectC, expectV = 1, inputA&inputV, inputB, inputC, inputV
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: addrV})
	m.SetWord(addrV, vm.Word{Op: op.CON, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}
	if valOfV := m.Core[addrV].Value; valOfV != expectV {
		t.Errorf("%s: *v: want %d: got %d\n", opc, inputV, expectV)
	}

	opc = op.BUMP
	inputA, inputB, inputC, addrV, inputV = 3, 4, 5, 1, 11
	expectPC, expectA, expectB, expectC, expectV = 1, inputA, inputB, inputC, inputV+inputV
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: addrV, ValueTwo: inputV})
	m.SetWord(addrV, vm.Word{Op: op.CON, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}
	if valOfV := m.Core[addrV].Value; valOfV != expectV {
		t.Errorf("%s: *v: want %d: got %d\n", opc, inputV, expectV)
	}

	opc = op.CLEAR
	inputA, inputB, inputC, addrV, inputV = 3, 4, 5, 1, 11
	expectPC, expectA, expectB, expectC, expectV = 1, inputA, inputB, inputC, 0
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: addrV})
	m.SetWord(addrV, vm.Word{Op: op.CON, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}
	if valOfV := m.Core[addrV].Value; valOfV != expectV {
		t.Errorf("%s: *v: want %d: got %d\n", opc, inputV, expectV)
	}

	opc = op.MULTL
	inputA, inputB, inputC, inputV = 5, 7, -4, 3
	expectPC, expectA, expectB, expectC = 1, inputA*inputV, inputB, inputC
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}

	opc = op.SAL
	inputA, inputB, inputC, inputV = 52, 17, 41, -12
	expectPC, expectA, expectB, expectC = 1, inputA-inputV, inputB, inputC
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}

	opc = op.SBL
	inputA, inputB, inputC, inputV = 9, 5, 13, 3
	expectPC, expectA, expectB, expectC = 1, inputA, inputB-inputV, inputC
	m = &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc, Value: inputV})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}
}

// TestVM_Misc tests things that don't fall into other categories.
func TestVM_Misc(t *testing.T) {
	opc := op.NOOP
	inputA, inputB, inputC := -42, 88, 32
	expectPC, expectA, expectB, expectC := 1, inputA, inputB, inputC
	m := &vm.VM{A: inputA, B: inputB, C: inputC}
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if m.A != expectA {
		t.Errorf("%s: r.A wants %d: got %d\n", opc, expectA, m.A)
	}
	if m.B != expectB {
		t.Errorf("%s: r.B wants %d: got %d\n", opc, expectB, m.B)
	}
	if m.C != expectC {
		t.Errorf("%s: r.C wants %d: got %d\n", opc, expectC, m.C)
	}
	if m.PC != expectPC {
		t.Errorf("%s: pc wants %d: got %d\n", opc, expectPC, m.PC)
	}
}

// TestVM_Streams tests stream instructions.
func TestVM_Streams(t *testing.T) {
	opc := op.MESS
	input := "$ABCDEFGHIJKLMNOPQRSTUVWXYZ$0123456789$.,;:()*/-+=\t \"$"
	expect := strings.ReplaceAll(input, "$", "\n")
	m, out := &vm.VM{}, &bytes.Buffer{}
	m.SetWord(0, vm.Word{Op: opc, Text: input})
	if err := m.Step(out, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	}
	if b := out.Bytes(); len(b) != len(expect) {
		t.Errorf("%s: out.len wants %d: got %d\n", opc, len(expect), len(b))
	} else if s := string(b); s != expect {
		t.Errorf("%s: out wants %q: got %q\n", opc, expect, s)
	}
}

// TestVM_MDxxx tests the custom MD instructions.
func TestVM_MDxxx(t *testing.T) {
	opc := op.MDERCH
	inputC := 'A'
	expect := string(inputC)
	m, out := &vm.VM{C: int(inputC)}, &bytes.Buffer{}
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(out, nil); err != nil {
		t.Errorf("%s: want nil: got %v\n", opc, err)
	} else {
		if b := out.Bytes(); len(b) != len(expect) {
			t.Errorf("%s: out.len wants %d: got %d\n", opc, len(expect), len(b))
		} else if s := string(b); s != expect {
			t.Errorf("%s: out wants %q: got %q\n", opc, expect, s)
		}
	}

	opc = op.MDQUIT
	m = &vm.VM{}
	m.SetWord(0, vm.Word{Op: opc})
	if expect := 0; m.PC != expect || m.Core[0].Op != opc {
		t.Fatalf("setw: want pc %d op %s: got %d %q\n", expect, opc, m.PC, m.Core[m.PC].Op)
	}
	if err := m.Run(nil, nil); err != nil {
		t.Fatalf("run: want ErrQuit: got %v\n", err)
	}
	if m.PC != 0 {
		t.Errorf("pc: want 0: got %d\n", m.PC)
	}
}
