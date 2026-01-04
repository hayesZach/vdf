package vdf

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestLexer_Read(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  []rune
	}{
		{
			name:  "readLoop",
			input: `"root"{"key""value"}`,

			want: []rune(`"root"{"key""value"}`),
		},
		{
			name: "readLoopWithWhitespace",
			input: `"root"
			{
				"key" "value"
			}`,
			want: []rune(`"root"
			{
				"key" "value"
			}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.input))
			lexer := newLexer(r, false /* ignoreWhitespace */)

			result := make([]rune, 0)
			for {
				ch, err := lexer.read()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("read(): %v", err)
				}

				result = append(result, ch)
			}

			if !reflect.DeepEqual(result, tc.want) {
				t.Fatalf("got runes %v, wanted %v", result, tc.want)
			}
		})
	}
}

func TestLexer_Unread(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  rune
	}{
		{
			name:  "unread",
			input: `test`,
			want:  't',
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.input))
			lexer := newLexer(r, false /* ignoreWhitespace */)

			_, err := lexer.read()
			if err != nil {
				t.Fatalf("read(): %v", err)
			}

			err = lexer.unread()
			if err != nil {
				t.Fatalf("unread(): %v", err)
			}

			ch, err := lexer.read()
			if err != nil {
				t.Fatalf("read(): %v", err)
			}

			if ch != tc.want {
				t.Fatalf("got rune %v, wanted %v", ch, tc.want)
			}
		})
	}
}

func TestLexer_Unread_EmptyString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "unreadEmptyString",
			input: ``,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.input))
			lexer := newLexer(r, false /* ignoreWhitespace */)

			if err := lexer.unread(); err == nil {
				t.Fatal("unread() succeeded, wanted error")
			}
		})
	}
}

func TestLexer_Next(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "simpleString",
			input: `"root"`,
			want: []TokenType{
				DOUBLEQUOTE,
				IDENTIFIER,
				IDENTIFIER,
				IDENTIFIER,
				IDENTIFIER,
				DOUBLEQUOTE,
				EOF,
			},
		},
		{
			name:  "withWhitespace",
			input: `" r o o t "`,
			want: []TokenType{
				DOUBLEQUOTE,
				WHITESPACE,
				IDENTIFIER,
				WHITESPACE,
				IDENTIFIER,
				WHITESPACE,
				IDENTIFIER,
				WHITESPACE,
				IDENTIFIER,
				WHITESPACE,
				DOUBLEQUOTE,
				EOF,
			},
		},
		{
			name:  "allTokenTypes",
			input: `{}"\ `,
			want: []TokenType{
				LBRACE,
				RBRACE,
				DOUBLEQUOTE,
				ESCAPE,
				WHITESPACE,
				EOF,
			},
		},
		{
			name:  "emptyInput",
			input: ``,
			want: []TokenType{
				EOF,
			},
		},
		{
			name:  "onlyWhitespace",
			input: " \t\n\r",
			want: []TokenType{
				WHITESPACE,
				WHITESPACE,
				WHITESPACE,
				WHITESPACE,
				EOF,
			},
		},
		{
			name:  "consecutiveSpecialChars",
			input: `{{}}""\\`,
			want: []TokenType{
				LBRACE,
				LBRACE,
				RBRACE,
				RBRACE,
				DOUBLEQUOTE,
				DOUBLEQUOTE,
				ESCAPE,
				ESCAPE,
				EOF,
			},
		},
		{
			name:  "realisticVDF",
			input: `"root"{"key""value"}`,
			want: []TokenType{
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // r o o t
				DOUBLEQUOTE,
				LBRACE,
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, // k e y
				DOUBLEQUOTE,
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // v a l u e
				DOUBLEQUOTE,
				RBRACE,
				EOF,
			},
		},
		{
			name:  "multilineVDF",
			input: "\"root\"\n{\n\t\"key\"\n}",
			want: []TokenType{
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // r o o t
				DOUBLEQUOTE,
				WHITESPACE, // \n
				LBRACE,
				WHITESPACE, // \n
				WHITESPACE, // \t
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, // k e y
				DOUBLEQUOTE,
				WHITESPACE, // \n
				RBRACE,
				EOF,
			},
		},
		{
			name:  "escapeSequences",
			input: `"text\"with\"escapes"`,
			want: []TokenType{
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // t e x t
				ESCAPE,
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // w i t h
				ESCAPE,
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // e s c a p e s
				DOUBLEQUOTE,
				EOF,
			},
		},
		{
			name:  "numbersAsIdentifiers",
			input: `"123""456"`,
			want: []TokenType{
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, // 1 2 3
				DOUBLEQUOTE,
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, // 4 5 6
				DOUBLEQUOTE,
				EOF,
			},
		},
		{
			name:  "specialCharsAsIdentifiers",
			input: `abc!@#$%^&*()`,
			want: []TokenType{
				IDENTIFIER, IDENTIFIER, IDENTIFIER, // a b c
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // ! @ # $ %
				IDENTIFIER, IDENTIFIER, IDENTIFIER, // ^ & *
				IDENTIFIER, IDENTIFIER, // ( )
				EOF,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.input))
			lexer := newLexer(r, false /* ignoreWhitespace */)

			result := make([]TokenType, 0)
			for {
				token, err := lexer.next()
				if err != nil {
					t.Fatalf("next(): %v", err)
				}

				result = append(result, token.Type)
				if token.Type == EOF {
					break
				}
			}

			if !reflect.DeepEqual(result, tc.want) {
				t.Fatalf("got runes %v, wanted %v", result, tc.want)
			}
		})
	}
}

