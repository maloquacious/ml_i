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
	Cmp     vm.CMPRSLT
	Text    string
	V, V2   val_t
}
type expect_t struct {
	PC      int
	A, B, C int
	Cmp     vm.CMPRSLT
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
		if input.V2.address != 0 {
			m.SetWord(input.V2.address, vm.Word{Op: op.CON, Value: input.V2.value})
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
	testCmpResult := func() {
		if m.Registers.Cmp != expect.Cmp {
			t.Errorf("%s: cmp: want %q: got %q\n", opc, expect.Cmp, m.Registers.Cmp)
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
		testCmpResult()
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

	opc = op.ALIGN
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

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

	opc = op.CAI
	input = input_t{A: 7, B: 4, C: 5, V: val_t{1, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_LT}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)
	input = input_t{A: 8, B: 4, C: 5, V: val_t{1, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_EQ}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)
	input = input_t{A: 9, B: 4, C: 5, V: val_t{1, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_GR}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.CAL
	input = input_t{A: 7, B: 4, C: 5, V: val_t{0, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_LT}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)
	input = input_t{A: 8, B: 4, C: 5, V: val_t{0, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_EQ}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)
	input = input_t{A: 9, B: 4, C: 5, V: val_t{0, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_GR}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	opc = op.CAV
	input = input_t{A: 0, B: 4, C: 5, V: val_t{1, -5}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_LT}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)
	input = input_t{A: 1, B: 4, C: 5, V: val_t{1, 18}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_EQ}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)
	input = input_t{A: 2, B: 4, C: 5, V: val_t{1, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_GR}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.CCI
	input = input_t{A: 9, B: 4, C: 7, V: val_t{1, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_LT}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)
	input = input_t{A: 18, B: 4, C: 8, V: val_t{1, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_EQ}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)
	input = input_t{A: 3, B: 4, C: 9, V: val_t{1, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_GR}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.CCN
	input = input_t{A: 17, B: 4, C: 7, V: val_t{0, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_LT}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)
	input = input_t{A: 38, B: 4, C: 8, V: val_t{0, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_EQ}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)
	input = input_t{A: -99, B: 4, C: 9, V: val_t{0, 8}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: vm.IS_GR}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	t.Errorf("%s: not tested\n", op.CFSTK)

	opc = op.CLEAR
	input = input_t{A: 3, B: 12, C: 49, V: val_t{1, 11}}
	expect = expect_t{PC: 1, A: input.A & input.V.value, B: input.B, C: input.C, V: val_t{input.V.address, input.V.value}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	opc = op.CON
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	t.Errorf("%s: not tested\n", op.CSS)

	opc = op.DCL
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	opc = op.EQU
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

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

	opc = op.IDENT
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	opc = op.LAA
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.V.address, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.LAI
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}, V2: val_t{88, 837}}
	expect = expect_t{PC: 1, A: input.V2.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.LAL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{0, 88}}
	expect = expect_t{PC: 1, A: input.V.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	// LAM derive the pointer given by adding N-OF to the contents of B,
	//     and load A with the value pointed at by this (i.e. load A modified).
	opc = op.LAM
	input = input_t{A: 3, B: 84, C: 5, V: val_t{1, 837}}
	expect = expect_t{PC: 1, A: input.V.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: -83})
	test(nil, nil)

	opc = op.LAV
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.V.value, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.LBV
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A, B: input.V.value, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.LCI
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}, V2: val_t{88, 837}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.V2.value, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	// LCM derive the pointer given by adding N-OF to the contents of B,
	//     and load C with the value pointed at by this (i.e. load C modified).
	opc = op.LCM
	input = input_t{A: 3, B: 84, C: 5, V: val_t{1, 837}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.V.value, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: -83})
	test(nil, nil)

	opc = op.LCN
	input = input_t{A: 3, B: 4, C: 5, V: val_t{0, 88}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.V.value, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

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

	opc = op.MDLABEL
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

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

	opc = op.NB
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

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