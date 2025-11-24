package vdf

import (
	"bufio"
	"io"
)

type Lexer struct {
	r *bufio.Reader

	ignoreWhitespace bool
}

func NewLexer(r io.Reader, ignoreWhitespace bool) *Lexer {
	return &Lexer{
		r:                bufio.NewReader(r),
		ignoreWhitespace: ignoreWhitespace,
	}
}

func (l *Lexer) read() (rune, error) {
	ch, _, err := l.r.ReadRune()
	if err != nil {
		return 0, err
	}
	return ch, nil
}

func (l *Lexer) unread() error {
	return l.r.UnreadRune()
}

func (l *Lexer) peek() (rune, error) {
	ch, err := l.read()
	if err != nil {
		return 0, err
	}

	if err := l.unread(); err != nil {
		return 0, err
	}

	return ch, nil
}

func (l *Lexer) skipWhitespace() error {
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

func (l *Lexer) Next() (*Token, error) {
	if l.ignoreWhitespace {
		if err := l.skipWhitespace(); err != nil {
			return nil, err
		}
	}

	ch, err := l.read()
	if err == io.EOF {
		return NewToken(ch), err
	}
	if err != nil {
		return nil, err
	}

	return NewToken(ch), nil
}

func (l *Lexer) Peek() (token *Token, err error) {
	defer func() {
		unreadErr := l.unread()
		if unreadErr != nil {

			// Check if l.Next() already returned an error
			if err == nil {
				err = unreadErr
			}
		}
	}()

	token, err = l.Next()
	return
}
