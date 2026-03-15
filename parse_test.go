package vdf

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser_Parse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		input              string
		useEscapeSequences bool
		expected           *KeyValue
	}{
		{
			name: "emptyObject",
			input: `"root"
					{
					}`,
			expected: &KeyValue{
				Key:   "root",
				Value: []*KeyValue{},
			},
		},
		{
			name: "simpleKeyValue",
			input: `"root"
					{
						"key" "value"
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key", Value: "value"},
				},
			},
		},
		{
			name: "duplicateKeys",
			input: `"root"
					{
						"duplicate" "value1"
						"duplicate" "value2"
						"duplicate" "value3"
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "duplicate", Value: "value1"},
					{Key: "duplicate", Value: "value2"},
					{Key: "duplicate", Value: "value3"},
				},
			},
		},
		{
			name: "handleInconsistentWhitespace",
			input: `
	"root"
 {
	
	 "key1"		"value1"
	
		 "key2"  "value2"
	
}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
			},
		},
		{
			name: "unquotedIdentifier",
			input: `root
					{
						key value
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key", Value: "value"},
				},
			},
		},
		{
			name: "mixedQuotedAndUnquoted",
			input: `root
					{
						"key1" value1
						key2 value2
						key3 "value3"
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
					{Key: "key3", Value: "value3"},
				},
			},
		},
		{
			name: "noSpaceBetweenKeysAndValues",
			input: `root{
				"key1"value1
				key2 value2
				key3"value3"
			}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
					{Key: "key3", Value: "value3"},
				},
			},
		},
		{
			name: "unquotedFollowedByRBrace",
			input: `root{
				key value
			}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key", Value: "value"},
				},
			},
		},
		{
			name: "unquotedFollowedByBrace",
			input: `root{
				key value}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key", Value: "value"},
				},
			},
		},
		{
			name: "escapeSequences",
			input: `"root"
					{
						"\"key1\"" "\"value1\""
						"\nkey2" "\nvalue2"
						"key\t3" "value\t3"
					}`,
			useEscapeSequences: true,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "\"key1\"", Value: "\"value1\""},
					{Key: "\nkey2", Value: "\nvalue2"},
					{Key: "key\t3", Value: "value\t3"},
				},
			},
		},
		{
			name: "quotedControlCharacters",
			input: `"root"
					{
						"some{\"}key" "value"
					}`,
			useEscapeSequences: true,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "some{\"}key", Value: "value"},
				},
			},
		},
		{
			name: "nested",
			input: `"root"
					{
						"nested"
						{
							"key1" "value1"
							"key2" "value2"
						}
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{
						Key: "nested",
						Value: []*KeyValue{
							{Key: "key1", Value: "value1"},
							{Key: "key2", Value: "value2"},
						},
					},
				},
			},
		},
		{
			name: "deeplyNested",
			input: `"root"
					{
						"nested"
						{
							"nested2"
							{
								"nested2_key1" "nested2_value1"
								"nested2_key2" "nested2_value2"
								"nested3"
								{
									"nested3_key1" "nested3_value1"
									"nested3_key2" "nested3_value2"
								}
							}
						}
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{
						Key: "nested",
						Value: []*KeyValue{
							{
								Key: "nested2",
								Value: []*KeyValue{
									{Key: "nested2_key1", Value: "nested2_value1"},
									{Key: "nested2_key2", Value: "nested2_value2"},
									{
										Key: "nested3",
										Value: []*KeyValue{
											{Key: "nested3_key1", Value: "nested3_value1"},
											{Key: "nested3_key2", Value: "nested3_value2"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "nestedUnquoted",
			input: `"root"
					{
						nested
						{
							key value
						}
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{
						Key: "nested",
						Value: []*KeyValue{
							{Key: "key", Value: "value"},
						},
					},
				},
			},
		},
		{
			name: "multipleRootSiblings",
			input: `"root"
					{
						"sibling1"
						{
							"key1" "value1"
						}
						"sibling2"
						{
							"key2" "value2"
						}
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{
						Key: "sibling1",
						Value: []*KeyValue{
							{Key: "key1", Value: "value1"},
						},
					},
					{
						Key: "sibling2",
						Value: []*KeyValue{
							{Key: "key2", Value: "value2"},
						},
					},
				},
			},
		},
		{
			name: "specialCharactersInStrings",
			input: `"root"
					{
						"key with spaces" "value with spaces"
						"key{brace" "value}brace"
						"key/slash" "value/slash"
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{
						Key:   "key with spaces",
						Value: "value with spaces",
					},
					{
						Key:   "key{brace",
						Value: "value}brace",
					},
					{
						Key:   "key/slash",
						Value: "value/slash",
					},
				},
			},
		},
		{
			name: "unquotedSpecialCharacters",
			input: `root123!@#$%^&*(){
						some!key some%/value
					}`,
			expected: &KeyValue{
				Key: "root123!@#$%^&*()",
				Value: []*KeyValue{
					{Key: "some!key", Value: "some%/value"},
				},
			},
		},
		{
			name:  "noWhitespace",
			input: `root{key"value"key2"value2"}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key", Value: "value"},
					{Key: "key2", Value: "value2"},
				},
			},
		},
		{
			name:  "noWhitespaceNested",
			input: `root{nested{key"value"key2"value2"}}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{
						Key: "nested",
						Value: []*KeyValue{
							{Key: "key", Value: "value"},
							{Key: "key2", Value: "value2"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &parser{lexer: newLexer([]byte(tc.input), tc.useEscapeSequences)}
			kv, err := p.parse()
			if err != nil {
				t.Fatalf("Parse(): %v", err)
			}

			if diff := cmp.Diff(tc.expected, kv); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestParser_Parse_Errors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			name: "missingClosingBrace",
			input: `"root"
					{
						"key" "value"
					`,
			expectedError: &SyntaxError{
				Line:    4,
				Column:  6,
				Message: "unexpected EOF",
			},
		},
		{
			name: "missingOpeningBrace",
			input: `"root"
	"key" "value"
}`,
			expectedError: &SyntaxError{
				Line:    2,
				Column:  2,
				Message: fmt.Sprintf("invalid token %s, expected LBRACE", STRING.String()),
			},
		},
		{
			name: "invalidTokenAtRoot",
			input: `{
	"key" "value"
}`,
			expectedError: &SyntaxError{
				Line:    1,
				Column:  1,
				Message: fmt.Sprintf("invalid token type %s for root key", LBRACE),
			},
		},
		{
			name: "invalidTokenInObject",
			input: `"root"
{
	{ "nested" "value" }
}`,
			expectedError: &SyntaxError{
				Line:    3,
				Column:  2,
				Message: fmt.Sprintf("invalid token %s, expected STRING or IDENTIFIER", LBRACE),
			},
		},
		{
			name: "unterminatedString",
			input: `"root
{
}`,
			expectedError: &SyntaxError{
				Line:    3,
				Column:  2,
				Message: "unterminated string literal",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &parser{lexer: newLexer([]byte(tc.input), false)}
			_, err := p.parse()
			if err == nil {
				t.Fatalf("Parse() succeeded, expected error %v", tc.expectedError)
			}

			if diff := cmp.Diff(tc.expectedError, err); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestParser_Parse_Comments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected *KeyValue
	}{
		{
			name: "fullLineComments",
			input: `// This is a comment
					"root"
					{
						// Another comment
						"key1" "value1"
						/* block comment */
						"key2" "value2"
					}`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
			},
		},
		{
			name: "commentsAtEnd",
			input: `// This is a comment
					"root" // next comment
					{
						"key" "value" // another comment
					} /* final comment`,
			expected: &KeyValue{
				Key: "root",
				Value: []*KeyValue{
					{Key: "key", Value: "value"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &parser{lexer: newLexer([]byte(tc.input), false)}
			kv, err := p.parse()
			if err != nil {
				t.Fatalf("parse(): %v", err)
			}

			if diff := cmp.Diff(tc.expected, kv); diff != "" {
				t.Error(diff)
			}
		})
	}
}
