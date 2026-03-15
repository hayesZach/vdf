package vdf

import "testing"

func TestUnmarshal_SimpleKeyValue(t *testing.T) {
	t.Parallel()

	data := []byte(`"root"
{
	"key" "value"
}`)

	var kv KeyValue
	if err := Unmarshal(data, &kv); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
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

func TestUnmarshal_WithOptions(t *testing.T) {
	t.Parallel()

	data := []byte(`"root"
{
	"key\t1" "value\n1"
}`)

	var kv KeyValue
	if err := Unmarshal(data, &kv, UseEscapeSequences()); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}

	subValues, ok := kv.Value.([]*KeyValue)
	if !ok {
		t.Fatalf("got Value of type %T, expected []*KeyValue", kv.Value)
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

	var kv KeyValue
	err := Unmarshal(data, &kv)
	if err == nil {
		t.Fatal("Unmarshal() succeeded, expected error")
	}
}
