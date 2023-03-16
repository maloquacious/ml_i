// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package lowl

import "fmt"

type command struct {
	line  int
	label *Token
	op    *Token
	args  []*Token
	err   error
}

// String implements the Stringer interface
func (c command) String() string {
	var s string
	if c.label != nil {
		s = fmt.Sprintf("%s\t", c.label)
	} else {
		s = "\t"
	}
	if c.op != nil {
		s += fmt.Sprintf("%s\t", c.op)
	} else {
		s += "\t"
	}
	for _, arg := range c.args {
		s += fmt.Sprintf("%s\t", arg)
	}
	//if c.line != 0 {
	//	s += fmt.Sprintf("\t\t\t; %d", c.line)
	//}
	return s
}
