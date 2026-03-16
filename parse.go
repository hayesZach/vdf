package vdf

import (
	"fmt"
	"io"
)

type parser struct {
	lexer *lexer
}

func (p *parser) parse() (*KeyValue, error) {
	root := &KeyValue{}

	token, err := p.lexer.peek()
	if err != nil {
		return nil, err
	}

	if token.Type == WHITESPACE {
		if err := p.lexer.skipWhitespace(); err != nil {
			return nil, err
		}
		token, err = p.lexer.peek()
		if err != nil {
			return nil, err
		}
	}

	switch token.Type {
	case EOF:
		return root, nil
	case STRING:
		root.Key = token.Lexeme
		p.lexer.next()
	case IDENTIFIER:
		key, err := p.parseUnquotedIdentifier()
		if err != nil {
			return nil, err
		}
		root.Key = key
	default:
		return nil, &SyntaxError{
			Line:    token.Line,
			Column:  token.Column,
			Message: fmt.Sprintf("invalid token type %s for root key", token.Type.String()),
		}
	}

	// Parse the value, should be an object for root
	val, err := p.parseObject()
	if err != nil {
		return nil, err
	}
	root.Value = val

	return root, nil
}

func (p *parser) parseObject() ([]*KeyValue, error) {
	token, err := p.lexer.next()
	if err != nil {
		return nil, err
	}

	if token.Type == WHITESPACE {
		if err := p.lexer.skipWhitespace(); err != nil {
			return nil, err
		}
		token, err = p.lexer.next()
		if err != nil {
			return nil, err
		}
	}

	if token.Type != LBRACE {
		return nil, &SyntaxError{
			Line:    token.Line,
			Column:  token.Column,
			Message: fmt.Sprintf("invalid token %s, expected LBRACE", token.Type.String()),
		}
	}

	subKeyValues := make([]*KeyValue, 0)
	for {
		token, err = p.lexer.peek()
		if err != nil {
			return nil, err
		}

		if token.Type == WHITESPACE {
			if err := p.lexer.skipWhitespace(); err != nil {
				if err == io.EOF {
					line, col := calcLineAndColumn(p.lexer.input, p.lexer.pos, p.lexer.lineStarts)
					return nil, &SyntaxError{
						Line:    line,
						Column:  col,
						Message: "unexpected EOF",
					}
				}
				return nil, err
			}
			token, err = p.lexer.peek()
			if err != nil {
				return nil, err
			}
		}

		if token.Type == EOF {
			return nil, &SyntaxError{
				Line:    token.Line,
				Column:  token.Column,
				Message: "unexpected EOF",
			}
		} else if token.Type == RBRACE {
			p.lexer.next()
			break
		}

		kv := &KeyValue{}
		switch token.Type {
		case STRING:
			kv.Key = token.Lexeme
			p.lexer.next()
		case IDENTIFIER:
			key, err := p.parseUnquotedIdentifier()
			if err != nil {
				return nil, err
			}
			kv.Key = key
		default:
			return nil, &SyntaxError{
				Line:    token.Line,
				Column:  token.Column,
				Message: fmt.Sprintf("invalid token %s, expected STRING or IDENTIFIER", token.Type.String()),
			}
		}

		token, err = p.lexer.peek()
		if err != nil {
			return nil, err
		}

		if token.Type == WHITESPACE {
			if err := p.lexer.skipWhitespace(); err != nil {
				if err == io.EOF {
					line, col := calcLineAndColumn(p.lexer.input, p.lexer.pos, p.lexer.lineStarts)
					return nil, &SyntaxError{
						Line:    line,
						Column:  col,
						Message: "unexpected EOF",
					}
				}
				return nil, err
			}
			token, err = p.lexer.peek()
			if err != nil {
				return nil, err
			}
		}

		switch token.Type {
		case STRING:
			kv.Value = token.Lexeme
			p.lexer.next()
		case IDENTIFIER:
			value, err := p.parseUnquotedIdentifier()
			if err != nil {
				return nil, err
			}
			kv.Value = value
		case LBRACE:
			obj, err := p.parseObject()
			if err != nil {
				return nil, err
			}
			kv.Value = obj
		default:
			return nil, &SyntaxError{
				Line:    token.Line,
				Column:  token.Column,
				Message: fmt.Sprintf("invalid token %s, expected STRING, IDENTIFIER, or LBRACE", token.Type.String()),
			}
		}
		subKeyValues = append(subKeyValues, kv)
	}
	return subKeyValues, nil
}

func (p *parser) parseUnquotedIdentifier() (string, error) {
	var value string
	for {
		token, err := p.lexer.peek()
		if err != nil {
			return "", err
		}

		if token.Type == WHITESPACE || token.Type == LBRACE || token.Type == RBRACE || token.Type == STRING {
			break
		} else if token.Type == IDENTIFIER {
			value += token.Lexeme
			p.lexer.next()
		} else {
			return "", &SyntaxError{
				Line:    token.Line,
				Column:  token.Column,
				Message: fmt.Sprintf("invalid token type %s for unquoted identifier", token.Type.String()),
			}
		}
	}
	return value, nil
}
