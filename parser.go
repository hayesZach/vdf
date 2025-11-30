package vdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Parser struct {
	lexer               *Lexer
	usesEscapeSequences bool
}

func NewParser(r io.Reader) *Parser {
	return NewParserWithOptions(r, false /* ignoreWhitespace */, false /* usesEscapeSequences */)
}

func NewParserWithOptions(r io.Reader, ignoreWhitespace bool, usesEscapeSequences bool) *Parser {
	return &Parser{
		lexer:               NewLexer(r, ignoreWhitespace),
		usesEscapeSequences: usesEscapeSequences,
	}
}

func Parse(b []byte) (*KeyValue, error) {
	parser := NewParser(bytes.NewReader(b))
	return parser.parse()
}

func (p *Parser) parseStringValue(token *Token) (string, error) {
	isQuoted := token.Type == DOUBLEQUOTE
	return p.parseToken(isQuoted)
}

func (p *Parser) parse() (*KeyValue, error) {
	root := &KeyValue{}

	// Ignore whitespace initially
	if err := p.lexer.skipWhitespace(); err != nil {
		return nil, err
	}

	token, err := p.lexer.peek()
	if err != nil {
		return nil, err
	}

	if token.Type == EOF {
		return root, nil
	} else if token.Type == IDENTIFIER || token.Type == DOUBLEQUOTE {
		key, err := p.parseStringValue(token)
		if err != nil {
			return nil, err
		}
		root.Key = key
	}

	// Ignore whitespace after parsing key
	if err := p.lexer.skipWhitespace(); err != nil {
		return nil, err
	}

	// Parse the value (should be a sub-object for root)
	val, err := p.parseObject()
	if err != nil {
		return nil, err
	}
	root.Value = val

	return root, nil
}

func (p *Parser) parseObject() ([]*KeyValue, error) {
	if err := p.lexer.skipWhitespace(); err != nil {
		return nil, err
	}

	token, err := p.lexer.next()
	if err != nil {
		return nil, err
	}

	if token.Type != LBRACE {
		return nil, fmt.Errorf("invalid token %q, expected LBRACE", token.Type.String())
	}

	subKeyValues := make([]*KeyValue, 0)

	for {
		if err := p.lexer.skipWhitespace(); err != nil {
			return nil, err
		}

		token, err = p.lexer.peek()
		if err != nil {
			return nil, err
		}

		if token.Type == EOF {
			return nil, errors.New("unexpected EOF")
		}

		if token.Type == RBRACE {
			// Finished parsing object -- consume the closing brace
			if _, err := p.lexer.next(); err != nil {
				return nil, err
			}
			break
		}

		kv := &KeyValue{}

		// Handle quoted and unquoted keys and values
		if token.Type == IDENTIFIER || token.Type == DOUBLEQUOTE {
			key, err := p.parseStringValue(token)
			if err != nil {
				return nil, err
			}
			kv.Key = key
		}

		// Skip whitespace after key
		if err := p.lexer.skipWhitespace(); err != nil {
			return nil, err
		}

		// Handle parsing value
		token, err = p.lexer.peek()
		if err != nil {
			return nil, err
		}

		if token.Type == LBRACE {
			val, err := p.parseObject()
			if err != nil {
				return nil, err
			}
			kv.Value = val
		} else if token.Type == IDENTIFIER || token.Type == DOUBLEQUOTE {
			value, err := p.parseStringValue(token)
			if err != nil {
				return nil, err
			}
			kv.Value = value
		}
		subKeyValues = append(subKeyValues, kv)
	}
	return subKeyValues, nil
}

func (p *Parser) parseToken(isQuoted bool) (string, error) {
	if isQuoted {
		// Consume the opening double quote
		if _, err := p.lexer.next(); err != nil {
			return "", err
		}
	}

	var value string
	for {
		token, err := p.lexer.next()
		if err != nil {
			return "", err
		}

		if isQuoted {
			if token.Type == DOUBLEQUOTE {
				// Finished consuming the token
				break
			} else if token.Type == IDENTIFIER || token.Type == WHITESPACE || token.Type == LBRACE || token.Type == RBRACE {
				value += token.Lexeme
			} else if token.Type == ESCAPE {
				if !p.usesEscapeSequences {
					return "", errors.New("escape sequence not allowed")
				}

				token, err := p.lexer.peek()
				if err != nil {
					return "", err
				} else if token.Type != DOUBLEQUOTE {
					return "", errors.New("escape sequence must be followed by double quote")
				}

				token, err = p.lexer.next()
				if err != nil {
					return "", err
				}
				value += token.Lexeme
			}
		} else {
			if token.Type == WHITESPACE || token.Type == LBRACE || token.Type == RBRACE || token.Type == DOUBLEQUOTE {
				// Unquoted tokens end with whitespace or any control character
				break
			} else if token.Type != IDENTIFIER {
				return "", fmt.Errorf("invalid token type %q for unquoted identifier", token.Type.String())
			}
			value += token.Lexeme
		}
	}
	return value, nil
}
