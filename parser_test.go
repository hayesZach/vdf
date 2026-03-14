package vdf

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser_Parse_SimpleKeyValue(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"key" "value"
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}
	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(subValues))
	}

	if subValues[0].Key != "key" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key")
	}

	if subValues[0].Value != "value" {
		t.Errorf("got value %v, expected %q", subValues[0].Value, "value")
	}
}

func TestParser_Parse_DuplicateKeys(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"duplicate" "value1"
	"duplicate" "value2"
	"duplicate" "value3"
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Fatalf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 3 {
		t.Fatalf("got %d sub-values, expected 3", len(subValues))
	}

	// Verify all keys are "duplicate" with different values
	expectedValues := []string{"value1", "value2", "value3"}
	for i, sv := range subValues {
		if sv.Key != "duplicate" {
			t.Errorf("got subValues[%d].Key = %q, expected %q", i, sv.Key, "duplicate")
		}
		if sv.Value != expectedValues[i] {
			t.Errorf("got subValues[%d].Value = %q, expected %q", i, sv.Value, expectedValues[i])
		}
	}
}

func TestParser_Parse_NestedKeyValues(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"nested"
	{
		"key1" "value1"
		"key2" "value2"
	}
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 2", len(subValues))
	}

	if subValues[0].Key != "nested" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "nested")
	}

	nestedObj, ok := subValues[0].Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", subValues[0].Value)
	}
	if len(nestedObj) != 2 {
		t.Fatalf("got %d sub-values, expected 2", len(nestedObj))
	}

	if nestedObj[0].Key != "key1" {
		t.Errorf("got key %q, expected %q", nestedObj[0].Key, "key1")
	}
	if nestedObj[0].Value != "value1" {
		t.Errorf("got value %q, expected %q", nestedObj[0].Value, "value1")
	}

	if nestedObj[1].Key != "key2" {
		t.Errorf("got key %q, expected %q", nestedObj[1].Key, "key2")
	}
	if nestedObj[1].Value != "value2" {
		t.Errorf("got value %q, expected %q", nestedObj[1].Value, "value2")
	}
}

func TestParser_Parse_DeeplyNestedKeyValues(t *testing.T) {
	t.Parallel()

	testString := `"root"
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
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Fatalf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}
	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(subValues))
	}
	if subValues[0].Key != "nested" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "nested")
	}

	nestedObj, ok := subValues[0].Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", subValues[0].Value)
	}
	if len(nestedObj) != 1 {
		t.Errorf("got %d sub-values, expected 3", len(nestedObj))
	}
	if nestedObj[0].Key != "nested2" {
		t.Errorf("got key %q, expected %q", nestedObj[0].Key, "nested2")
	}

	nestedObj2, ok := nestedObj[0].Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", nestedObj[0].Value)
	}
	if len(nestedObj2) != 3 {
		t.Fatalf("got %d sub-values, expected 3", len(nestedObj2))
	}
	if nestedObj2[0].Key != "nested2_key1" {
		t.Errorf("got key %q, expected %q", nestedObj2[0].Key, "nested2_key1")
	}
	if nestedObj2[0].Value != "nested2_value1" {
		t.Errorf("got value %q, expected %q", nestedObj2[0].Value, "nested2_value1")
	}
	if nestedObj2[1].Key != "nested2_key2" {
		t.Errorf("got key %q, expected %q", nestedObj2[1].Key, "nested2_key2")
	}
	if nestedObj2[1].Value != "nested2_value2" {
		t.Errorf("got value %q, expected %q", nestedObj2[1].Value, "nested2_value2")
	}
	if nestedObj2[2].Key != "nested3" {
		t.Errorf("got key %q, expected %q", nestedObj2[2].Key, "nested3")
	}

	nestedObj3, ok := nestedObj2[2].Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", nestedObj2[2].Value)
	}
	if len(nestedObj3) != 2 {
		t.Fatalf("got %d sub-values, expected 2", len(nestedObj3))
	}
	if nestedObj3[0].Key != "nested3_key1" {
		t.Errorf("got key %q, expected %q", nestedObj3[0].Key, "nested3_key1")
	}
	if nestedObj3[0].Value != "nested3_value1" {
		t.Errorf("got value %q, expected %q", nestedObj3[0].Value, "nested3_value1")
	}
	if nestedObj3[1].Key != "nested3_key2" {
		t.Errorf("got key %q, expected %q", nestedObj3[1].Key, "nested3_key2")
	}
	if nestedObj3[1].Value != "nested3_value2" {
		t.Errorf("got value %q, expected %q", nestedObj3[1].Value, "nested3_value2")
	}
}

func TestParser_Parse_UnquotedIdentifiers(t *testing.T) {
	t.Parallel()

	testString := `root
{
	key value
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}
	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(subValues))
	}

	if subValues[0].Key != "key" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key")
	}
	if subValues[0].Value != "value" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value")
	}
}

