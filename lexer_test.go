package vdf

import (
	"io"
	"reflect"
	"testing"
)

func TestLexer_read(t *testing.T) {
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
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */)

			result := make([]rune, 0)
			for {
				ch, _, err := lexer.read()
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

func TestLexer_unread(t *testing.T) {
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
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */)

			_, size, err := lexer.read()
			if err != nil {
				t.Fatalf("read(): %v", err)
			}

			err = lexer.unread(size)
			if err != nil {
				t.Fatalf("unread(): %v", err)
			}

			ch, _, err := lexer.read()
			if err != nil {
				t.Fatalf("read(): %v", err)
			}

			if ch != tc.want {
				t.Fatalf("got rune %v, wanted %v", ch, tc.want)
			}
		})
	}
}

func TestLexer_unread_EmptyString(t *testing.T) {
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
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */)

			if err := lexer.unread(1); err == nil {
				t.Fatal("unread() succeeded, wanted error")
			}
		})
	}
}

func TestLexer_skipComments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantPos int  // expected position after skipComments
		wantErr bool // only true for EOF on empty input
	}{
		{
			name:    "lineComment",
			input:   "// this is a comment\nrest",
			wantPos: 21, // position after newline
		},
		{
			name:    "lineCommentAtEOF",
			input:   "// comment without newline",
			wantPos: 26, // consumes entire input
		},
		{
			name:    "blockCommentEndsAtNewline",
			input:   "/* block comment\nrest",
			wantPos: 17, // position after newline (block comment ends at \n)
		},
		{
			name:    "blockCommentAtEOF",
			input:   "/* comment without newline",
			wantPos: 26, // consumes entire input
		},
		{
			name:    "blockCommentTraditionalSyntax",
			input:   "/* comment */\nrest",
			wantPos: 14, // ends at newline, not at */
		},
		{
			name:    "multipleLineComments",
			input:   "// first\n// second\nrest",
			wantPos: 19, // skips both comments
		},
		{
			name:    "multipleBlockComments",
			input:   "/* first\n/* second\nrest",
			wantPos: 19, // skips both comments (each ends at newline)
		},
		{
			name:    "mixedComments",
			input:   "// line\n/* block\nrest",
			wantPos: 17,
		},
		{
			name:    "notAComment",
			input:   "notacomment",
			wantPos: 0, // unreads the 'n', stays at start
		},
		{
			name:    "slashFollowedByOther",
			input:   "/notacomment",
			wantPos: 0, // unreads both '/' and 'n'
		},
		{
			name:    "slashAtEOF",
			input:   "/",
			wantPos: 0, // unreads the '/'
		},
		{
			name:    "emptyInput",
			input:   "",
			wantPos: 0,
			wantErr: true, // EOF error
		},
		{
			name:    "commentThenQuotedString",
			input:   "// comment\n\"key\"",
			wantPos: 11, // stops before "key"
		},
		{
			name:    "onlyNewline",
			input:   "\nrest",
			wantPos: 0, // not a comment, unreads
		},
		{
			name:    "slashSlashNoContent",
			input:   "//\nrest",
			wantPos: 3, // skips // and newline
		},
		{
			name:    "slashStarNoContent",
			input:   "/*\nrest",
			wantPos: 3, // skips /* and newline
		},
		{
			name:    "commentsOnSameLine",
			input:   "// first // second\nrest",
			wantPos: 19, // entire line is one comment
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */)

			err := lexer.skipComments()
			if tc.wantErr {
				if err == nil {
					t.Fatal("skipComments() succeeded, wanted error")
				}
				return
			}

			if err != nil {
				t.Fatalf("skipComments(): %v", err)
			}

			if lexer.pos != tc.wantPos {
				t.Errorf("wanted pos %d, got %d", tc.wantPos, lexer.pos)
			}
		})
	}
}

