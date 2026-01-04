package vdf

import (
	"bufio"
	"io"
)

type lexer struct {
	r                *bufio.Reader
	ignoreWhitespace bool
}

func newLexer(r io.Reader, ignoreWhitespace bool) *lexer {
	return &lexer{
		r:                bufio.NewReader(r),
		ignoreWhitespace: ignoreWhitespace,
	}
}

func (l *lexer) read() (rune, error) {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return 0, err
	}
	return ch, nil
}

func (l *lexer) unread() error {
	return l.r.UnreadRune()
}

func (l *lexer) skipWhitespace() error {
	for {
		ch, err := l.read()
		if err != nil {
			return err
		}

		if isWhitespace(ch) {
			continue
		}

		if err := l.unread(); err != nil {
			return err
		}
		return nil
	}
}

func (l *lexer) next() (*Token, error) {
	if l.ignoreWhitespace {
		if err := l.skipWhitespace(); err != nil {
			if err == io.EOF {
				return NewEOFToken(), nil
			}
			return nil, err
		}
	}

	ch, err := l.read()
	if err == io.EOF {
		return NewEOFToken(), nil
	}
	if err != nil {
		return nil, err
	}

	return NewToken(ch), nil
}

func (l *lexer) peek() (*Token, error) {
	token, err := l.next()
	if err != nil {
		return nil, err
	}

	// Cannot unread EOF
	if token.Type == EOF {
		return token, nil
	}

	if err := l.unread(); err != nil {
		return nil, err
	}

	return token, nil
}
