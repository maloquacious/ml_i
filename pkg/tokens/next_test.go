package tokens_test

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/tokens"
	"testing"
)

func TestNext(t *testing.T) {
	type test_case struct {
		id    int
		input string
		want  []string
	}

	for _, tc := range []test_case{
		{id: 1, input: "[ALPHA]  STR     'ABCDEFGHIJKLMNOPQRSTUVWXYZ'\n",
			want: []string{"[ALPHA]", "STR", "'ABCDEFGHIJKLMNOPQRSTUVWXYZ'", "\n"}},
		{id: 2, input: "[NUMS]   STR     '0123456789' \n ",
			want: []string{"[NUMS]", "STR", "'0123456789'", "\n", ""}},
		{id: 3, input: "[DELIMS] STR     '.,;:()*/-+='",
			want: []string{"[DELIMS]", "STR", "'.,;:()*/-+='"}},
		{id: 4, input: "         EQU     KK,AA,6*-6",
			want: []string{"EQU", "KK", ",", "AA", ",", "6", "*", "-", "6"}},
		{id: 5, input: " [ABC] [D-F] (G) <H> 9+4*IJ(6) K-9L ",
			want: []string{"[ABC]", "[D", "-", "F", "]", "(", "G", ")", "<", "H", ">", "9", "+", "4", "*", "IJ", "(", "6", ")", "K", "-", "9L", ""}},
	} {
		var tok []byte
		rest := []byte(tc.input)
		for n, want := range tc.want {
			tok, rest = tokens.Next(rest)
			if want != string(tok) {
				t.Errorf("%d:%d: want %q: got %q\n", tc.id, n+1, want, string(tok))
			}
		}
		if len(rest) != 0 {
			t.Errorf("%d:%d: want nil: got %q\n", tc.id, len(tc.want)+1, string(rest))
		}
	}
}

func TestNextLine(t *testing.T) {
	type test_case struct {
		id    int
		input string
		wants [][]string
	}
	for _, tc := range []test_case{
		{id: 1, input: "\n \n  \n ",
			wants: [][]string{
				{"\n", ""},
				{"\n", ""},
				{"\n", ""}}},
		{id: 2, input: " [AB] 'foo' \n [C \n 'one' 'two",
			wants: [][]string{
				{"[AB]", "'foo'", "\n", ""},
				{"[C", "\n", ""},
				{"'one'", "'two", ""}}},
	} {
		toks, rest := tokens.NextLine([]byte(tc.input))
		for i, wants := range tc.wants {
			for n, want := range wants {
				var tok []byte
				if len(toks) != 0 {
					tok, toks = toks[0], toks[1:]
				}
				if want != string(tok) {
					fmt.Printf("%d:%d:%d: want %q: got %q\n", tc.id, i, n, want, string(tok))
				}
			}
			toks, rest = tokens.NextLine(rest)
		}
		want, got := "", []string{}
		for _, tok := range toks {
			got = append(got, string(tok))
		}
		if len(got) != 0 {
			t.Errorf("%d:toks: want %q: got %v\n", tc.id, want, got)
		}
		if len(rest) != 0 {
			t.Errorf("%d:rest: want %q: got %q\n", tc.id, want, string(rest))
		}
	}
}
