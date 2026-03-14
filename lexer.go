package vdf

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type lexer struct {
	input               []byte
	pos                 int
	lineStarts          []int
	usesEscapeSequences bool
	peekedToken         *Token
}

func newLexer(data []byte, usesEscapeSequences bool) *lexer {
	return &lexer{
		input:               data,
		pos:                 0,
		lineStarts:          []int{0},
		usesEscapeSequences: usesEscapeSequences,
	}
}

func (l *lexer) read() (r rune, size int, err error) {
	if l.pos >= len(l.input) {
		return 0, 0, io.EOF
	}
	current := l.input[l.pos:]
	r, size = utf8.DecodeRune(current)

	l.pos += size

	if r == '\n' {
		l.lineStarts = append(l.lineStarts, l.pos)
	}
	return r, size, nil
}

func (l *lexer) unread(size int) error {
	if size < 0 || size > l.pos {
		return fmt.Errorf("invalid size: %d", size)
	}

	if len(l.lineStarts) > 1 && l.pos == l.lineStarts[len(l.lineStarts)-1] {
		l.lineStarts = l.lineStarts[:len(l.lineStarts)-1]
	}

	l.pos -= size
	return nil
}

func (l *lexer) calcLineAndColumn() (line int, col int) {
	low := 0
	high := len(l.lineStarts)

	for low < high {
		mid := low + (high-low)/2
		if l.lineStarts[mid] <= l.pos {
			low = mid + 1
		} else {
			high = mid
		}
	}

	lineIdx := low - 1
	if lineIdx < 0 {
		lineIdx = 0
	}

	line = lineIdx + 1
	col = l.pos - l.lineStarts[lineIdx] + 1
	return line, col
}

func (l *lexer) skipWhitespace() error {
	l.peekedToken = nil
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
	l.peekedToken = nil
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
			if err := l.unread(size + nextSz); err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}

		if next != '/' && next != '*' {
			if err := l.unread(size + nextSz); err != nil {
				return err
			}
			return nil
		}

		// Line comments and block comments end with a newline
		for {
			r, _, err := l.read()
			if err == io.EOF {
				// Comment ended at EOF
				return nil
			}
			if err != nil {
				return err
			}

			if r == '\n' {
				// Newline found, end of current comment
				break
			}
		}
	}
}

func (l *lexer) peek() (*Token, error) {
	if l.peekedToken != nil {
		return l.peekedToken, nil
	}

	token, err := l.next()
	if err != nil {
		return nil, err
	}

	l.peekedToken = token
	return token, nil
}

func (l *lexer) next() (*Token, error) {
	if l.peekedToken != nil {
		token := l.peekedToken
		l.peekedToken = nil
		return token, nil
	}

	for {
		startPos := l.pos

		if err := l.skipComments(); err != nil {
			if err == io.EOF {
				line, col := l.calcLineAndColumn()
				return NewEOFToken(line, col), nil
			}
			return nil, err
		}

		// Position didn't change, done skipping
		if startPos == l.pos {
			break
		}
	}

	line, col := l.calcLineAndColumn()

	r, _, err := l.read()
	if err == io.EOF {
		return NewEOFToken(line, col), nil
	}
	if err != nil {
		return nil, err
	}

	if r == '"' {
		value, err := l.readString()
		if err != nil {
			return nil, &SyntaxError{
				Line:    line,
				Column:  col,
				Message: err.Error(),
			}
		}
		return NewStringToken(value, line, col), nil
	}

	return NewToken(r, line, col), nil
}

func (l *lexer) readString() (string, error) {
	var sb strings.Builder
	for {
		r, _, err := l.read()
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("unterminated string literal")
			}
			return "", err
		}

		if r == '\\' {
			if !l.usesEscapeSequences {
				return "", fmt.Errorf("escape sequence not allowed")
			}

			// Handle escape sequences
			next, _, err := l.read()
			if err != nil {
				if err == io.EOF {
					return "", fmt.Errorf("unterminated string literal")
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
			case 'r':
				sb.WriteRune('\r')
			default:
				return "", fmt.Errorf("invalid escape sequence: %c", next)
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
