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

func parseFile(path string) (*Document, error) {
	data, err := loadFile(path)
	if err != nil {
		return nil, err
	}

	var doc Document
	if err := Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func TestKeyValue_Traversal(t *testing.T) {
	t.Parallel()

	path := "./static/items_game.txt"
	doc, err := parseFile(path)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	// prefabs := kv.GetAll("prefabs")
	// if prefabs == nil {
	// 	t.Fatalf("default rarity value is nil")
	// }

	fmt.Printf("%+v\n", doc)
}