func TestLexer_peekN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		n       int
		want    rune
		wantErr bool
	}{
		{
			name:  "peekOneRune",
			input: "test",
			n:     1,
			want:  't',
		},
		{
			name:  "peekMultipleRunes",
			input: "test",
			n:     3,
			want:  's',
		},
		{
			name:  "simpleString",
			input: `"root"`,
			n:     2,
			want:  'r',
		},
		{
			name:  "peekLBrace",
			input: `{`,
			n:     1,
			want:  '{',
		},
		{
			name:  "peekRBrace",
			input: `}`,
			n:     1,
			want:  '}',
		},
		{
			name:  "peekEscape",
			input: `\n`,
			n:     1,
			want:  '\\',
		},
		{
			name:  "peekWhitespace",
			input: " test",
			n:     1,
			want:  ' ',
		},
		{
			name:  "peekTab",
			input: "\tvalue",
			n:     1,
			want:  '\t',
		},
		{
			name:  "peekNewline",
			input: "\nvalue",
			n:     1,
			want:  '\n',
		},
		{
			name:  "peekIdentifier",
			input: "abc",
			n:     1,
			want:  'a',
		},
		{
			name:  "peekDigit",
			input: "123",
			n:     1,
			want:  '1',
		},
		{
			name:  "peekSpecialChar",
			input: "@test",
			n:     1,
			want:  '@',
		},
		{
			name:    "peekBeyondInput",
			input:   "test",
			n:       5,
			want:    0,
			wantErr: true,
		},
		{
			name:    "peekFurther",
			input:   "test",
			n:       7,
			want:    0,
			wantErr: true,
		},
		{
			name:    "emptyInput",
			input:   "",
			n:       1,
			want:    0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false)

			r, err := lexer.peekN(tc.n)
			if tc.wantErr {
				if err == nil {
					t.Fatal("peekN() succeeded, wanted error")
				}
				if r != tc.want {
					t.Errorf("wanted rune %v, got %v", tc.want, r)
				}
				return
			}
			if err != nil {
				t.Fatalf("peekN(): %v", err)
			}

			if r != tc.want {
				t.Fatalf("wanted rune %v, got %v", tc.want, r)
			}
		})
	}
}

func TestLexer_peekN_DoesNotConsume(t *testing.T) {
	t.Parallel()

	input := `"root"`
	lexer := newLexer([]byte(input), false /* ignoreWhitespace */)

	expectedRune := '"'

	for i := 0; i < 3; i++ {
		r, err := lexer.peekN(1)
		if err != nil {
			t.Fatalf("peekN(): %v", err)
		}
		if r != expectedRune {
			t.Errorf("got rune %v, wanted %v", r, expectedRune)
		}
	}

	r, _, err := lexer.read()
	if err != nil {
		t.Fatalf("read(): %v", err)
	}
	if r != expectedRune {
		t.Errorf("got rune %v, wanted %v", r, expectedRune)
	}
}

func TestLexer_peek(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    rune
		wantErr bool
	}{
		{
			name:  "simpleString",
			input: `"root"`,
			want:  '"',
		},
		{
			name:  "peekLBrace",
			input: `{`,
			want:  '{',
		},
		{
			name:  "peekRBrace",
			input: `}`,
			want:  '}',
		},
		{
			name:  "peekEscape",
			input: `\n`,
			want:  '\\',
		},
		{
			name:  "peekWhitespace",
			input: " test",
			want:  ' ',
		},
		{
			name:  "peekTab",
			input: "\tvalue",
			want:  '\t',
		},
		{
			name:  "peekNewline",
			input: "\nvalue",
			want:  '\n',
		},
		{
			name:  "peekIdentifier",
			input: "abc",
			want:  'a',
		},
		{
			name:  "peekDigit",
			input: "123",
			want:  '1',
		},
		{
			name:  "peekSpecialChar",
			input: "@test",
			want:  '@',
		},
		{
			name:    "emptyInput",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */)

			r, err := lexer.peek()
			if tc.wantErr {
				if err == nil {
					t.Fatal("peek() succeeded, wanted error")
				}
				if r != tc.want {
					t.Errorf("got rune %v, wanted %v", r, tc.want)
				}
				return
			}
			if err != nil {
				t.Fatalf("peek(): %v", err)
			}

			if r != tc.want {
				t.Errorf("got rune %v, wanted %v", r, tc.want)
			}
		})
	}
}

func TestLexer_peek_DoesNotConsume(t *testing.T) {
	t.Parallel()

	input := `"root"`
	lexer := newLexer([]byte(input), false /* ignoreWhitespace */)

	expectedRune := '"'

	// Peek multiple times to ensure that the rune is not consumed
	for i := 0; i < 3; i++ {
		r, err := lexer.peek()
		if err != nil {
			t.Fatalf("peek(): %v", err)
		}
		if r != expectedRune {
			t.Errorf("got rune %v, wanted %v", r, expectedRune)
		}
	}

	firstRune, _, err := lexer.read()
	if err != nil {
		t.Fatalf("read(): %v", err)
	}
	if firstRune != expectedRune {
		t.Errorf("got rune %v, wanted %v", firstRune, expectedRune)
	}
}

