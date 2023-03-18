// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package vm

import (
	"fmt"
	"io"
)

// directLoad returns the value of variable v
func (m *VM) directLoad(v int) int {
	return m.Core[v].Value
}

// directStore saves the value into variable v
func (m *VM) directStore(v, value int) {
	m.Core[v].Value = value
}

// indexedLoad returns the contents of the address pointed to by B + n
func (m *VM) indexedLoad(n int) int {
	return m.Core[m.B+n].Value
}

// indirectLoad returns the contents of the address pointed to by V
func (m *VM) indirectLoad(v int) int {
	return m.Core[m.Core[v].Value].Value
}

// indirectStore saves the value into the address pointed to by v
func (m *VM) indirectStore(v, value int) {
	m.Core[m.Core[v].Value].Value = value
}

func printf(w io.Writer, format string, args ...any) {
	if w != nil {
		_, _ = fmt.Fprintf(w, format, args...)
	}
}

func isalpha(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}
func isdigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
func ispunct(ch byte) bool {
	return !(isalpha(ch) || isdigit(ch))
}
