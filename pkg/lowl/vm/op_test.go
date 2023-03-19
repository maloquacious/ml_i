// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm_test

import (
	"bytes"
	"errors"
	"fmt"
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
	RS      []int
	Text    string
	V, V2   val_t
}
type expect_t struct {
	PC      int
	A, B, C int
	Cmp     vm.CMPRSLT
	RS      []int
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
		m.Registers.Cmp = input.Cmp
		m.RS = append(m.RS, input.RS...)
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
	testRS := func() {
		want := fmt.Sprintf("%v", expect.RS)
		got := fmt.Sprintf("%v", m.RS)
		if got != want {
			t.Errorf("%s: rs: want %s: got %s\n", opc, want, got)
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
		testRS()
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

	opc = op.BMOVE
	t.Errorf("%s: not tested\n", opc)

	opc = op.BSTK
	t.Errorf("%s: not tested\n", opc)

	opc = op.BUMP
	input = input_t{A: 3, B: 12, C: 49, V: val_t{1, 11}, V2: val_t{value: 2}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: val_t{input.V.address, input.V.value + input.V2.value}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address, ValueTwo: input.V2.value})
	test(nil, nil)

	opc = op.CAI
	for _, tc := range []struct {
		a, i int
		cmp  vm.CMPRSLT
	}{
		{23, 29, vm.IS_LT},
		{29, 29, vm.IS_EQ},
		{31, 29, vm.IS_GR},
	} {
		input = input_t{A: tc.a, B: 4, C: 5, V: val_t{1, 8}, V2: val_t{8, tc.i}}
		expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, V2: input.V2, Cmp: tc.cmp}
		newvm()
		m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
		test(nil, nil)
	}

	opc = op.CAL
	for _, tc := range []struct {
		a, i int
		cmp  vm.CMPRSLT
	}{
		{23, 29, vm.IS_LT},
		{29, 29, vm.IS_EQ},
		{31, 29, vm.IS_GR},
	} {
		input = input_t{A: tc.a, B: 4, C: 5, V: val_t{1, tc.i}}
		expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: tc.cmp}
		newvm()
		m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
		test(nil, nil)
	}

	opc = op.CAV
	for _, tc := range []struct {
		a, i int
		cmp  vm.CMPRSLT
	}{
		{23, 29, vm.IS_LT},
		{29, 29, vm.IS_EQ},
		{31, 29, vm.IS_GR},
	} {
		input = input_t{A: tc.a, B: 4, C: 5, V: val_t{1, tc.i}}
		expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: tc.cmp}
		newvm()
		m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
		test(nil, nil)
	}

	opc = op.CCI
	for _, tc := range []struct {
		c, i int
		cmp  vm.CMPRSLT
	}{
		{23, 29, vm.IS_LT},
		{29, 29, vm.IS_EQ},
		{31, 29, vm.IS_GR},
	} {
		input = input_t{A: 13, B: 4, C: tc.c, V: val_t{1, 8}, V2: val_t{8, tc.i}}
		expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, V2: input.V2, Cmp: tc.cmp}
		newvm()
		m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
		test(nil, nil)
	}

	opc = op.CCL
	for _, tc := range []struct {
		c, i int
		cmp  vm.CMPRSLT
	}{
		{23, 29, vm.IS_LT},
		{29, 29, vm.IS_EQ},
		{31, 29, vm.IS_GR},
	} {
		input = input_t{A: 15, B: 4, C: tc.c, V: val_t{1, tc.i}}
		expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: tc.cmp}
		newvm()
		m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
		test(nil, nil)
	}

	opc = op.CCN
	for _, tc := range []struct {
		c, i int
		cmp  vm.CMPRSLT
	}{
		{23, 29, vm.IS_LT},
		{29, 29, vm.IS_EQ},
		{31, 29, vm.IS_GR},
	} {
		input = input_t{A: 15, B: 4, C: tc.c, V: val_t{1, tc.i}}
		expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, Cmp: tc.cmp}
		newvm()
		m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
		test(nil, nil)
	}

	opc = op.CFSTK
	t.Errorf("%s: not tested\n", opc)

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

	opc = op.CSS
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want underflow: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrStackUnderflow) {
		t.Errorf("%s: want underflow: got %v\n", opc, err)
	}
	input = input_t{RS: []int{99}}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	test(nil, nil)

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

	opc = op.EXIT
	t.Errorf("%s: not tested\n", opc)
	opc = op.FMOVE
	t.Errorf("%s: not tested\n", opc)
	opc = op.FSTK
	t.Errorf("%s: not tested\n", opc)

	opc = op.GO
	input = input_t{Cmp: vm.IS_LT}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_EQ}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_GR}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GOADD
	t.Errorf("%s: not tested\n", opc)
	//input = input_t{}
	//expect = expect_t{PC: 8}
	//newvm()
	//m.SetWord(0, vm.Word{Op: opc, Value: 8})
	//test(nil, nil)

	opc = op.GOEQ
	input = input_t{Cmp: vm.IS_LT}
	expect = expect_t{PC: 1, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_EQ}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_GR}
	expect = expect_t{PC: 1, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GOGE
	input = input_t{Cmp: vm.IS_LT}
	expect = expect_t{PC: 1, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_EQ}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_GR}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GOGR
	input = input_t{Cmp: vm.IS_LT}
	expect = expect_t{PC: 1, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_EQ}
	expect = expect_t{PC: 1, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_GR}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GOLE
	input = input_t{Cmp: vm.IS_LT}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_EQ}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_GR}
	expect = expect_t{PC: 1, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GOND
	input = input_t{C: '0'}
	expect = expect_t{PC: 1, A: 0, C: '0'}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{C: '9'}
	expect = expect_t{PC: 1, A: 9, C: '9'}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{C: 'A'}
	expect = expect_t{PC: 8, C: 'A'}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GONE
	input = input_t{Cmp: vm.IS_LT}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_EQ}
	expect = expect_t{PC: 1, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{Cmp: vm.IS_GR}
	expect = expect_t{PC: 8, Cmp: input.Cmp}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GOPC
	input = input_t{C: '0'}
	expect = expect_t{PC: 1, C: '0'}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{C: 'A'}
	expect = expect_t{PC: 1, C: 'A'}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)
	input = input_t{C: '$'}
	expect = expect_t{PC: 8, C: '$'}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: 8})
	test(nil, nil)

	opc = op.GOSUB
	t.Errorf("%s: not tested\n", opc)
	opc = op.GOTBL
	t.Errorf("%s: not tested\n", opc)

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

	opc = op.NCH
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	opc = op.NOOP
	input = input_t{A: 3, B: 4, C: 5}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	opc = op.PRGEN
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	opc = op.PRGST
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	opc = op.SAL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A - input.V.value, B: input.B, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	opc = op.SAV
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A - input.V.value, B: input.B, C: input.C, V: input.V}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.SBL
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A, B: input.B - input.V.value, C: input.C}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.value})
	test(nil, nil)

	opc = op.SBV
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A, B: input.B - input.V.value, C: input.C, V: val_t{1, input.V.value}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.STI
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 21}, V2: val_t{21, 144}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: input.V, V2: val_t{input.V2.address, input.A}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.STR
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	opc = op.STV
	input = input_t{A: 3, B: 4, C: 5, V: val_t{1, 88}}
	expect = expect_t{PC: 1, A: input.A, B: input.B, C: input.C, V: val_t{1, input.A}}
	newvm()
	m.SetWord(0, vm.Word{Op: opc, Value: input.V.address})
	test(nil, nil)

	opc = op.SUBR
	input = input_t{}
	expect = expect_t{PC: 1}
	newvm()
	m.SetWord(0, vm.Word{Op: opc})
	if err := m.Step(nil, nil); err == nil {
		t.Errorf("%s: want invalid op: got nil\n", opc)
	} else if !errors.Is(err, vm.ErrInvalidOp) {
		t.Errorf("%s: want invalid op: got %v\n", opc, err)
	}

	opc = op.UNSTK
	t.Errorf("%s: not tested\n", opc)
}
