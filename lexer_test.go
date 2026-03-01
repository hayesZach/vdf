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
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, false /* usesEscapeSequences */)

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
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, false /* usesEscapeSequences */)

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
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, false /* usesEscapeSequences */)

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
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, false /* usesEscapeSequences */)

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

func TestLexer_peek(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		wantType TokenType
	}{
		{
			name:     "simpleString",
			input:    `"root"`,
			wantType: STRING,
		},
		{
			name:     "peekLBrace",
			input:    `{`,
			wantType: LBRACE,
		},
		{
			name:     "peekRBrace",
			input:    `}`,
			wantType: RBRACE,
		},
		{
			name:     "peekEscape",
			input:    `\n`,
			wantType: ESCAPE,
		},
		{
			name:     "peekWhitespace",
			input:    " test",
			wantType: WHITESPACE,
		},
		{
			name:     "peekIdentifier",
			input:    "abc",
			wantType: IDENTIFIER,
		},
		{
			name:     "emptyInput",
			input:    "",
			wantType: EOF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, false /* usesEscapeSequences */)

			token, err := lexer.peek()
			if err != nil {
				t.Fatalf("peek(): %v", err)
			}

			if token.Type != tc.wantType {
				t.Errorf("got type %v, wanted %v", token.Type, tc.wantType)
			}
		})
	}
}

func TestLexer_peek_DoesNotConsume(t *testing.T) {
	t.Parallel()

	input := `"root"`
	lexer := newLexer([]byte(input), false /* ignoreWhitespace */, false /* usesEscapeSequences */)

	expectedType := STRING

	// Peek multiple times to ensure that the token is not consumed
	for i := 0; i < 3; i++ {
		token, err := lexer.peek()
		if err != nil {
			t.Fatalf("peek(): %v", err)
		}
		if token.Type != expectedType {
			t.Errorf("got type %v, wanted %v", token.Type, expectedType)
		}
	}

	token, err := lexer.next()
	if err != nil {
		t.Fatalf("next(): %v", err)
	}
	if token.Type != expectedType {
		t.Errorf("got type %v, wanted %v", token.Type, expectedType)
	}
}

