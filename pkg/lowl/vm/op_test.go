// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm_test

import (
	"bytes"
	"errors"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"github.com/maloquacious/ml_i/pkg/lowl/vm"
	"io"
	"strings"
	"testing"
)

type input_t struct {
	PC      int
	A, B, C int
	Text    string
	V, V2   val_t
}
type expect_t struct {
	PC      int
	A, B, C int
	Text    string
	V, V2   val_t
}

type val_t struct {
	address int
	value   int
}

// TestVM tests running, stepping, and all the op codes in the machine.
func TestVM(t *testing.T) {
	var m *vm.VM
	var opc op.Code
	var input input_t
	var expect expect_t
	var out *bytes.Buffer

	newvm := func() {
		m = &vm.VM{PC: input.PC, A: input.A, B: input.B, C: input.C}
		if input.V.address != 0 {
			m.SetWord(input.V.address, vm.Word{Op: op.CON, Value: input.V.value})
		}
	}
	step := func(stdout, stdmsg io.Writer) {
		if err := m.Step(stdout, stdmsg); err != nil {
			t.Errorf("%s: want nil: got %v\n", opc, err)
		}
	}
	testA := func() {
		if m.A != expect.A {
			t.Errorf("%s: r.A: want %d: got %d\n", opc, expect.A, m.A)
		}
	}
	testB := func() {
		if m.B != expect.B {
			t.Errorf("%s: r.B: want %d: got %d\n", opc, expect.B, m.B)
		}
	}
	testC := func() {
		if m.C != expect.C {
			t.Errorf("%s: r.C: want %d: got %d\n", opc, expect.C, m.C)
		}
	}
	testPC := func() {
		if m.PC != expect.PC {
			t.Errorf("%s: pc: want %d: got %d\n", opc, expect.PC, m.PC)
		}
	}
	testV := func() {
		if expect.V.address != 0 {
			valOfV := m.Core[expect.V.address].Value
			if valOfV != expect.V.value {
				t.Errorf("%s: *v: want %d: got %d\n", opc, expect.V.value, valOfV)
			}
		}
	}
	test := func(stdout, stdmsg io.Writer) {
		step(stdout, stdmsg)
		testA()
		testB()
		testC()
		testPC()
		testV()
	}

	// halting test must pass before running further tests
	opc = op.HALT
	input = input_t{}
	expect = expect_t{}
	newvm()
	if m.PC != expect.PC || m.Core[0].Op != opc {
		t.Fatalf("setw: want pc %d op HALT: got %d %q\n", expect.PC, m.PC, m.Core[m.PC].Op)
	}
	if err := m.Run(nil, nil); err != nil {
		if !errors.Is(err, vm.ErrHalted) {
			t.Fatalf("run: want ErrHalt: got %v\n", err)
		}
	}
	if m.PC != expect.PC {
		t.Fatalf("pc: wants %d: got %d\n", expect.PC, m.PC)
	}

	opc = op.AAL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A + input.V.value, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	opc = op.AAV
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A + input.V.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.ABV
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A, B: input.B + input.V.value, C: input.C, V: val_t{1, input.V.value}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.ALIGN)

	opc = op.ANDL
	input = input_t{A: 3, B: 12, C: 49, V: val_t{1, 11}}
	expect = expect_t{PC: 1, A: input.A & input.V.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	opc = op.ANDV
	input = input_t{A: 3, B: 12, C: 49, V: val_t{1, 11}}
	expect = expect_t{PC: 1, A: input.A & input.V.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.BMOVE)
	t.Errorf("%s: not tested\n", op.BSTK)

	opc = op.BUMP
	input = input_t{A: 3, B: 12, C: 49, V: val_t{1, 11}, V2: val_t{value: 2}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: val_t{input.V.address, input.V.value + input.V2.value}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address, ValueTwo: input.V2.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.CAI)
	t.Errorf("%s: not tested\n", op.CAL)
	t.Errorf("%s: not tested\n", op.CAV)
	t.Errorf("%s: not tested\n", op.CCI)
	t.Errorf("%s: not tested\n", op.CCN)
	t.Errorf("%s: not tested\n", op.CFSTK)

	opc = op.CLEAR
	input = input_t{A: 3, B: 12, C: 49, V: val_t{1, 11}}
	expect = expect_t{PC: 1, A: input.A & input.V.value, B: input.B, C: input.C, V: val_t{input.V.address, input.V.value}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.CON)
	t.Errorf("%s: not tested\n", op.CSS)
	t.Errorf("%s: not tested\n", op.DCL)
	t.Errorf("%s: not tested\n", op.EQU)
	t.Errorf("%s: not tested\n", op.EXIT)
	t.Errorf("%s: not tested\n", op.FMOVE)
	t.Errorf("%s: not tested\n", op.FSTK)
	t.Errorf("%s: not tested\n", op.GO)
	t.Errorf("%s: not tested\n", op.GOADD)
	t.Errorf("%s: not tested\n", op.GOEQ)
	t.Errorf("%s: not tested\n", op.GOGE)
	t.Errorf("%s: not tested\n", op.GOGR)
	t.Errorf("%s: not tested\n", op.GOLE)
	t.Errorf("%s: not tested\n", op.GOND)
	t.Errorf("%s: not tested\n", op.GONE)
	t.Errorf("%s: not tested\n", op.GOPC)
	t.Errorf("%s: not tested\n", op.GOSUB)
	t.Errorf("%s: not tested\n", op.GOTBL)

	opc = op.HALT
	input = input_t{A: 3, B: 4, C: 5}
	expect = expect_t{PC: 0, A: input.A, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want halted: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrHalted) {
		t.Errorf("%s: want halted: got %v\n", opc, err)
	}
	if !m.Registers.Halted {
		t.Errorf("%s: want halted: got running\n", opc)
	}
	testA()
	testB()
	testC()
	testPC()
	testV()

	t.Errorf("%s: not tested\n", op.IDENT)
	t.Errorf("%s: not tested\n", op.LAA)
	t.Errorf("%s: not tested\n", op.LAI)

	opc = op.LAL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{0, 88}}
	expect = expect_t{PC: 1, A: input.V.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.LAM)
	t.Errorf("%s: not tested\n", op.LAV)
	t.Errorf("%s: not tested\n", op.LBV)
	t.Errorf("%s: not tested\n", op.LCI)
	t.Errorf("%s: not tested\n", op.LCM)
	t.Errorf("%s: not tested\n", op.LCN)

	opc = op.MDERCH
	input = input_t{A: 3, B: 4, C: 'A'}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, Text: "A"}
	out = &bytes.Buffer{}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	test(out, nil)
	if b := out.Bytes(); len(b) != len(expect.Text) {
		t.Errorf("%s: out.len: want %d: got %d\n", opc, len(expect.Text), len(b))
	} else if s := string(b); s != expect.Text {
		t.Errorf("%s: out.text: want %q: got %q\n", opc, expect, s)
	}
	out = nil

	t.Errorf("%s: not tested\n", op.MDLABEL)

	opc = op.MDQUIT
	input = input_t{A: 3, B: 4, C: 5}
	expect = expect_t{PC: 0, A: input.A, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want quit: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrQuit) {
		t.Errorf("%s: want quit: got %v\n", opc, err)
	}
	if !m.Registers.Halted {
		t.Errorf("%s: want halted: got running\n", opc)
	}
	testA()
	testB()
	testC()
	testPC()
	testV()

	opc = op.MESS
	input = input_t{A: 3, B: 4, C: 5, Text: "$ABCDEFGHIJKLMNOPQRSTUVWXYZ$0123456789$.,;:()*/-+=\t \"$"}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, Text: strings.ReplaceAll(input.Text, "$", "\n")}
	out = &bytes.Buffer{}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Text: input.Text})
	test(out, nil)
	if b := out.Bytes(); len(b) != len(expect.Text) {
		t.Errorf("%s: out.len: want %d: got %d\n", opc, len(expect.Text), len(b))
	} else if s := string(b); s != expect.Text {
		t.Errorf("%s: out.text: want %q: got %q\n", opc, expect, s)
	}
	out = nil

	opc = op.MULTL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A * input.V.value, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.NB)
	t.Errorf("%s: not tested\n", op.NCH)

	opc = op.NOOP
	input = input_t{A: 3, B: 4, C: 5}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.PRGEN)
	t.Errorf("%s: not tested\n", op.PRGST)

	opc = op.SAL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A - input.V.value, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.SAV)

	opc = op.SBL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A, B: input.B - input.V.value, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.SBV)
	t.Errorf("%s: not tested\n", op.STI)
	t.Errorf("%s: not tested\n", op.STR)
	t.Errorf("%s: not tested\n", op.STV)
	t.Errorf("%s: not tested\n", op.SUBR)
	t.Errorf("%s: not tested\n", op.UNSTK)
}
