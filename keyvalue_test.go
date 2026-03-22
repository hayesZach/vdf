package vdf

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func loadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

func parseFile(path string) (*KeyValue, error) {
	data, err := loadFile(path)
	if err != nil {
		return nil, err
	}

	var kv KeyValue
	if err := Unmarshal(data, &kv); err != nil {
		return nil, err
	}
	return &kv, nil
}

func TestKeyValue_Traversal(t *testing.T) {
	t.Parallel()

	path := "./static/items_game.txt"
	kv, err := parseFile(path)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	index := buildIndex(kv.Value.([]*KeyValue))

	// prefabs := kv.GetAll("prefabs")
	// if prefabs == nil {
	// 	t.Fatalf("default rarity value is nil")
	// }

	fmt.Printf("%+v\n", index)
}