func TestLexer_next(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		input               string
		usesEscapeSequences bool
		want                []TokenType
		wantErr             string
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
			name:                "escapeSequences",
			input:               `"text\"with\"escapes"`,
			usesEscapeSequences: true,
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
			name:    "unterminatedString",
			input:   `"unterminated`,
			wantErr: "1:1: syntax error: unterminated string literal",
		},
		{
			name:  "emptyInput",
			input: "",
			want:  []TokenType{EOF},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, tc.usesEscapeSequences)

			result := make([]TokenType, 0)
			for {
				token, err := lexer.next()
				if tc.wantErr != "" {
					if err == nil {
						t.Fatalf("wanted error %v, got nil", tc.wantErr)
					}
					if err.Error() != tc.wantErr {
						t.Fatalf("wanted error %v, got %v", tc.wantErr, err)
					}
					return
				}
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
			lexer := newLexer([]byte(tc.input), true /* ignoreWhitespace */, false /* usesEscapeSequences */)

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
		name                string
		input               string
		usesEscapeSequences bool
		want                string
		wantErr             string
	}{
		{
			name:                "simpleString",
			input:               `root"`,
			usesEscapeSequences: false,
			want:                "root",
		},
		{
			name:                "escapedString",
			input:               `text\"with\"escapes"`,
			usesEscapeSequences: true,
			want:                "text\"with\"escapes",
		},
		{
			name:                "escapedStringWithNewline",
			input:               `text\nwith\nescapes"`,
			usesEscapeSequences: true,
			want:                "text\nwith\nescapes",
		},
		{
			name:                "escapedStringWithTab",
			input:               `text\twith\tescapes"`,
			usesEscapeSequences: true,
			want:                "text\twith\tescapes",
		},
		{
			name:                "escapedStringWithBackslash",
			input:               `text\\with\\escapes"`,
			usesEscapeSequences: true,
			want:                "text\\with\\escapes",
		},
		{
			name:                "escapedStringWithCarriageReturn",
			input:               `text\rwith\rescapes"`,
			usesEscapeSequences: true,
			want:                "text\rwith\rescapes",
		},
		{
			name:                "escapedStringWithBackslashAndDoubleQuote",
			input:               `text\\\"with\\\"escapes"`,
			usesEscapeSequences: true,
			want:                "text\\\"with\\\"escapes",
		},
		{
			name:                "escapedStringWithBackslashAndNewline",
			input:               `text\\\nwith\\\nescapes"`,
			usesEscapeSequences: true,
			want:                "text\\\nwith\\\nescapes",
		},
		{
			name:                "escapedStringWithBackslashAndTab",
			input:               `text\\\twith\\\tescapes"`,
			usesEscapeSequences: true,
			want:                "text\\\twith\\\tescapes",
		},
		{
			name:                "unterminatedString",
			input:               `unterminated`,
			usesEscapeSequences: false,
			want:                "",
			wantErr:             "unterminated string literal",
		},
		{
			name:                "incompleteEscapeSequence",
			input:               `text\`,
			usesEscapeSequences: true,
			want:                "",
			wantErr:             "unterminated string literal",
		},
		{
			name:                "invalidEscapeSequence",
			input:               `text\xwith\xescapes"`,
			usesEscapeSequences: true,
			want:                "",
			wantErr:             "invalid escape sequence: x",
		},
		{
			name:                "escapeSequencesNotAllowed",
			input:               `text\with\escapes"`,
			usesEscapeSequences: false,
			want:                "",
			wantErr:             "escape sequence not allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, tc.usesEscapeSequences)

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

func TestLexer_next_TokenPositions(t *testing.T) {
	t.Parallel()

	type expectedToken struct {
		Type   TokenType
		Lexeme string
		Line   int
		Column int
	}

	testCases := []struct {
		name             string
		input            string
		ignoreWhitespace bool
		want             []expectedToken
	}{
		{
			name:             "singleLineTokens",
			input:            `"key""value"`,
			ignoreWhitespace: true,
			want: []expectedToken{
				{STRING, "key", 1, 1},
				{STRING, "value", 1, 6},
				{EOF, "", 1, 13},
			},
		},
		{
			name:             "multiLineTokens",
			input:            "\"root\"\n{\n\"key\"\n}",
			ignoreWhitespace: true,
			want: []expectedToken{
				{STRING, "root", 1, 1},
				{LBRACE, "{", 2, 1},
				{STRING, "key", 3, 1},
				{RBRACE, "}", 4, 1},
				{EOF, "", 4, 2},
			},
		},
		{
			name: "multiLineWithIndentation",
			input: `"root"
{
	"key" "value"
}`,
			ignoreWhitespace: true,
			want: []expectedToken{
				{STRING, "root", 1, 1},
				{LBRACE, "{", 2, 1},
				{STRING, "key", 3, 2},
				{STRING, "value", 3, 8},
				{RBRACE, "}", 4, 1},
				{EOF, "", 4, 2},
			},
		},
		{
			name:             "whitespaceTokensIncluded",
			input:            "\"a\" \"b\"",
			ignoreWhitespace: false,
			want: []expectedToken{
				{STRING, "a", 1, 1},
				{WHITESPACE, " ", 1, 4},
				{STRING, "b", 1, 5},
				{EOF, "", 1, 8},
			},
		},
		{
			name:             "newlineAsWhitespace",
			input:            "\"a\"\n\"b\"",
			ignoreWhitespace: false,
			want: []expectedToken{
				{STRING, "a", 1, 1},
				{WHITESPACE, "\n", 1, 4},
				{STRING, "b", 2, 1},
				{EOF, "", 2, 4},
			},
		},
		{
			name:             "commentsSkipped",
			input:            "// comment\n\"key\"",
			ignoreWhitespace: true,
			want: []expectedToken{
				{STRING, "key", 2, 1},
				{EOF, "", 2, 6},
			},
		},
		{
			name:             "tokensAfterMultipleComments",
			input:            "// first\n// second\n\"key\"",
			ignoreWhitespace: true,
			want: []expectedToken{
				{STRING, "key", 3, 1},
				{EOF, "", 3, 6},
			},
		},
		{
			name:             "bracesOnSameLine",
			input:            "{}",
			ignoreWhitespace: true,
			want: []expectedToken{
				{LBRACE, "{", 1, 1},
				{RBRACE, "}", 1, 2},
				{EOF, "", 1, 3},
			},
		},
		{
			name:             "emptyInput",
			input:            "",
			ignoreWhitespace: true,
			want: []expectedToken{
				{EOF, "", 1, 1},
			},
		},
		{
			name: "complexVDF",
			input: `"root"
{
	"nested"
	{
		"key" "value"
	}
}`,
			ignoreWhitespace: true,
			want: []expectedToken{
				{STRING, "root", 1, 1},
				{LBRACE, "{", 2, 1},
				{STRING, "nested", 3, 2},
				{LBRACE, "{", 4, 2},
				{STRING, "key", 5, 3},
				{STRING, "value", 5, 9},
				{RBRACE, "}", 6, 2},
				{RBRACE, "}", 7, 1},
				{EOF, "", 7, 2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), tc.ignoreWhitespace, false /* usesEscapeSequences */)

			for i, expected := range tc.want {
				token, err := lexer.next()
				if err != nil {
					t.Fatalf("token %d: next(): %v", i, err)
				}

				if token.Type != expected.Type {
					t.Errorf("token %d: wanted type %v, got %v", i, expected.Type, token.Type)
				}
				if token.Lexeme != expected.Lexeme {
					t.Errorf("token %d: wanted lexeme %q, got %q", i, expected.Lexeme, token.Lexeme)
				}
				if token.Line != expected.Line {
					t.Errorf("token %d (%q): wanted line %d, got %d", i, token.Lexeme, expected.Line, token.Line)
				}
				if token.Column != expected.Column {
					t.Errorf("token %d (%q): wanted column %d, got %d", i, token.Lexeme, expected.Column, token.Column)
				}
			}
		})
	}
}

func TestLexer_calcLineAndColumn(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		input      string
		lineStarts []int
		pos        int
		wantLine   int
		wantCol    int
	}{
		{
			name:       "emptyLines",
			input:      "\n\n",
			pos:        1,
			lineStarts: []int{0, 1},
			wantLine:   2,
			wantCol:    1,
		},
		{
			name:       "firstLineIndexAtNewline",
			input:      "hello\nworld",
			pos:        5,
			lineStarts: []int{0, 6},
			wantLine:   1,
			wantCol:    6,
		},
		{
			name:       "firstLineFirstCharacter",
			input:      "hello",
			lineStarts: []int{0},
			pos:        0,
			wantLine:   1,
			wantCol:    1,
		},
		{
			name:       "firstLineMiddleCharacter",
			input:      "hello",
			lineStarts: []int{0},
			pos:        2,
			wantLine:   1,
			wantCol:    3,
		},
		{
			name:       "firstLineLastCharacter",
			input:      "hello",
			lineStarts: []int{0},
			pos:        4,
			wantLine:   1,
			wantCol:    5,
		},
		{
			name:       "secondLineFirstCharacter",
			input:      "hello\nworld",
			lineStarts: []int{0, 6},
			pos:        6,
			wantLine:   2,
			wantCol:    1,
		},
		{
			name:       "secondLineMiddleCharacter",
			input:      "hello\nworld",
			lineStarts: []int{0, 6},
			pos:        8,
			wantLine:   2,
			wantCol:    3,
		},
		{
			name:       "secondLineLastCharacter",
			input:      "hello\nworld",
			lineStarts: []int{0, 6},
			pos:        10,
			wantLine:   2,
			wantCol:    5,
		},
		{
			name:       "thirdLineFirstCharacter",
			input:      "hello\nworld\nhello",
			lineStarts: []int{0, 6, 12},
			pos:        12,
			wantLine:   3,
			wantCol:    1,
		},
		{
			name:       "manyLinesMiddleCharacter",
			input:      "abc\ndef\nghi\njkl\nmno\npqr",
			lineStarts: []int{0, 4, 8, 12, 16, 20},
			pos:        13,
			wantLine:   4,
			wantCol:    2,
		},
		{
			name:       "lastLineFirstCharacter",
			input:      "abc\ndef\nghi\njkl\nmno\npqr",
			lineStarts: []int{0, 4, 8, 12, 16, 20},
			pos:        20,
			wantLine:   6,
			wantCol:    1,
		},
		{
			name:       "unicodeCharacters",
			input:      "hello 世界\nworld",
			lineStarts: []int{0, 9},
			pos:        11,
			wantLine:   2,
			wantCol:    3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer([]byte(tc.input), false /* ignoreWhitespace */, false /* usesEscapeSequences */)
			lexer.lineStarts = tc.lineStarts
			lexer.pos = tc.pos
			line, col := lexer.calcLineAndColumn()
			if line != tc.wantLine {
				t.Errorf("wanted line %d, got %d", tc.wantLine, line)
			}
			if col != tc.wantCol {
				t.Errorf("wanted col %d, got %d", tc.wantCol, col)
			}
		})
	}
}
