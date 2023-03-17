// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package assembler

import (
	"fmt"
)

// A symbol table is a data structure used by an assembler to keep track of symbolic information, such as the names and values of labels and variables in the source code. The specific functions available on a symbol table may vary depending on the assembler implementation, but some common functions include:

func newSymbolTable() *symbolTable {
	return &symbolTable{symbols: make(map[string]*symbolNode)}
}

type symbolTable struct {
	symbols map[string]*symbolNode
}

type symbolNode struct {
	name string // name of the symbol
	kind string // kind of the symbol
	line int    // set when the symbol is defined
	// value of the symbol
	address  int
	alias    string
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
		sym = &symbolNode{
			name: name,
			kind: "undefined",
		}
		st.symbols[name] = sym
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
func (st *symbolTable) InsertAddress(line int, name string, address int) bool {
	if sym, ok := st.symbols[name]; ok {
		if sym.kind != "undefined" {
			return false
		}
		sym.kind, sym.address, sym.line = "address", address, line
		return true
	}
	st.symbols[name] = &symbolNode{
		name:    name,
		kind:    "address",
		address: address,
		line:    line,
	}
	return true
}

// InsertAlias adds a new symbol to the symbol table with its name and value.
func (st *symbolTable) InsertAlias(line int, name string, text string) bool {
	if sym, ok := st.symbols[name]; ok {
		return false
	} else if sym, ok = st.symbols[text]; ok && sym.kind == "alias" {
		panic(fmt.Sprintf("alias %q references alias %q", name, text))
	}
	st.symbols[name] = &symbolNode{
		name:  name,
		kind:  "alias",
		alias: text,
		line:  line,
	}
	return true
}

// InsertConstant adds a new symbol to the symbol table with its name and value.
func (st *symbolTable) InsertConstant(line int, name string, number int) bool {
	if _, ok := st.symbols[name]; ok {
		return false
	}
	st.symbols[name] = &symbolNode{
		name:     name,
		kind:     "constant",
		constant: number,
		line:     line,
	}
	return true
}

// InsertLiteral adds a new symbol to the symbol table with its name and value.
func (st *symbolTable) InsertLiteral(line int, name string, text string) bool {
	if _, ok := st.symbols[name]; ok {
		return false
	}
	st.symbols[name] = &symbolNode{
		name:    name,
		kind:    "literal",
		literal: text,
		line:    line,
	}
	return true
}

// Lookup searches the symbol table for a specific symbol by name and returns its value.
func (st *symbolTable) Lookup(name string) (*symbolNode, bool) {
	sym, ok := st.symbols[name]
	if ok && sym.kind == "alias" {
		sym, ok = st.symbols[sym.alias]
	}
	if ok && sym.kind == "undefined" {
		return nil, false
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
