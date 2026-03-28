package vdf

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type lexer struct {
	input              []byte
	pos                int
	lineStarts         []int
	useEscapeSequences bool
	peekedToken        *Token
}

func newLexer(data []byte, useEscapeSequences bool) *lexer {
	return &lexer{
		input:              data,
		pos:                0,
		lineStarts:         []int{0},
		useEscapeSequences: useEscapeSequences,
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
	newPos := l.pos - size
	if newPos < 0 {
		return fmt.Errorf("invalid size: %d", size)
	}

	// Remove line starts that are after the new position
	if len(l.lineStarts) > 1 {
		for newPos < l.lineStarts[len(l.lineStarts)-1] {
			l.lineStarts = l.lineStarts[:len(l.lineStarts)-1]
		}
	}

	l.pos = newPos
	return nil
}

func calcLineAndColumn(input []byte, pos int, lineStarts []int) (line int, col int) {
	low := 0
	high := len(lineStarts)

	for low < high {
		mid := low + (high-low)/2
		if lineStarts[mid] <= pos {
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

	runeCount := utf8.RuneCount(input[lineStarts[lineIdx]:pos])
	col = runeCount + 1
	return line, col
}

func (l *lexer) skipWhitespace() error {
	l.peekedToken = nil
	for {
		startPos := l.pos

		for {
			r, size, err := l.read()
			if err != nil {
				return err
			}
			if !isWhitespace(r) {
				if err := l.unread(size); err != nil {
					return err
				}
				break
			}
		}

		if err := l.skipComments(); err != nil {
			return err
		}

		if startPos == l.pos {
			break
		}
	}
	return nil
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

		// Consume comment until newline is encountered
		for {
			r, _, err := l.read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			if r == '\n' {
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
				line, col := calcLineAndColumn(l.input, l.pos, l.lineStarts)
				return newEOFToken(line, col), nil
			}
			return nil, err
		}

		// Position didn't change, done skipping
		if startPos == l.pos {
			break
		}
	}

	line, col := calcLineAndColumn(l.input, l.pos, l.lineStarts)

	r, _, err := l.read()
	if err == io.EOF {
		return newEOFToken(line, col), nil
	}
	if err != nil {
		return nil, err
	}

	if r == '"' {
		value, err := l.readString()
		if err != nil {
			return nil, err
		}
		return newStringToken(value, line, col), nil
	}

	return newToken(r, line, col), nil
}

func (l *lexer) readString() (string, error) {
	var sb strings.Builder
	for {
		startPos := l.pos

		r, _, err := l.read()
		if err != nil {
			if err == io.EOF {
				line, col := calcLineAndColumn(l.input, startPos, l.lineStarts)
				return "", &SyntaxError{
					Line:    line,
					Column:  col,
					Message: "unterminated string literal",
				}
			}
			return "", err
		}

		if r == '\\' {
			if !l.useEscapeSequences {
				line, col := calcLineAndColumn(l.input, startPos, l.lineStarts)
				return "", &SyntaxError{
					Line:    line,
					Column:  col,
					Message: "escape sequence not allowed",
				}
			}

			next, _, err := l.read()
			if err != nil {
				if err == io.EOF {
					line, col := calcLineAndColumn(l.input, startPos, l.lineStarts)
					return "", &SyntaxError{
						Line:    line,
						Column:  col,
						Message: "unterminated string literal",
					}
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
				line, col := calcLineAndColumn(l.input, startPos, l.lineStarts)
				return "", &SyntaxError{
					Line:    line,
					Column:  col,
					Message: fmt.Sprintf("invalid escape sequence: \\%c", next),
				}
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
