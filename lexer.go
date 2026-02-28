package vdf

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type lexer struct {
	input []byte

	pos        int
	lineStarts []int

	ignoreWhitespace bool
}

func newLexer(data []byte, ignoreWhitespace bool) *lexer {
	return &lexer{
		input:            data,
		pos:              0,
		lineStarts:       make([]int, 0),
		ignoreWhitespace: ignoreWhitespace,
	}
}

func (l *lexer) read() (r rune, size int, err error) {
	if l.pos >= len(l.input) {
		return 0, 0, io.EOF
	}
	current := l.input[l.pos:]
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

func (l *lexer) peek() (rune, error) {
	if l.pos >= len(l.input) {
		return 0, io.EOF
	}
	r, _ := utf8.DecodeRune(l.input[l.pos:])
	return r, nil
}

func (l *lexer) peekN(n int) (rune, error) {
	checkBounds := func(index int) error {
		if index >= len(l.input) {
			return io.EOF
		}
		return nil
	}
	if err := checkBounds(l.pos); err != nil {
		return 0, err
	}

	var r rune
	pos := l.pos
	for i := 0; i < n; i++ {
		if err := checkBounds(pos); err != nil {
			return 0, err
		}

		var size int
		r, size = utf8.DecodeRune(l.input[pos:])
		pos += size
	}
	return r, nil
}

func (l *lexer) next() (*Token, error) {
	for {
		startPos := l.pos

		if l.ignoreWhitespace {
			if err := l.skipWhitespace(); err != nil {
				if err == io.EOF {
					line, col := l.calcLineAndColumn()
					return NewEOFToken(line, col), nil
				}
				return nil, err
			}
		}

		// Always skip comments
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
			return nil, fmt.Errorf("line %d:%d: %w", line, col, err)
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
