package vdf

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type lexer struct {
	buf []byte
	pos int

	ignoreWhitespace bool
}

func newLexer(data []byte, ignoreWhitespace bool) *lexer {
	return &lexer{
		buf:              data,
		pos:              0,
		ignoreWhitespace: ignoreWhitespace,
	}
}

func (l *lexer) read() (r rune, size int, err error) {
	if l.pos >= len(l.buf) {
		return 0, 0, io.EOF
	}
	current := l.buf[l.pos:]
	r, size = utf8.DecodeRune(current)

	l.pos += size
	return r, size, nil
}

func (l *lexer) unread(size int) error {
	if size < 0 || size > l.pos {
		return fmt.Errorf("invalid size: %d", size)
	}
	l.pos -= size
	return nil
}

func (l *lexer) skipWhitespace() error {
	for {
		r, size, err := l.read()
		if err != nil {
			return err
		}
		if isWhitespace(r) {
			continue
		}

		if err := l.unread(size); err != nil {
			return err
		}
		return nil
	}
}

func (l *lexer) skipComments() error {
	for {
		r, size, err := l.read()
		if err != nil {
			return err
		}

		if r != '/' {
			if err := l.unread(size); err != nil {
				return err
			}
			return nil
		}

		next, nextSz, err := l.read()
		if err == io.EOF {
			if err := l.unread(size); err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}

		switch next {
		case '/':
			// Line comment - consume until newline encountered
			for {
				r, _, err := l.read()
				if err == io.EOF {
					// Line comment ended at EOF
					return nil
				}
				if err != nil {
					return err
				}

				if r == '\n' {
					break
				}
			}
		case '*':
			// Block comment - consume until block comment ends
			var prev rune
			for {
				r, _, err := l.read()
				if err == io.EOF {
					return fmt.Errorf("block comment ends without being terminated")
				}
				if err != nil {
					return err
				}

				if prev == '*' && r == '/' {
					break
				}
				prev = r
			}
		default:
			if err := l.unread(size + nextSz); err != nil {
				return err
			}
			return nil
		}
	}
}

func (l *lexer) peek() (rune, error) {
	if l.pos >= len(l.buf) {
		return 0, io.EOF
	}
	r, _ := utf8.DecodeRune(l.buf[l.pos:])
	return r, nil
}

func (l *lexer) peekN(n int) (rune, error) {
	boundsCheck := func(index int) error {
		if index >= len(l.buf) {
			return io.EOF
		}
		return nil
	}
	if err := boundsCheck(l.pos); err != nil {
		return 0, err
	}

	pos := l.pos
	for i := 0; i < n-1; i++ {
		if err := boundsCheck(pos); err != nil {
			return 0, err
		}
		_, size := utf8.DecodeRune(l.buf[pos:])
		pos += size
	}

	if err := boundsCheck(pos); err != nil {
		return 0, err
	}

	r, _ := utf8.DecodeRune(l.buf[pos:])
	return r, nil
}

func (l *lexer) next() (*Token, error) {
	for {
		startPos := l.pos

		if l.ignoreWhitespace {
			if err := l.skipWhitespace(); err != nil {
				if err == io.EOF {
					return NewEOFToken(), nil
				}
				return nil, err
			}
		}

		// Always skip comments
		if err := l.skipComments(); err != nil {
			if err == io.EOF {
				return NewEOFToken(), nil
			}
			return nil, err
		}

		// Position didn't change, done skipping
		if startPos == l.pos {
			break
		}
	}

	r, _, err := l.read()
	if err != nil {
		if err == io.EOF {
			return NewEOFToken(), nil
		}
		return nil, err
	}

	if r == '"' {
		value, err := l.readString()
		if err != nil {
			return nil, err
		}
		return NewStringToken(value), nil
	}
	return NewToken(r), nil
}

func (l *lexer) readString() (string, error) {
	var sb strings.Builder
	for {
		r, _, err := l.read()
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("unterminated string")
			}
			return "", err
		}

		if r == '\\' {
			// Handle escape sequences
			next, _, err := l.read()
			if err != nil {
				if err == io.EOF {
					return "", fmt.Errorf("unterminated escape sequence")
				}
				return "", err
			}

			switch next {
			case '"':
				sb.WriteRune('"')
			case '\\':
				sb.WriteRune('\\')
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(next)
			}
			continue
		}

		if r == '"' {
			break
		}
		sb.WriteRune(r)
	}
	return sb.String(), nil
}
