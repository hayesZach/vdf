package vdf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func testTree() *KeyValue {
	return &KeyValue{
		Key: "root",
		Value: []*KeyValue{
			{Key: "name", Value: "test"},
			{Key: "version", Value: "1"},
			{Key: "section", Value: []*KeyValue{
				{Key: "a", Value: "1"},
				{Key: "b", Value: "2"},
				{Key: "a", Value: "3"},
			}},
			{Key: "section", Value: []*KeyValue{
				{Key: "c", Value: "4"},
			}},
			{Key: "deep", Value: []*KeyValue{
				{Key: "level1", Value: []*KeyValue{
					{Key: "level2", Value: []*KeyValue{
						{Key: "target", Value: "found"},
					}},
				}},
			}},
		},
	}
}

func TestKeyValue_ExtractUnusuals(t *testing.T) {
	t.Parallel()

	doc, err := parseFile("./static/items_game.txt")
	if err != nil {
		t.Fatal(err)
	}

	// Get all items
	items, err := doc.Root.GetAll("items")
	if err != nil {
		t.Fatal(err)
	}

	type pair struct {
		itemName string
		itemSet  string
	}

	// Extract knives and their likely corresponding case
	var pairs []pair
	for _, item := range items {
		children, err := item.Children()
		if err != nil {
			t.Fatal(err)
		}

		for _, child := range children {
			subMap := child.GetSubMap()

			itemName := subMap["name"].(string)
			if !strings.HasPrefix(itemName, "weapon_knife_") {
				continue
			}

			itemPrefab := subMap["prefab"].(string)

			if itemPrefab != "melee_unusual" {
				continue
			}

			if tag, err := child.Get("tags", "ItemSet", "tag_value"); err == nil {
				if tagValue, ok := tag.Value.(string); ok {
					pairs = append(pairs, pair{itemName, tagValue})
				}
			} else {
				pairs = append(pairs, pair{itemName, "unusual_revolving_list"})
			}
		}
	}

	fmt.Printf("%v", pairs)
}

