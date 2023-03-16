// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package lowl

import (
	"bytes"
)

// token returns the next delimiter, quoted text, or word from the input.
// it always skips whitespace (newlines aren't whitespace for us).
func token(b []byte) (tok []byte, rest []byte) {
	// define delimiters and whitespace.
	// note that newlines are delimiters, not whitespace!
	delims := []byte(",+-*/:<>[]()=\n")
	ws := []byte{' ', '\t', '\r'}

	// skip leading whitespace (but not newlines)
	for len(b) != 0 && bytes.IndexByte(ws, b[0]) != -1 {
		b = b[1:]
	}

	// test for end of input
	if len(b) == 0 {
		return nil, nil
	}

	// initialize the token and advance the buffer
	tok, b = []byte{b[0]}, b[1:]

	// return delimiters
	if bytes.IndexByte(delims, tok[0]) != -1 {
		// token is just the delimiter
	} else if tok[0] == '\'' { // and quoted text
		// delimiters for quoted text are the quote or a newline
		delims = []byte{tok[0], '\n'}
		for len(b) != 0 && bytes.IndexByte(delims, b[0]) == -1 {
			// append the character and advance the buffer
			tok, b = append(tok, b[0]), b[1:]
		}
		// include the closing quote, if there was one
		if len(b) != 0 && b[0] == tok[0] {
			// append the character and advance the buffer
			tok, b = append(tok, b[0]), b[1:]
		}
	} else { // token is everything up to whitespace or a delimiter
		for len(b) != 0 && bytes.IndexByte(ws, b[0]) == -1 && bytes.IndexByte(delims, b[0]) == -1 {
			// append the character and advance the buffer
			tok, b = append(tok, b[0]), b[1:]
		}
	}

	return tok, b
}
