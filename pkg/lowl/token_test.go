// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package lowl

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	type tokline struct {
		length int
		line   int
		kind   TokenKind
		value  string
		err    error
	}
	for _, tc := range []struct {
		program string
		tokens  []tokline
	}{
		{program: `
		NB		'These are LOWL statements'
[ENDST]	LBV		IDPT
		BMOVE
		SUBR	CHEKID,X,1
		MESS	'Error - - stack overflow'
`, tokens: []tokline{
			{length: 112, line: 1, kind: EOL, value: `*EOL*`, err: nil},
			{length: 108, line: 2, kind: TEXT, value: `NB`, err: nil},
			{length: 79, line: 2, kind: QTEXT, value: `"These are LOWL statements"`, err: nil},
			{length: 78, line: 2, kind: EOL, value: `*EOL*`, err: nil},
			{length: 71, line: 3, kind: LABEL, value: `[ENDST]`, err: nil},
			{length: 67, line: 3, kind: TEXT, value: `LBV`, err: nil},
			{length: 61, line: 3, kind: TEXT, value: `IDPT`, err: nil},
			{length: 60, line: 3, kind: EOL, value: `*EOL*`, err: nil},
			{length: 53, line: 4, kind: TEXT, value: `BMOVE`, err: nil},
			{length: 52, line: 4, kind: EOL, value: `*EOL*`, err: nil},
			{length: 46, line: 5, kind: TEXT, value: `SUBR`, err: nil},
			{length: 39, line: 5, kind: TEXT, value: `CHEKID`, err: nil},
			{length: 38, line: 5, kind: COMMA, value: `,`, err: nil},
			{length: 37, line: 5, kind: TEXT, value: `X`, err: nil},
			{length: 36, line: 5, kind: COMMA, value: `,`, err: nil},
			{length: 35, line: 5, kind: INTEGER, value: `1`, err: nil},
			{length: 34, line: 5, kind: EOL, value: `*EOL*`, err: nil},
			{length: 28, line: 6, kind: TEXT, value: `MESS`, err: nil},
			{length: 1, line: 6, kind: QTEXT, value: `"Error - - stack overflow"`, err: nil},
			{length: 0, line: 6, kind: EOL, value: `*EOL*`, err: nil},
			{length: 0, line: 7, kind: EOF, value: `*EOF*`, err: nil},
			{length: 0, line: 7, kind: EOF, value: `*EOF*`, err: nil},
		},
		}} {
		input := []byte(tc.program)
		for _, want := range tc.tokens {
			rest, tok, err := nextToken(input)
			if err != nil {
				if want.err == nil {
					t.Errorf("error: want nil: got %v", err)
					break
				}
			} else if tok.Error != nil {
				t.Errorf("error: token: %v", tok.Error)
			} else {
				if want.length != len(rest) {
					t.Errorf("length: want %d: got %d\n", want.length, len(rest))
				}
				if want.kind != tok.Kind {
					t.Errorf("kind: want %s: got %s\n", want.kind, tok.Kind)
				}
				if want.value != tok.String() {
					t.Errorf("value: want %s: got %s\n", want.value, tok.Kind.String())
				}
			}
			input = rest
		}
		if len(input) != 0 {
			t.Errorf("did not consume input: %q\n", string(input))
		}
	}
}
