package vdf

type TokenType uint8

const (
	ILLEGAL TokenType = iota

	// EOF marks the End-of-File
	EOF

	// IDENTIFIER can be a key or value
	IDENTIFIER

	// STRING is a quoted string
	STRING

	// WHITESPACE can be space, return, newline, or tabulator
	WHITESPACE

	// LBRACE is the literal opening curly brace `{`
	// Represents the start of VDF scope
	LBRACE

	// RBRACE is the literal closing curly brace `}`
	// Represents the end of VDF scope
	RBRACE

	// DOUBLEQUOTE is the literal `"`
	// Often represents the start or end of a key or value
	DOUBLEQUOTE

	// ESCAPE is the literal backslash `\`
	// Represents an escape character
	ESCAPE
)

func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENTIFIER:
		return "IDENTIFIER"
	case STRING:
		return "STRING"
	case WHITESPACE:
		return "WHITESPACE"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case DOUBLEQUOTE:
		return "DOUBLEQUOTE"
	case ESCAPE:
		return "ESCAPE"
	default:
		return "UNKNOWN"
	}
}

type Token struct {
	Type   TokenType
	Lexeme string
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\r' || ch == '\n' || ch == '\t'
}

func isLetter(ch rune) bool {
	return ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isControlCharacter(ch rune) bool {
	return ch == '{' || ch == '}' || ch == '"'
}

func isIdentifier(ch rune) bool {
	return !isControlCharacter(ch)
}

func NewToken(ch rune) *Token {
	token := &Token{
		Type:   ILLEGAL,
		Lexeme: string(ch),
	}

	if isWhitespace(ch) {
		token.Type = WHITESPACE
	} else if ch == '\\' {
		token.Type = ESCAPE
	} else if ch == '{' {
		token.Type = LBRACE
	} else if ch == '}' {
		token.Type = RBRACE
	} else if ch == '"' {
		token.Type = DOUBLEQUOTE
	} else if isIdentifier(ch) {
		token.Type = IDENTIFIER
	}
	return token
}

func NewStringToken(value string) *Token {
	return &Token{
		Type:   STRING,
		Lexeme: value,
	}
}

func NewEOFToken() *Token {
	return &Token{
		Type:   EOF,
		Lexeme: "",
	}
}
