package vdf

import (
	"fmt"
	"io"
)

type parser struct {
	lexer *lexer
}

func (p *parser) parse() (*Document, error) {
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

	var rootKey string
	switch token.Type {
	case EOF:
		return &Document{}, nil
	case STRING:
		rootKey = token.Lexeme
		p.lexer.next()
	case IDENTIFIER:
		key, err := p.parseUnquotedIdentifier()
		if err != nil {
			return nil, err
		}
		rootKey = key
	default:
		return nil, &SyntaxError{
			Line:    token.Line,
			Column:  token.Column,
			Message: fmt.Sprintf("invalid token type %s for root key", token.Type.String()),
		}
	}

	// Parse the root object
	subSlice, subMap, err := p.parseObject()
	if err != nil {
		return nil, err
	}

	doc := &Document{
		Root: &KeyValue{
			Key:   rootKey,
			Value: subSlice,
		},
		Map: map[string]any{
			rootKey: subMap,
		},
	}

	return doc, nil
}

func (p *parser) parseObject() ([]*KeyValue, map[string]any, error) {
	token, err := p.lexer.next()
	if err != nil {
		return nil, nil, err
	}

	if token.Type == WHITESPACE {
		if err := p.lexer.skipWhitespace(); err != nil {
			return nil, nil, err
		}
		token, err = p.lexer.next()
		if err != nil {
			return nil, nil, err
		}
	}

	if token.Type != LBRACE {
		return nil, nil, &SyntaxError{
			Line:    token.Line,
			Column:  token.Column,
			Message: fmt.Sprintf("invalid token %s, expected LBRACE", token.Type.String()),
		}
	}

	subKeyValues := make([]*KeyValue, 0)
	objMap := make(map[string]any)

	for {
		token, err = p.lexer.peek()
		if err != nil {
			return nil, nil, err
		}

		if token.Type == WHITESPACE {
			if err := p.lexer.skipWhitespace(); err != nil {
				if err == io.EOF {
					line, col := calcLineAndColumn(p.lexer.input, p.lexer.pos, p.lexer.lineStarts)
					return nil, nil, &SyntaxError{
						Line:    line,
						Column:  col,
						Message: "unexpected EOF",
					}
				}
				return nil, nil, err
			}
			token, err = p.lexer.peek()
			if err != nil {
				return nil, nil, err
			}
		}

		if token.Type == EOF {
			return nil, nil, &SyntaxError{
				Line:    token.Line,
				Column:  token.Column,
				Message: "unexpected EOF",
			}
		} else if token.Type == RBRACE {
			p.lexer.next()
			break
		}

		var parsedKey string
		switch token.Type {
		case STRING:
			parsedKey = token.Lexeme
			p.lexer.next()
		case IDENTIFIER:
			key, err := p.parseUnquotedIdentifier()
			if err != nil {
				return nil, nil, err
			}
			parsedKey = key
		default:
			return nil, nil, &SyntaxError{
				Line:    token.Line,
				Column:  token.Column,
				Message: fmt.Sprintf("invalid token %s, expected STRING or IDENTIFIER", token.Type.String()),
			}
		}

		kv := &KeyValue{Key: parsedKey}

		token, err = p.lexer.peek()
		if err != nil {
			return nil, nil, err
		}

		if token.Type == WHITESPACE {
			if err := p.lexer.skipWhitespace(); err != nil {
				if err == io.EOF {
					line, col := calcLineAndColumn(p.lexer.input, p.lexer.pos, p.lexer.lineStarts)
					return nil, nil, &SyntaxError{
						Line:    line,
						Column:  col,
						Message: "unexpected EOF",
					}
				}
				return nil, nil, err
			}
			token, err = p.lexer.peek()
			if err != nil {
				return nil, nil, err
			}
		}

		switch token.Type {
		case STRING:
			kv.Value = token.Lexeme
			objMap[kv.Key] = token.Lexeme
			p.lexer.next()
		case IDENTIFIER:
			value, err := p.parseUnquotedIdentifier()
			if err != nil {
				return nil, nil, err
			}
			kv.Value = value
			objMap[kv.Key] = value
		case LBRACE:
			obj, nestedMap, err := p.parseObject()
			if err != nil {
				return nil, nil, err
			}
			kv.Value = obj

			if existing, exists := objMap[kv.Key]; exists {
				if existingMap, ok := existing.(map[string]any); ok {
					mergeMaps(existingMap, nestedMap)
				} else {
					objMap[kv.Key] = nestedMap
				}
			} else {
				objMap[kv.Key] = nestedMap
			}
		default:
			return nil, nil, &SyntaxError{
				Line:    token.Line,
				Column:  token.Column,
				Message: fmt.Sprintf("invalid token %s, expected STRING, IDENTIFIER, or LBRACE", token.Type.String()),
			}
		}

		subKeyValues = append(subKeyValues, kv)
	}
	return subKeyValues, objMap, nil
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
