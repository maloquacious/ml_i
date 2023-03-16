// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

package lowl

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

func Parse(name string) ([]*command, error) {
	input, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	commands, line := []*command{}, 1
	for len(input) != 0 {
		rest, cmd := nextCommand(input)
		input = rest

		cmd.line, line = line, line+1
		if cmd.err != nil {
			return nil, fmt.Errorf("parse: %d: %w", cmd.line, err)
		} else if cmd.label == nil && cmd.op == nil && len(cmd.args) == 0 {
			// empty line
		} else {
			commands = append(commands, &cmd)
		}
	}

	return commands, nil
}

func nextCommand(input []byte) ([]byte, command) {
	var c command
	for len(input) != 0 {
		rest, tok, err := nextToken(input)
		input = rest

		if err != nil {
			c.err = fmt.Errorf("%d: %w", c.line, err)
			break
		} else if tok.Error != nil {
			c.err = fmt.Errorf("%d: %w", c.line, err)
			break
		} else if tok.Kind == EOL {
			break
		}

		if tok.Kind == LABEL {
			if c.label != nil {
				c.err = fmt.Errorf("%d: label: multiple", c.line)
				break
			} else if c.op != nil {
				c.err = fmt.Errorf("%d: label: after op", c.line)
				break
			}
			c.label = tok
		} else if c.op == nil {
			c.op = tok
		} else {
			c.args = append(c.args, tok)
		}
	}
	return input, c
}

func nextToken(b []byte) ([]byte, *Token, error) {
	// skip leading spaces
	for len(b) != 0 && b[0] != '\n' && unicode.IsSpace(rune(b[0])) {
		b = b[1:]
	}
	if len(b) == 0 {
		return b, &Token{Kind: EOF}, nil
	}

	var ch byte
	var text []byte

	ch, b = b[0], b[1:]
	switch ch {
	case '[': // bracket left
		if len(b) == 0 {
			return b, &Token{Kind: BRACKETL}, nil
		}
		for !(len(b) == 0 || b[0] == ']') {
			if unicode.IsSpace(rune(b[0])) {
				return b, &Token{
					Kind:  UNKNOWN,
					Text:  string(text),
					Error: fmt.Errorf("malformed label"),
				}, fmt.Errorf("malformed label")
			}
			text, b = append(text, b[0]), b[1:]
		}
		if len(b) == 0 || b[0] != ']' {
			return b, &Token{
				Kind:  UNKNOWN,
				Text:  string(text),
				Error: fmt.Errorf("unterminated label"),
			}, fmt.Errorf("unterminated label")
		}
		return b[1:], &Token{Kind: LABEL, Label: string(text)}, nil
	case ']': // bracket right
		return b, &Token{Kind: BRACKETR}, nil
	case ':': // colon
		return b, &Token{Kind: COLON}, nil
	case ',': // comma
		return b, &Token{Kind: COMMA}, nil
	case '<': // comp left
		return b, &Token{Kind: COMPL}, nil
	case '>': // comp right
		return b, &Token{Kind: COMPR}, nil
	case '-': // dash or start of a number
		if len(b) != 0 && unicode.IsDigit(rune(b[0])) {
			break // let it fall through to the number check
		}
		return b, &Token{Kind: DASH}, nil
	case '$': // dollar
		return b, &Token{Kind: DOLLAR}, nil
	case '=': // equals
		return b, &Token{Kind: EQUALS}, nil
	case '(': // paren left
		return b, &Token{Kind: PARENL}, nil
	case ')': // paren right
		return b, &Token{Kind: PARENR}, nil
	case '+': // plus or start of a number
		if len(b) != 0 && unicode.IsDigit(rune(b[0])) {
			break // let it fall through to the number check
		}
		return b, &Token{Kind: PLUS}, nil
	case '/': // slash
		return b, &Token{Kind: SLASH}, nil
	case '*': // star
		return b, &Token{Kind: STAR}, nil
	case '\n': // end of line
		return b, &Token{Kind: EOL}, nil
	case '\'': // quoted text
		for !(len(b) == 0 || b[0] == '\n' || b[0] == '\'') {
			text, b = append(text, b[0]), b[1:]
		}
		if len(b) == 0 || b[0] != '\'' {
			return b, &Token{
				Kind:  UNKNOWN,
				Text:  string(text),
				Error: fmt.Errorf("unterminated qtext"),
			}, fmt.Errorf("unterminated qtext")
		}
		return b[1:], &Token{Kind: QTEXT, QuotedText: string(text)}, nil
	}

	// everything else must be text or an integer
	text = append(text, ch)
	for !(len(b) == 0 || b[0] == ',' || unicode.IsSpace(rune(b[0]))) {
		text, b = append(text, b[0]), b[1:]
	}

	// it's an integer only if the stdlib says it is an integer
	if i, err := strconv.Atoi(string(text)); err == nil {
		return b, &Token{Kind: INTEGER, Integer: i}, nil
	}

	// otherwise it is just text
	return b, &Token{Kind: TEXT, Text: string(text)}, nil
}