func TestKeyValue_IsLeaf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		kv   *KeyValue
		want bool
	}{
		{
			name: "stringValueIsLeaf",
			kv:   &KeyValue{Key: "k", Value: "v"},
			want: true,
		},
		{
			name: "childrenValueIsNotLeaf",
			kv:   &KeyValue{Key: "k", Value: []*KeyValue{}},
			want: false,
		},
		{
			name: "nilValueIsNotLeaf",
			kv:   &KeyValue{Key: "k", Value: nil},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.kv.IsLeaf(); got != tc.want {
				t.Errorf("IsLeaf() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestKeyValue_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		kv   *KeyValue
		want string
	}{
		{
			name: "leafReturnsValue",
			kv:   &KeyValue{Key: "k", Value: "hello"},
			want: "hello",
		},
		{
			name: "nonLeafReturnsEmpty",
			kv:   &KeyValue{Key: "k", Value: []*KeyValue{}},
			want: "",
		},
		{
			name: "emptyStringLeaf",
			kv:   &KeyValue{Key: "k", Value: ""},
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.kv.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestKeyValue_Children(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		kv      *KeyValue
		want    []*KeyValue
		wantErr bool
	}{
		{
			name: "hasChildren",
			kv: &KeyValue{Key: "k", Value: []*KeyValue{
				{Key: "a", Value: "1"},
			}},
			want: []*KeyValue{
				{Key: "a", Value: "1"},
			},
		},
		{
			name:    "leafHasNoChildren",
			kv:      &KeyValue{Key: "k", Value: "v"},
			wantErr: true,
		},
		{
			name:    "nilValueHasNoChildren",
			kv:      &KeyValue{Key: "k", Value: nil},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.kv.Children()
			if tc.wantErr {
				if err == nil {
					t.Fatal("Children() succeeded, expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Children(): %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Children() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestKeyValue_Len(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		kv   *KeyValue
		want int
	}{
		{
			name: "withChildren",
			kv: &KeyValue{Key: "k", Value: []*KeyValue{
				{Key: "a", Value: "1"},
				{Key: "b", Value: "2"},
			}},
			want: 2,
		},
		{
			name: "emptyChildren",
			kv:   &KeyValue{Key: "k", Value: []*KeyValue{}},
			want: 0,
		},
		{
			name: "leafReturnsZero",
			kv:   &KeyValue{Key: "k", Value: "v"},
			want: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.kv.Len(); got != tc.want {
				t.Errorf("Len() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestKeyValue_At(t *testing.T) {
	t.Parallel()

	parent := &KeyValue{Key: "k", Value: []*KeyValue{
		{Key: "first", Value: "1"},
		{Key: "second", Value: "2"},
		{Key: "third", Value: "3"},
	}}

	testCases := []struct {
		name  string
		kv    *KeyValue
		index int
		want  *KeyValue
	}{
		{
			name:  "firstChild",
			kv:    parent,
			index: 0,
			want:  &KeyValue{Key: "first", Value: "1"},
		},
		{
			name:  "lastChild",
			kv:    parent,
			index: 2,
			want:  &KeyValue{Key: "third", Value: "3"},
		},
		{
			name:  "negativeIndex",
			kv:    parent,
			index: -1,
			want:  nil,
		},
		{
			name:  "outOfBounds",
			kv:    parent,
			index: 10,
			want:  nil,
		},
		{
			name:  "leafReturnsNil",
			kv:    &KeyValue{Key: "k", Value: "v"},
			index: 0,
			want:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.kv.At(tc.index)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("At(%d) mismatch (-want +got):\n%s", tc.index, diff)
			}
		})
	}
}

func TestKeyValue_Get(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name    string
		keys    []string
		want    *KeyValue
		wantErr bool
	}{
		{
			name: "noKeysReturnsSelf",
			keys: nil,
			want: tree,
		},
		{
			name: "singleKey",
			keys: []string{"name"},
			want: &KeyValue{Key: "name", Value: "test"},
		},
		{
			name: "nestedPath",
			keys: []string{"section", "a"},
			want: &KeyValue{Key: "a", Value: "1"},
		},
		{
			name: "deepPath",
			keys: []string{"deep", "level1", "level2", "target"},
			want: &KeyValue{Key: "target", Value: "found"},
		},
		{
			name:    "keyNotFound",
			keys:    []string{"nonexistent"},
			wantErr: true,
		},
		{
			name:    "pathThroughLeaf",
			keys:    []string{"name", "child"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tree.Get(tc.keys...)
			if tc.wantErr {
				if err == nil {
					t.Fatal("Get() succeeded, expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(): %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Get() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestKeyValue_GetAll(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name    string
		keys    []string
		want    []*KeyValue
		wantErr bool
	}{
		{
			name: "duplicateKeys",
			keys: []string{"section"},
			want: []*KeyValue{
				{Key: "section", Value: []*KeyValue{
					{Key: "a", Value: "1"},
					{Key: "b", Value: "2"},
					{Key: "a", Value: "3"},
				}},
				{Key: "section", Value: []*KeyValue{
					{Key: "c", Value: "4"},
				}},
			},
		},
		{
			name: "duplicateNestedKeys",
			keys: []string{"section", "a"},
			want: []*KeyValue{
				{Key: "a", Value: "1"},
				{Key: "a", Value: "3"},
			},
		},
		{
			name: "uniqueKey",
			keys: []string{"name"},
			want: []*KeyValue{
				{Key: "name", Value: "test"},
			},
		},
		{
			name: "noMatch",
			keys: []string{"nonexistent"},
			want: nil,
		},
		{
			name:    "pathThroughLeaf",
			keys:    []string{"name", "child"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tree.GetAll(tc.keys...)
			if tc.wantErr {
				if err == nil {
					t.Fatal("GetAll() succeeded, expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetAll(): %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetAll() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestKeyValue_Has(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name string
		key  string
		want bool
	}{
		{name: "existingKey", key: "name", want: true},
		{name: "duplicateKey", key: "section", want: true},
		{name: "missingKey", key: "nonexistent", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tree.Has(tc.key); got != tc.want {
				t.Errorf("Has(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestKeyValue_Count(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name string
		key  string
		want int
	}{
		{name: "uniqueKey", key: "name", want: 1},
		{name: "duplicateKey", key: "section", want: 2},
		{name: "missingKey", key: "nonexistent", want: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tree.Count(tc.key); got != tc.want {
				t.Errorf("Count(%q) = %d, want %d", tc.key, got, tc.want)
			}
		})
	}
}

func TestKeyValue_Keys(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name   string
		unique bool
		want   []string
	}{
		{
			name:   "allKeys",
			unique: false,
			want:   []string{"name", "version", "section", "section", "deep"},
		},
		{
			name:   "uniqueKeys",
			unique: true,
			want:   []string{"name", "version", "section", "deep"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tree.Keys(tc.unique)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Keys(%v) mismatch (-want +got):\n%s", tc.unique, diff)
			}
		})
	}
}

func TestKeyValue_Keys_Leaf(t *testing.T) {
	t.Parallel()

	leaf := &KeyValue{Key: "k", Value: "v"}
	got := leaf.Keys(false)
	if got != nil {
		t.Errorf("Keys() on leaf = %v, want nil", got)
	}
}

func TestKeyValue_Each(t *testing.T) {
	t.Parallel()

	tree := testTree()

	t.Run("visitsAllChildren", func(t *testing.T) {
		var keys []string
		tree.Each(func(child *KeyValue) bool {
			keys = append(keys, child.Key)
			return true
		})
		want := []string{"name", "version", "section", "section", "deep"}
		if diff := cmp.Diff(want, keys); diff != "" {
			t.Errorf("Each() visited keys mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("earlyStop", func(t *testing.T) {
		count := 0
		tree.Each(func(child *KeyValue) bool {
			count++
			return count < 2
		})
		if count != 2 {
			t.Errorf("Each() with early stop visited %d, want 2", count)
		}
	})

	t.Run("leafNoOp", func(t *testing.T) {
		leaf := &KeyValue{Key: "k", Value: "v"}
		called := false
		leaf.Each(func(child *KeyValue) bool {
			called = true
			return true
		})
		if called {
			t.Error("Each() on leaf should not call fn")
		}
	})
}

func TestKeyValue_Walk(t *testing.T) {
	t.Parallel()

	tree := &KeyValue{
		Key: "root",
		Value: []*KeyValue{
			{Key: "a", Value: "1"},
			{Key: "b", Value: []*KeyValue{
				{Key: "c", Value: "2"},
			}},
		},
	}

	t.Run("visitsAllNodes", func(t *testing.T) {
		var visited []string
		err := tree.Walk(func(path []string, node *KeyValue) error {
			visited = append(visited, node.Key)
			return nil
		})
		if err != nil {
			t.Fatalf("Walk(): %v", err)
		}

		want := []string{"root", "a", "b", "c"}
		if diff := cmp.Diff(want, visited); diff != "" {
			t.Errorf("Walk() visited mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("pathTracking", func(t *testing.T) {
		var paths [][]string
		err := tree.Walk(func(path []string, node *KeyValue) error {
			cp := make([]string, len(path))
			copy(cp, path)
			paths = append(paths, cp)
			return nil
		})
		if err != nil {
			t.Fatalf("Walk(): %v", err)
		}

		want := [][]string{
			{},
			{"root"},
			{"root"},
			{"root", "b"},
		}
		if diff := cmp.Diff(want, paths); diff != "" {
			t.Errorf("Walk() paths mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("earlyStop", func(t *testing.T) {
		count := 0
		_ = tree.Walk(func(path []string, node *KeyValue) error {
			count++
			if count < 2 {
				return nil
			}
			return errors.New("")
		})

		if count != 2 {
			t.Errorf("Walk() with early stop visited %d, want 2", count)
		}
	})
}

func TestKeyValue_Find(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name      string
		predicate func(*KeyValue) bool
		wantKey   string
		wantNil   bool
	}{
		{
			name:      "findDeepNode",
			predicate: func(kv *KeyValue) bool { return kv.Key == "target" },
			wantKey:   "target",
		},
		{
			name:      "findByValue",
			predicate: func(kv *KeyValue) bool { return kv.String() == "found" },
			wantKey:   "target",
		},
		{
			name:      "noMatch",
			predicate: func(kv *KeyValue) bool { return kv.Key == "nonexistent" },
			wantNil:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tree.Find(tc.predicate)
			if tc.wantNil {
				if got != nil {
					t.Errorf("Find() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("Find() returned nil")
			}
			if got.Key != tc.wantKey {
				t.Errorf("Find().Key = %q, want %q", got.Key, tc.wantKey)
			}
		})
	}
}

func TestKeyValue_FindAll(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name      string
		predicate func(*KeyValue) bool
		wantCount int
	}{
		{
			name:      "multipleMatches",
			predicate: func(kv *KeyValue) bool { return kv.Key == "a" },
			wantCount: 2,
		},
		{
			name:      "singleMatch",
			predicate: func(kv *KeyValue) bool { return kv.Key == "target" },
			wantCount: 1,
		},
		{
			name:      "allLeaves",
			predicate: func(kv *KeyValue) bool { return kv.IsLeaf() },
			wantCount: 7,
		},
		{
			name:      "noMatches",
			predicate: func(kv *KeyValue) bool { return kv.Key == "nonexistent" },
			wantCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tree.FindAll(tc.predicate)
			if len(got) != tc.wantCount {
				t.Errorf("FindAll() returned %d results, want %d", len(got), tc.wantCount)
			}
		})
	}
}

func TestKeyValue_FindByKey(t *testing.T) {
	t.Parallel()

	tree := testTree()

	testCases := []struct {
		name      string
		key       string
		wantCount int
	}{
		{name: "duplicateNestedKey", key: "a", wantCount: 2},
		{name: "uniqueDeepKey", key: "target", wantCount: 1},
		{name: "duplicateDirectKey", key: "section", wantCount: 2},
		{name: "missingKey", key: "nonexistent", wantCount: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tree.FindByKey(tc.key)
			if len(got) != tc.wantCount {
				t.Errorf("FindByKey(%q) returned %d results, want %d", tc.key, len(got), tc.wantCount)
			}
		})
	}
}
