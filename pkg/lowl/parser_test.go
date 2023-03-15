/*
 * Copyright (c) 2023 Michael D Henderson. All rights reserved.
 */

package lowl_test

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl"
	"testing"
)

func TestParseExpr(t *testing.T) {
	for _, tc := range []struct {
		input  string
		expect string
	}{
		{"(10*LCH)", "[( 10 * LCH )]"},
		{"(2*LCH)", "[( 2 * LCH )]"},
		{"(2*LCH+LNM)", "[( 2 * LCH + LNM )]"},
		{"(2*LNM)", "[( 2 * LNM )]"},
		{"(2*LNM-LCH)", "[( 2 * LNM - LCH )]"},
		{"(3*LNM)", "[( 3 * LNM )]"},
		{"(3*LNM+LCH)", "[( 3 * LNM + LCH )]"},
		{"(9*LCH)", "[( 9 * LCH )]"},
		{"(LCH)", "[( LCH )]"},
		{"(LCH+LCH)", "[( LCH + LCH )]"},
		{"(LCH-LCH)", "[( LCH - LCH )]"},
		{"(LNM)", "[( LNM )]"},
		{"(LNM+LCH)", "[( LNM + LCH )]"},
		{"(LNM-LNM)", "[( LNM - LNM )]"},
	} {
		got := lowl.ParseOF(tc.input)
		if tc.expect != fmt.Sprintf("%v", got) {
			t.Errorf("input %q: want %q: got %v\n", tc.input, tc.expect, got)
		}
	}
}
