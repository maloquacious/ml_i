/*
 * Copyright (c) 2023 Michael D Henderson. All rights reserved.
 */

// Package assembler assembles the instructions and returns a VM that can run them.
package assembler

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/ast"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"github.com/maloquacious/ml_i/pkg/lowl/vm"
)

func Assemble(nodes ast.Nodes) (*vm.VM, error) {
	// create symbol table
	symtab := newSymbolTable(nodes)

	machine := &vm.VM{
		PC: 1, // leave address 0 for the initial jump to code
	}

	// allocate all storage up front
	for _, node := range nodes {
		if node.Op != op.DCL {
			continue
		}
		fmt.Printf("nst: declare var    %s\n", node.Parameters[0].Text)
		sym := &symbolNode{name: node.Parameters[0].Text, kind: "variable", address: machine.PC}
		symtab[sym.name] = sym
		machine.Core[machine.PC], machine.PC = vm.Word{Op: op.GO}, machine.PC+1
	}
	fmt.Printf("asm: %8d words created\n", machine.PC)
	for _, sym := range symtab {
		fmt.Printf("var %-12s address %8d\n", sym.name, sym.address)
	}

	machine.Core[0] = vm.Word{Op: op.GO, Address: machine.PC}
	machine.Core[machine.PC], machine.PC = vm.Word{Op: op.NOOP}, machine.PC+1

	panic("ast.Assemble is not implemented!")
}

func newSymbolTable(nodes ast.Nodes) symbolTable {
	st := make(map[string]*symbolNode)
	//for _, node := range nodes {
	//	var sym *symbolNode
	//	switch node.Op {
	//	case op.CON:
	//		fmt.Printf("nst: define  const %-12s %6d\n", node.Parameters[0].Text, node.Parameters[1].Number)
	//		sym = &symbolNode{name: node.Parameters[0].Text, kind: "const", number: node.Parameters[1].Number}
	//	case op.DCL:
	//		fmt.Printf("nst: declare var    %s\n", node.Parameters[0].Text)
	//		sym = &symbolNode{name: node.Parameters[0].Text, kind: "variable", address: 0}
	//	case op.EQU:
	//		fmt.Printf("nst: define  alias  %s\n", node.Parameters[0].Text)
	//		sym = &symbolNode{name: node.Parameters[0].Text, kind: "variable", text: node.Parameters[1].Text}
	//	case op.MDLABEL:
	//		fmt.Printf("nst: define  label  %s\n", node.Parameters[0].Text)
	//		sym = &symbolNode{name: node.Parameters[0].Text, kind: "label", address: 0}
	//	case op.SUBR:
	//		fmt.Printf("nst: declare sub    %s\n", node.Parameters[0].Text)
	//		sym = &symbolNode{name: node.Parameters[0].Text, kind: "subroutine", address: 0}
	//	}
	//	if sym != nil {
	//		st[sym.name] = sym
	//	}
	//}
	return st
}

type symbolTable map[string]*symbolNode

type symbolNode struct {
	name string // name of the symbol
	kind string // kind of the symbol
	// value of the symbol
	address int
	number  int
	text    string
}