func TestParser_Parse_MixedQuotedAndUnquoted(t *testing.T) {
	t.Parallel()

	testString := `root
	{
		"key1" value1
		key2 value2
		key3 "value3"
	}
	`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}
	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 3 {
		t.Fatalf("got %d sub-values, expected 3", len(subValues))
	}

	if subValues[0].Key != "key1" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key1")
	}
	if subValues[0].Value != "value1" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value1")
	}
	if subValues[1].Key != "key2" {
		t.Errorf("got key %q, expected %q", subValues[1].Key, "key2")
	}
	if subValues[1].Value != "value2" {
		t.Errorf("got value %q, expected %q", subValues[1].Value, "value2")
	}
	if subValues[2].Key != "key3" {
		t.Errorf("got key %q, expected %q", subValues[2].Key, "key3")
	}
	if subValues[2].Value != "value3" {
		t.Errorf("got value %q, expected %q", subValues[2].Value, "value3")
	}
}

func TestParser_Parse_MixedQuotedAndUnquotedWithWhitespace(t *testing.T) {
	t.Parallel()

	testString := `root
	{
		"key 1" value1
		key2 "value 2"
	}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 2 {
		t.Fatalf("got %d sub-values, expected 3", len(subValues))
	}

	if subValues[0].Key != "key 1" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key 1")
	}
	if subValues[0].Value != "value1" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value1")
	}
	if subValues[1].Key != "key2" {
		t.Errorf("got key %q, expected %q", subValues[1].Key, "key2")
	}
	if subValues[1].Value != "value 2" {
		t.Errorf("got value %q, expected %q", subValues[1].Value, "value 2")
	}
}

func TestParser_Parse_EmptyObject(t *testing.T) {
	t.Parallel()

	testString := `"root"
	{
	}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}
	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 0 {
		t.Fatalf("got %d sub-values, expected 0", len(subValues))
	}
}

func TestParser_Parse_WhitespaceHandling(t *testing.T) {
	t.Parallel()

	testString := `
	"root"
 {
	
	 "key1"		"value1"
	
		 "key2"  "value2"
	
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 2 {
		t.Fatalf("got %d sub-values, expected 2", len(subValues))
	}

	if subValues[0].Key != "key1" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key1")
	}
	if subValues[0].Value != "value1" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value1")
	}
	if subValues[1].Key != "key2" {
		t.Errorf("got key %q, expected %q", subValues[1].Key, "key2")
	}
	if subValues[1].Value != "value2" {
		t.Errorf("got value %q, expected %q", subValues[1].Value, "value2")
	}
}

func TestParser_Parse_EscapeSequences(t *testing.T) {
	t.Parallel()

	testString := `"root"
	{
		"\"key1\"" "\"value1\""
		"\nkey2" "\nvalue2"
		"key\t3" "value\t3"
	}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)), UseEscapeSequences())
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 3 {
		t.Fatalf("got %d sub-values, expected 3", len(subValues))
	}

	if subValues[0].Key != "\"key1\"" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "\"key1\"")
	}
	if subValues[0].Value != "\"value1\"" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "\"value1\"")
	}
	if subValues[1].Key != "\nkey2" {
		t.Errorf("got key %q, expected %q", subValues[1].Key, "\nkey2")
	}
	if subValues[1].Value != "\nvalue2" {
		t.Errorf("got value %q, expected %q", subValues[1].Value, "\nvalue2")
	}
	if subValues[2].Key != "key\t3" {
		t.Errorf("got key %q, expected %q", subValues[2].Key, "key\t3")
	}
	if subValues[2].Value != "value\t3" {
		t.Errorf("got value %q, expected %q", subValues[2].Value, "value\t3")
	}
}

