package vdf

import (
	"testing"
)

func TestParser_Parse_SimpleKeyValue(t *testing.T) {
	t.Parallel()

	testString := `"root"
{
	"key" "value"
}`

	kv, err := Parse([]byte(testString))
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

	kv, err := Parse([]byte(testString))
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

	kv, err := Parse([]byte(testString))
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

	kv, err := Parse([]byte(testString))
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

	kv, err := Parse([]byte(testString))
	if err != nil {
		t.Fatalf("Parse(): = %v", err)
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

	kv, err := Parse([]byte(testString))
	if err != nil {
		t.Fatalf("Parse(): = %v", err)
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

	kv, err := Parse([]byte(testString))
	if err != nil {
		t.Fatalf("Parse(): = %v", err)
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

	kv, err := Parse([]byte(testString))
	if err != nil {
		t.Fatalf("Parse(): = %v", err)
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

	kv, err := Parse([]byte(testString))
	if err != nil {
		t.Fatalf("Parse(): = %v", err)
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

	kv, err := ParseWithOptions([]byte(testString), false /* ignoreWhitespace */, true /* usesEscapeSequences */)
	if err != nil {
		t.Fatalf("Parse(): = %v", err)
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
