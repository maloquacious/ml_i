// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package assembler

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/ast"
)

// A symbol table is a data structure used by an assembler to keep track of symbolic information, such as the names and values of labels and variables in the source code. The specific functions available on a symbol table may vary depending on the assembler implementation, but some common functions include:

func newSymbolTable(nodes ast.Nodes) *symbolTable {
	st := &symbolTable{symbols: make(map[string]*symbolNode)}
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

type symbolTable struct {
	symbols map[string]*symbolNode
}

type symbolNode struct {
	name string // name of the symbol
	kind string // kind of the symbol
	// value of the symbol
	address  int
	constant int
	literal  string
	// back-fill queue
	backFill []int
}

// AddReference adds a new address to the symbol's back-fill list.
// If the symbol does not exist, create it with the type "undefined."
func (st *symbolTable) AddReference(name string, address int) {
	sym, ok := st.symbols[name]
	if !ok {
		sym = &symbolNode{name: name, kind: "undefined"}
	}
	sym.backFill = append(sym.backFill, address)
}

// GetEnv returns an environment variable table
func (st *symbolTable) GetEnv() map[string]int {
	env := make(map[string]int)
	for _, sym := range st.symbols {
		if sym.kind == "constant" {
			env[sym.name] = sym.constant
		}
	}
	return env
}

// InsertAddress adds a new symbol to the symbol table with its name and value.
func (st *symbolTable) InsertAddress(name string, address int) bool {
	if _, ok := st.symbols[name]; ok {
		return false
	}
	st.symbols[name] = &symbolNode{name: name, kind: "address", address: address}
	return true
}

// InsertAlias adds a new symbol to the symbol table with its name and value.
func (st *symbolTable) InsertAlias(name string, text string) bool {
	if sym, ok := st.symbols[name]; ok {
		return false
	} else if sym, ok = st.symbols[text]; ok && sym.kind == "alias" {
		panic(fmt.Sprintf("alias %q references alias %q", name, text))
	}
	st.symbols[name] = &symbolNode{name: name, kind: "alias", literal: text}
	return true
}

// InsertConstant adds a new symbol to the symbol table with its name and value.
func (st *symbolTable) InsertConstant(name string, number int) bool {
	if _, ok := st.symbols[name]; ok {
		return false
	}
	st.symbols[name] = &symbolNode{name: name, kind: "constant", constant: number}
	return true
}

// InsertLiteral adds a new symbol to the symbol table with its name and value.
func (st *symbolTable) InsertLiteral(name string, text string) bool {
	if _, ok := st.symbols[name]; ok {
		return false
	}
	st.symbols[name] = &symbolNode{name: name, kind: "literal", literal: text}
	return true
}

// Lookup searches the symbol table for a specific symbol by name and returns its value.
func (st *symbolTable) Lookup(name string) (*symbolNode, bool) {
	sym, ok := st.symbols[name]
	if ok && sym.kind == "alias" {
		sym, ok = st.symbols[sym.literal]
	}
	return sym, ok
}

// UpdateAddress changes the value of a symbol that is already in the symbol table.
func (st *symbolTable) UpdateAddress(name string, address int) {
	sym, ok := st.symbols[name]
	if !ok {
		sym = &symbolNode{name: name, kind: "address"}
	}
	sym.address = address
}

// UpdateConstant changes the value of a symbol that is already in the symbol table.
func (st *symbolTable) UpdateConstant(name string, number int) {
	panic("should never call symbolTable.UpdateConstant")
}

// UpdateLiteral changes the value of a symbol that is already in the symbol table.
func (st *symbolTable) UpdateLiteral(name string, number int) {
	panic("should never call symbolTable.UpdateLiteral")
}

//Deletion: This function removes a symbol from the symbol table.

//Scoping: This function keeps track of the current scope of the code and resolves symbol conflicts by giving priority to symbols defined in the innermost scope.

//Error reporting: This function detects and reports errors such as duplicate symbol definitions, undefined symbols, and incorrect symbol usage.

//Memory allocation: This function calculates the memory locations where symbols will be stored in the final executable code.