func TestParser_Parse_UnquotedIdentifierFollowedByBrace(t *testing.T) {
	t.Parallel()

	testString := `root{
	"key" "value"
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	if kv.Key != "root" {
		t.Errorf("got key %q, expected %q", kv.Key, "root")
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(subValues))
	}

	if subValues[0].Key != "key" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key")
	}
	if subValues[0].Value != "value" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value")
	}
}

func TestParser_Parse_Error_MissingClosingBrace(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"key" "value"
`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Fatal("Parse() succeeded, expected error for missing closing brace")
	}

	if err != io.EOF {
		t.Errorf("got error %q, expected %q", err, io.EOF)
	}
}

func TestParser_Parse_Error_MissingOpeningBrace(t *testing.T) {
	t.Parallel()

	testString := `"root"
	"key" "value"
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Fatal("Parse() succeeded, expected error for missing opening brace")
	}

	wantErr := fmt.Sprintf("invalid token %q, expected LBRACE", STRING)
	if err.Error() != wantErr {
		t.Errorf("got error %q, expected %q", err, wantErr)
	}
}

func TestParser_Parse_Error_InvalidTokenAtRoot(t *testing.T) {
	t.Parallel()

	testString := `{
	"key" "value"
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Fatal("Parse() succeeded, expected error for invalid token at root")
	}

	wantErr := fmt.Sprintf("invalid token type %q for root key", LBRACE)
	if err.Error() != wantErr {
		t.Errorf("got error %q, expected %q", err, wantErr)
	}
}

func TestParser_Parse_Error_InvalidTokenInObject(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	{ "nested" "value" }
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Fatal("Parse() succeeded, expected error for invalid token in object")
	}

	wantErr := fmt.Sprintf("invalid token %q, expected STRING or IDENTIFIER", LBRACE)
	if err.Error() != wantErr {
		t.Errorf("got error %q, expected %q", err, wantErr)
	}
}

func TestParser_Parse_Error_UnterminatedString(t *testing.T) {
	t.Parallel()

	testString := `"root
{
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Fatal("Parse() succeeded, expected error for unterminated string")
	}

	wantErr := &SyntaxError{
		Line:    1,
		Column:  1,
		Message: "unterminated string literal",
	}

	if diff := cmp.Diff(wantErr, err); diff != "" {
		t.Error(diff)
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
			parser, err := NewParser(bytes.NewReader([]byte(tc.input)))
			if err != nil {
				t.Fatalf("NewParser(): %v", err)
			}

			kv, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse(): %v", err)
			}

			if diff := cmp.Diff(tc.expected, kv); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestParser_Parse_NestedObjectWithUnquotedKey(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	nested
	{
		"key" "value"
	}
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(subValues))
	}

	if subValues[0].Key != "nested" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "nested")
	}

	nestedObj, ok := subValues[0].Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", subValues[0].Value)
	}

	if len(nestedObj) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(nestedObj))
	}
}

func TestParser_Parse_MultipleRootSiblings(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"sibling1"
	{
		"key1" "value1"
	}
	"sibling2"
	{
		"key2" "value2"
	}
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 2 {
		t.Fatalf("got %d sub-values, expected 2", len(subValues))
	}

	if subValues[0].Key != "sibling1" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "sibling1")
	}
	if subValues[1].Key != "sibling2" {
		t.Errorf("got key %q, expected %q", subValues[1].Key, "sibling2")
	}
}

func TestParser_Parse_SpecialCharactersInStrings(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"key with spaces" "value with spaces"
	"key{brace" "value}brace"
	"key/slash" "value/slash"
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 3 {
		t.Fatalf("got %d sub-values, expected 3", len(subValues))
	}

	if subValues[0].Key != "key with spaces" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key with spaces")
	}
	if subValues[0].Value != "value with spaces" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value with spaces")
	}
	if subValues[1].Key != "key{brace" {
		t.Errorf("got key %q, expected %q", subValues[1].Key, "key{brace")
	}
	if subValues[1].Value != "value}brace" {
		t.Errorf("got value %q, expected %q", subValues[1].Value, "value}brace")
	}
}

func TestParser_Parse_UnquotedValueFollowedByBrace(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	nested{
		"key" "value"
	}
}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(subValues))
	}

	if subValues[0].Key != "nested" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "nested")
	}

	nestedObj, ok := subValues[0].Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", subValues[0].Value)
	}

	if len(nestedObj) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(nestedObj))
	}
}

func TestParser_Parse_UnquotedValueFollowedByRBrace(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"key" value}`

	parser, err := NewParser(bytes.NewReader([]byte(testString)))
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	kv, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse(): %v", err)
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
	}

	if len(subValues) != 1 {
		t.Fatalf("got %d sub-values, expected 1", len(subValues))
	}

	if subValues[0].Key != "key" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key")
	}
	if subValues[0].Value != "value" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value")
	}
}

