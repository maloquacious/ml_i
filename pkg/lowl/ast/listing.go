/*
 * Copyright (c) 2023 Michael D Henderson. All rights reserved.
 */

package ast

import (
	"bytes"
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
)

func (nodes Nodes) Listing() []byte {
	bb := &bytes.Buffer{}
	for _, node := range nodes {
		blabel := ""
		if node.Op == op.MDLABEL {
			blabel = node.Parameters.String() + ":"
		}
		_, _ = fmt.Fprintf(bb, "%-12s %-12s %-55s ;; %4d\n", blabel, node.Op, node.Parameters.String(), node.Line)
	}
	return bb.Bytes()
}