func TestLexer_next(t *testing.T) {
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
				STRING,
				EOF,
			},
		},
		{
			name:  "withWhitespace",
			input: `" r o o t "`,
			want: []TokenType{
				STRING,
				EOF,
			},
		},
		{
			name:  "allTokenTypesExceptDoubleQuotes",
			input: `{}\ `,
			want: []TokenType{
				LBRACE,
				RBRACE,
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
				STRING,
				ESCAPE,
				ESCAPE,
				EOF,
			},
		},
		{
			name:  "realisticVDF",
			input: `"root"{"key""value"}`,
			want: []TokenType{
				STRING, // root
				LBRACE,
				STRING, // key
				STRING, // value
				RBRACE,
				EOF,
			},
		},
		{
			name:  "multiLineVDF",
			input: "\"root\"\n{\n\t\"key\"\n}",
			want: []TokenType{
				STRING,     // root
				WHITESPACE, // \n
				LBRACE,
				WHITESPACE, // \n
				WHITESPACE, // \t
				STRING,     // key
				WHITESPACE, // \n
				RBRACE,
				EOF,
			},
		},
		{
			name: "multiLineVDFRaw",
			input: `"root"
{
	"key" "value"
}`,
			want: []TokenType{
				STRING,     // root
				WHITESPACE, // \n
				LBRACE,
				WHITESPACE, // \n
				WHITESPACE, // \t
				STRING,     // key
				WHITESPACE,
				STRING,     // value
				WHITESPACE, // \n
				RBRACE,
				EOF,
			},
		},
		{
			name:  "escapeSequences",
			input: `"text\"with\"escapes"`,
			want: []TokenType{
				STRING,
				EOF,
			},
		},
		{
			name:  "numbersAsIdentifiers",
			input: `"123""456"`,
			want: []TokenType{
				STRING, // 123
				STRING, // 456
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
		{
			name:  "onlyComments",
			input: "// comment\n/* block comment */",
			want:  []TokenType{EOF},
		},
		{
			name:  "unterminatedString",
			input: `"unterminated`,
			want: []TokenType{
				STRING,
				EOF,
			},
		},
		{
			name:  "emptyInput",
			input: "",
			want:  []TokenType{EOF},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */)

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

func TestLexer_next_IgnoreWhitespace(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "whitespaceIgnored",
			input: `{ } \ `,
			want: []TokenType{
				LBRACE,
				RBRACE,
				ESCAPE,
				EOF,
			},
		},
		{
			name:  "multilineWithWhitespaceIgnored",
			input: "\t\n  \"key\"\n\t  \"value\"  \n",
			want: []TokenType{
				STRING, // key
				STRING, // value
				EOF,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), true /* ignoreWhitespace */)

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
				t.Fatalf("got %v, wanted %v", result, tc.want)
			}
		})
	}
}

func TestLexer_readString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{
			name:  "simpleString",
			input: `root"`,
			want:  "root",
		},
		{
			name:  "escapedString",
			input: `text\"with\"escapes"`,
			want:  "text\"with\"escapes",
		},
		{
			name:  "escapedStringWithNewline",
			input: `text\nwith\nescapes"`,
			want:  "text\nwith\nescapes",
		},
		{
			name:  "escapedStringWithTab",
			input: `text\twith\tescapes"`,
			want:  "text\twith\tescapes",
		},
		{
			name:  "escapedStringWithBackslash",
			input: `text\\with\\escapes"`,
			want:  "text\\with\\escapes",
		},
		{
			name:  "escapedStringWithCarriageReturn",
			input: `text\rwith\rescapes"`,
			want:  "text\rwith\rescapes",
		},
		{
			name:  "escapedStringWithBackslashAndDoubleQuote",
			input: `text\\\"with\\\"escapes"`,
			want:  "text\\\"with\\\"escapes",
		},
		{
			name:  "escapedStringWithBackslashAndNewline",
			input: `text\\\nwith\\\nescapes"`,
			want:  "text\\\nwith\\\nescapes",
		},
		{
			name:  "escapedStringWithBackslashAndTab",
			input: `text\\\twith\\\tescapes"`,
			want:  "text\\\twith\\\tescapes",
		},
		{
			name:    "unterminatedString",
			input:   `unterminated`,
			want:    "",
			wantErr: "unterminated string literal",
		},
		{
			name:    "incompleteEscapeSequence",
			input:   `text\`,
			want:    "",
			wantErr: "unterminated string literal",
		},
		{
			name:    "invalidEscapeSequence",
			input:   `text\xwith\xescapes"`,
			want:    "",
			wantErr: "invalid escape sequence: x",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */)

			got, err := lexer.readString()
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("readString() succeeded, wanted error %v", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Errorf("wanted error %v, got %v", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("readString(): %v", err)
			}

			if got != tc.want {
				t.Errorf("wanted %v, got %v", tc.want, got)
			}
		})
	}
}
