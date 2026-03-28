package vdf

import "testing"

func TestUnmarshal_SimpleKeyValue(t *testing.T) {
	t.Parallel()

	data := []byte(`"root"
{
	"key" "value"
}`)

	var doc Document
	if err := Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}

	if doc.Root.Key != "root" {
		t.Errorf("got key %q, expected %q", doc.Root.Key, "root")
	}

	subValues, ok := doc.Root.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", doc.Root.Value)
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

func TestUnmarshal_WithOptions(t *testing.T) {
	t.Parallel()

	data := []byte(`"root"
{
	"key\t1" "value\n1"
}`)

	var doc Document
	if err := Unmarshal(data, &doc, UseEscapeSequences()); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}

	subValues, ok := doc.Root.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", doc.Root.Value)
	}

	if subValues[0].Key != "key\t1" {
		t.Errorf("got key %q, expected %q", subValues[0].Key, "key\t1")
	}
	if subValues[0].Value != "value\n1" {
		t.Errorf("got value %q, expected %q", subValues[0].Value, "value\n1")
	}
}

func TestUnmarshal_Error(t *testing.T) {
	t.Parallel()

	data := []byte(`"root
{
}`)

	var doc Document
	err := Unmarshal(data, &doc)
	if err == nil {
		t.Fatal("Unmarshal() succeeded, expected error")
	}
}