func TestLexer_Next_IgnoreWhitespace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "whitespaceIgnored",
			input: `{ } " \ `,
			want: []TokenType{
				LBRACE,
				RBRACE,
				DOUBLEQUOTE,
				ESCAPE,
				EOF,
			},
		},
		{
			name:  "multilineWithWhitespaceIgnored",
			input: "\t\n  \"key\"\n\t  \"value\"  \n",
			want: []TokenType{
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, // k e y
				DOUBLEQUOTE,
				DOUBLEQUOTE,
				IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, IDENTIFIER, // v a l u e
				DOUBLEQUOTE,
				EOF,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.input))
			lexer := newLexer(r, true /* ignoreWhitespace */)

			result := make([]TokenType, 0)
			for {
				token, err := lexer.next()
				if err != nil {
					t.Fatalf("readNext(): %v", err)
				}

				result = append(result, token.Type)
				if token.Type == EOF {
					break
				}
			}

			if !reflect.DeepEqual(result, tc.want) {
				t.Fatalf("got %v, wanted %v", result, tc.want)
			}
		})
	}
}

func TestLexer_Peek(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  *Token
	}{
		{
			name:  "simpleString",
			input: `"root"`,
			want: &Token{
				Type:   DOUBLEQUOTE,
				Lexeme: "\"",
			},
		},
		{
			name:  "emptyString",
			input: "",
			want: &Token{
				Type:   EOF,
				Lexeme: "",
			},
		},
		{
			name:  "peekLBrace",
			input: `{`,
			want: &Token{
				Type:   LBRACE,
				Lexeme: "{",
			},
		},
		{
			name:  "peekRBrace",
			input: `}`,
			want: &Token{
				Type:   RBRACE,
				Lexeme: "}",
			},
		},
		{
			name:  "peekEscape",
			input: `\n`,
			want: &Token{
				Type:   ESCAPE,
				Lexeme: "\\",
			},
		},
		{
			name:  "peekWhitespace",
			input: " test",
			want: &Token{
				Type:   WHITESPACE,
				Lexeme: " ",
			},
		},
		{
			name:  "peekTab",
			input: "\tvalue",
			want: &Token{
				Type:   WHITESPACE,
				Lexeme: "\t",
			},
		},
		{
			name:  "peekNewline",
			input: "\nvalue",
			want: &Token{
				Type:   WHITESPACE,
				Lexeme: "\n",
			},
		},
		{
			name:  "peekIdentifier",
			input: "abc",
			want: &Token{
				Type:   IDENTIFIER,
				Lexeme: "a",
			},
		},
		{
			name:  "peekDigit",
			input: "123",
			want: &Token{
				Type:   IDENTIFIER,
				Lexeme: "1",
			},
		},
		{
			name:  "peekSpecialChar",
			input: "@test",
			want: &Token{
				Type:   IDENTIFIER,
				Lexeme: "@",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.input))
			lexer := newLexer(r, false /* ignoreWhitespace */)

			token, err := lexer.peek()
			if err != nil {
				t.Fatalf("peek(): %v", err)
			}

			if token.Type != tc.want.Type {
				t.Errorf("got token type %v, wanted %v", token.Type, tc.want.Type)
			}
			if token.Lexeme != tc.want.Lexeme {
				t.Errorf("got lexeme %s, wanted %s", token.Lexeme, tc.want.Lexeme)
			}
		})
	}
}

