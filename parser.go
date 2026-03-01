package vdf

import (
	"fmt"
	"io"
)

type options struct {
	ignoreWhitespace    bool
	usesEscapeSequences bool
}

type Option func(*options)

func IgnoreWhitespace() Option {
	return func(o *options) {
		o.ignoreWhitespace = true
	}
}

func UseEscapeSequences() Option {
	return func(o *options) {
		o.usesEscapeSequences = true
	}
}

type Parser struct {
	lexer *lexer
}

func NewParser(r io.Reader, opts ...Option) (*Parser, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	return &Parser{
		lexer: newLexer(data, o.ignoreWhitespace, o.usesEscapeSequences),
	}, nil
}

func (p *Parser) Parse() (*KeyValue, error) {
	return p.parse()
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
		return nil, fmt.Errorf("invalid token type %q for root key", token.Type.String())
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
			return nil, fmt.Errorf("encountered unexpected EOF while parsing object")
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
			return nil, fmt.Errorf("invalid token %q, expected STRING or IDENTIFIER", token.Type.String())
		}

		if err := p.lexer.skipWhitespace(); err != nil {
			return nil, err
		}

		token, err = p.lexer.peek()
		if err != nil {
			return nil, err
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
			return nil, fmt.Errorf("invalid token %q, expected STRING, IDENTIFIER, or LBRACE", token.Type.String())
		}
		subKeyValues = append(subKeyValues, kv)
	}
	return subKeyValues, nil
}

func (p *Parser) parseUnquotedIdentifier() (string, error) {
	var value string
	for {
		token, err := p.lexer.next()
		if err != nil {
			return "", err
		}

		if token.Type == WHITESPACE || token.Type == LBRACE || token.Type == RBRACE {
			break
		} else if token.Type == IDENTIFIER {
			value += token.Lexeme
		} else {
			return "", fmt.Errorf("invalid token type %q for unquoted identifier", token.Type.String())
		}
	}
	return value, nil
}