func TestNewParser_Options(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		opts []Option
	}{
		{
			name: "emptyOptions",
			opts: []Option{},
		},
		{
			name: "useEscapeSequences",
			opts: []Option{UseEscapeSequences()},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewParser(bytes.NewReader(nil), tc.opts...)
			if err != nil {
				t.Fatalf("NewParser(): %v", err)
			}

			if parser == nil {
				t.Fatalf("parser is nil")
			}
		})
	}
}

func TestParser_parseUnquotedIdentifier(t *testing.T) {
	t.Parallel()

	t.Run("simple identifier", func(t *testing.T) {
		t.Parallel()

		testString := `root
{
}`
		parser, err := NewParser(bytes.NewReader([]byte(testString)))
		if err != nil {
			t.Fatalf("NewParser(): %v", err)
		}

		kv, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse(): %v", err)
		}

		if kv.Key != "root" {
			t.Errorf("got key %q, expected %q", kv.Key, "root")
		}
	})

	t.Run("identifier with numbers", func(t *testing.T) {
		t.Parallel()

		testString := `root123
{
}`
		parser, err := NewParser(bytes.NewReader([]byte(testString)))
		if err != nil {
			t.Fatalf("NewParser(): %v", err)
		}

		kv, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse(): %v", err)
		}

		if kv.Key != "root123" {
			t.Errorf("got key %q, expected %q", kv.Key, "root123")
		}
	})

	t.Run("identifier with underscores", func(t *testing.T) {
		t.Parallel()

		testString := `root_key
{
}`
		parser, err := NewParser(bytes.NewReader([]byte(testString)))
		if err != nil {
			t.Fatalf("NewParser(): %v", err)
		}

		kv, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse(): %v", err)
		}

		if kv.Key != "root_key" {
			t.Errorf("got key %q, expected %q", kv.Key, "root_key")
		}
	})
}

func TestParser_parseObject(t *testing.T) {
	t.Parallel()

	t.Run("empty object", func(t *testing.T) {
		t.Parallel()

		testString := `"root" {}`
		parser, err := NewParser(bytes.NewReader([]byte(testString)))
		if err != nil {
			t.Fatalf("NewParser(): %v", err)
		}

		kv, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse(): %v", err)
		}

		subValues, ok := kv.Value.([]*KeyValue)
		if !ok {
			t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
		}

		if len(subValues) != 0 {
			t.Fatalf("got %d sub-values, expected 0", len(subValues))
		}
	})

	t.Run("object with single string value", func(t *testing.T) {
		t.Parallel()

		testString := `"root" { "key" "value" }`
		parser, err := NewParser(bytes.NewReader([]byte(testString)))
		if err != nil {
			t.Fatalf("NewParser(): %v", err)
		}

		kv, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse(): %v", err)
		}

		subValues, ok := kv.Value.([]*KeyValue)
		if !ok {
			t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
		}

		if len(subValues) != 1 {
			t.Fatalf("got %d sub-values, expected 1", len(subValues))
		}

		if subValues[0].Value != "value" {
			t.Errorf("got value %q, expected %q", subValues[0].Value, "value")
		}
	})

	t.Run("object with nested object value", func(t *testing.T) {
		t.Parallel()

		testString := `"root" { "nested" { "inner" "value" } }`
		parser, err := NewParser(bytes.NewReader([]byte(testString)))
		if err != nil {
			t.Fatalf("NewParser(): %v", err)
		}

		kv, err := parser.Parse()
		if err != nil {
			t.Fatalf("Parse(): %v", err)
		}

		subValues, ok := kv.Value.([]*KeyValue)
		if !ok {
			t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
		}

		if len(subValues) != 1 {
			t.Fatalf("got %d sub-values, expected 1", len(subValues))
		}

		nestedObj, ok := subValues[0].Value.([]*KeyValue)
		if !ok {
			t.Fatalf("got Value of type %T, expected []*KeyValue", subValues[0].Value)
		}

		if len(nestedObj) != 1 {
			t.Fatalf("got %d sub-values, expected 1", len(nestedObj))
		}

		if nestedObj[0].Key != "inner" {
			t.Errorf("got key %q, expected %q", nestedObj[0].Key, "inner")
		}
	})
}