func TestLexer_Peek_IgnoreWhitespace(t *testing.T) {
	t.Parallel()

	input := "   \"key\""
	r := bytes.NewReader([]byte(input))
	lexer := newLexer(r, true /* ignoreWhitespace */)

	token, err := lexer.peek()
	if err != nil {
		t.Fatalf("peek(): %v", err)
	}

	// Should skip whitespace and peek at the double quote
	if token.Type != DOUBLEQUOTE {
		t.Errorf("got token type %v, wanted %v", token.Type, DOUBLEQUOTE)
	}
	if token.Lexeme != "\"" {
		t.Errorf("got lexeme %q, wanted %q", token.Lexeme, "\"")
	}
}

func TestLexer_Peek_DoesNotConsume(t *testing.T) {
	t.Parallel()

	input := `"root"`
	r := bytes.NewReader([]byte(input))
	lexer := newLexer(r, false /* ignoreWhitespace */)

	expectedToken := &Token{
		Type:   DOUBLEQUOTE,
		Lexeme: `"`,
	}

	token1, err := lexer.peek()
	if err != nil {
		t.Fatalf("peek(): %v", err)
	}

	if token1.Type != expectedToken.Type {
		t.Errorf("got token type %v, wanted %v", token1.Type, expectedToken.Type)
	}
	if token1.Lexeme != expectedToken.Lexeme {
		t.Errorf("got lexeme %s, wanted %s", token1.Lexeme, expectedToken.Lexeme)
	}

	token2, err := lexer.peek()
	if err != nil {
		t.Fatalf("peek(): %v", err)
	}

	if token2.Type != expectedToken.Type {
		t.Errorf("got token type %v, wanted %v", token2.Type, expectedToken.Type)
	}
	if token2.Lexeme != expectedToken.Lexeme {
		t.Errorf("got lexeme %s, wanted %s", token2.Lexeme, expectedToken.Lexeme)
	}

	token3, err := lexer.peek()
	if err != nil {
		t.Fatalf("peek(): %v", err)
	}

	if token3.Type != expectedToken.Type {
		t.Errorf("got token type %v, wanted %v", token3.Type, expectedToken.Type)
	}
	if token3.Lexeme != expectedToken.Lexeme {
		t.Errorf("got lexeme %s, wanted %s", token3.Lexeme, expectedToken.Lexeme)
	}

	// Verify that next() returns the same token as the calls to peek()
	nextToken, err := lexer.next()
	if err != nil {
		t.Fatalf("next(): %v", err)
	}

	if nextToken.Type != expectedToken.Type {
		t.Errorf("got token type %v, wanted %v", nextToken.Type, expectedToken.Type)
	}
	if nextToken.Lexeme != expectedToken.Lexeme {
		t.Errorf("got lexeme %s, wanted %s", nextToken.Lexeme, expectedToken.Lexeme)
	}
}
