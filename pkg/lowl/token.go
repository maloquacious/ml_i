package lowl

import (
	"fmt"
)

type TokenKind int

const (
	UNKNOWN TokenKind = iota
	BRACKETL
	BRACKETR
	COLON
	COMMA
	COMPL
	COMPR
	DASH
	DOLLAR
	EOF
	EOL
	EQUALS
	INTEGER
	LABEL
	PARENL
	PARENR
	PLUS
	QTEXT
	SLASH
	STAR
	TEXT
)

type Token struct {
	Kind       TokenKind
	Integer    int
	Label      string
	QuotedText string
	Text       string
	Error      error
}

// String implements the Stringer interface.
func (t Token) String() string {
	switch t.Kind {
	case BRACKETL:
		return "["
	case BRACKETR:
		return "]"
	case COLON:
		return ";"
	case COMMA:
		return ","
	case COMPL:
		return "<"
	case COMPR:
		return ">"
	case DASH:
		return "-"
	case DOLLAR:
		return "$"
	case EOF:
		return "*EOF*"
	case EOL:
		return "*EOL*"
	case INTEGER:
		return fmt.Sprintf("%d", t.Integer)
	case LABEL:
		return fmt.Sprintf("[%s]", t.Label)
	case PARENL:
		return "("
	case PARENR:
		return ")"
	case PLUS:
		return "+"
	case QTEXT:
		return fmt.Sprintf("'%s'", t.QuotedText)
	case SLASH:
		return "/"
	case STAR:
		return "*"
	case TEXT:
		return t.Text
	case UNKNOWN:
		return fmt.Sprintf("unknown(%q)", t.Text)
	}
	panic(fmt.Sprintf("assert(t.Kind != %d)", t.Kind))
}

// String implements the Stringer interface.
func (t TokenKind) String() string {
	switch t {
	case BRACKETL:
		return "BRACKETL"
	case BRACKETR:
		return "BRACKETR"
	case COLON:
		return "COLON"
	case COMMA:
		return "COMMA"
	case COMPL:
		return "COMPL"
	case COMPR:
		return "COMPR"
	case DASH:
		return "DASH"
	case DOLLAR:
		return "DOLLAR"
	case EOF:
		return "EOF"
	case EOL:
		return "EOL"
	case EQUALS:
		return "EQUALS"
	case INTEGER:
		return "INTEGER"
	case LABEL:
		return "LABEL"
	case PARENL:
		return "PARENL"
	case PARENR:
		return "PARENR"
	case PLUS:
		return "PLUS"
	case QTEXT:
		return "QTEXT"
	case SLASH:
		return "SLASH"
	case STAR:
		return "STAR"
	case TEXT:
		return "TEXT"
	case UNKNOWN:
		return "UNKNOWN"
	}
	panic(fmt.Sprintf("assert(TokenKind != %d)", t))
}
