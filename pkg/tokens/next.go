// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package tokens

import (
	"bytes"
)

// define delimiters and whitespace.
// note that newlines are delimiters, not whitespace!
var delims = []byte(",+-*/:<>[]()=\n")
var delimsQT = []byte("'\n")
var ws = []byte{' ', '\t', '\r'}

// Next skips whitespace and then returns the next token
// (a delimiter, quoted text, or word) from the input.
// Newlines aren't considered whitespace for us.
func Next(b []byte) (tok []byte, rest []byte) {
	// skip leading whitespace (but not newlines)
	for len(b) != 0 && bytes.IndexByte(ws, b[0]) != -1 {
		b = b[1:]
	}

	// test for end of input
	if len(b) == 0 {
		return nil, nil
	}

	// initialize the token and rest of the buffer
	tok, rest = []byte{b[0]}, b[1:]

	// return delimiters
	if bytes.IndexByte(delims, tok[0]) != -1 {
		if tok[0] == '[' {
			// token should be everything up to the right bracket,
			// but should never include any delimiter or white space.
			for len(rest) != 0 && bytes.IndexByte(ws, rest[0]) == -1 && bytes.IndexByte(delims, rest[0]) == -1 && bytes.IndexByte(delimsQT, rest[0]) == -1 {
				// append the character and advance the buffer
				tok, rest = append(tok, rest[0]), rest[1:]
			}
			if len(rest) != 0 && rest[0] == ']' {
				tok, rest = append(tok, rest[0]), rest[1:]
			}
		} else {
			// token is just the delimiter
		}
		return tok, rest
	}

	if tok[0] == '\'' { // and quoted text
		// delimiters for quoted text are the quote or a newline
		for len(rest) != 0 && bytes.IndexByte(delimsQT, rest[0]) == -1 {
			// append the character and advance the buffer
			tok, rest = append(tok, rest[0]), rest[1:]
		}
		// include the closing quote, if there was one
		if len(rest) != 0 && rest[0] == tok[0] {
			// append the character and advance the buffer
			tok, rest = append(tok, rest[0]), rest[1:]
		}
	} else { // token is everything up to whitespace or a delimiter
		for len(rest) != 0 && bytes.IndexByte(ws, rest[0]) == -1 && bytes.IndexByte(delims, rest[0]) == -1 {
			// append the character and advance the buffer
			tok, rest = append(tok, rest[0]), rest[1:]
		}
	}

	return tok, rest
}

// NextLine returns all the tokens on the current line.
// The input buffer is advanced to the start of the next line.
func NextLine(b []byte) (toks [][]byte, rest []byte) {
	tok, rest := Next(b)
	for len(tok) != 0 {
		toks = append(toks, tok)
		if tok[0] == '\n' {
			break
		}
		tok, rest = Next(rest)
	}
	return toks, rest
}
