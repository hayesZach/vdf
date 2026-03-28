package vdf

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testMap() Map {
	return Map{
		"name":    "test",
		"version": "1",
		"section": Map{
			"a": "1",
			"b": "2",
		},
		"deep": Map{
			"level1": Map{
				"level2": Map{
					"target": "found",
				},
			},
		},
	}
}

func TestMap_Get(t *testing.T) {
	t.Parallel()

	m := testMap()

	testCases := []struct {
		name   string
		keys   []string
		want   any
		wantOk bool
	}{
		{
			name:   "noKeysReturnsSelf",
			keys:   nil,
			want:   m,
			wantOk: true,
		},
		{
			name:   "singleKey",
			keys:   []string{"name"},
			want:   "test",
			wantOk: true,
		},
		{
			name:   "nestedPath",
			keys:   []string{"section", "a"},
			want:   "1",
			wantOk: true,
		},
		{
			name:   "deepPath",
			keys:   []string{"deep", "level1", "level2", "target"},
			want:   "found",
			wantOk: true,
		},
		{
			name:   "missingKey",
			keys:   []string{"nonexistent"},
			want:   nil,
			wantOk: false,
		},
		{
			name:   "pathThroughLeaf",
			keys:   []string{"name", "child"},
			want:   nil,
			wantOk: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := m.Get(tc.keys...)
			if ok != tc.wantOk {
				t.Errorf("Get() ok = %v, want %v", ok, tc.wantOk)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Get() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMap_GetString(t *testing.T) {
	t.Parallel()

	m := testMap()

	testCases := []struct {
		name string
		keys []string
		want string
	}{
		{
			name: "leafValue",
			keys: []string{"name"},
			want: "test",
		},
		{
			name: "nestedLeafValue",
			keys: []string{"section", "a"},
			want: "1",
		},
		{
			name: "mapValueReturnsEmpty",
			keys: []string{"section"},
			want: "",
		},
		{
			name: "missingKeyReturnsEmpty",
			keys: []string{"nonexistent"},
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := m.GetString(tc.keys...); got != tc.want {
				t.Errorf("GetString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMap_GetMap(t *testing.T) {
	t.Parallel()

	m := testMap()

	testCases := []struct {
		name string
		keys []string
		want Map
	}{
		{
			name: "directSubMap",
			keys: []string{"section"},
			want: Map{"a": "1", "b": "2"},
		},
		{
			name: "nestedSubMap",
			keys: []string{"deep", "level1"},
			want: Map{"level2": Map{"target": "found"}},
		},
		{
			name: "leafReturnsNil",
			keys: []string{"name"},
			want: nil,
		},
		{
			name: "missingKeyReturnsNil",
			keys: []string{"nonexistent"},
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.GetSubMap(tc.keys...)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GetMap() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMap_IsLeaf(t *testing.T) {
	t.Parallel()

	m := testMap()

	testCases := []struct {
		name string
		key  string
		want bool
	}{
		{name: "stringValue", key: "name", want: true},
		{name: "mapValue", key: "section", want: false},
		{name: "missingKey", key: "nonexistent", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := m.IsLeaf(tc.key); got != tc.want {
				t.Errorf("IsLeaf(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestMap_Keys(t *testing.T) {
	t.Parallel()

	m := Map{"b": "2", "a": "1", "c": "3"}
	got := m.Keys()
	if len(got) != 3 {
		t.Fatalf("Keys() returned %d keys, want 3", len(got))
	}

	has := make(map[string]bool)
	for _, k := range got {
		has[k] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !has[want] {
			t.Errorf("Keys() missing %q", want)
		}
	}
}

func TestMap_Keys_Empty(t *testing.T) {
	t.Parallel()

	m := Map{}
	got := m.Keys()
	if len(got) != 0 {
		t.Errorf("Keys() on empty map returned %d keys, want 0", len(got))
	}
}

func TestMap_Each(t *testing.T) {
	t.Parallel()

	m := Map{"a": "1", "b": "2", "c": "3"}

	testCases := []struct {
		name     string
		eachFunc func(m Map) any
		want     any
	}{
		{
			name: "visitsAll",
			eachFunc: func(m Map) any {
				visited := make(map[string]any)
				m.Each(func(key string, value any) bool {
					visited[key] = value
					return true
				})
				return Map(visited)
			},
			want: m,
		},
		{
			name: "earlyStop",
			eachFunc: func(m Map) any {
				count := 0
				m.Each(func(key string, value any) bool {
					count++
					return false
				})
				return count
			},
			want: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.eachFunc(m)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Each() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMap_Walk(t *testing.T) {
	t.Parallel()

	m := Map{
		"a": "1",
		"b": Map{
			"c": "2",
		},
	}

	testCases := []struct {
		name string
		walk func(m Map) any
		want any
	}{
		{
			name: "visitsAll",
			walk: func(m Map) any {
				visited := make(map[string]any)
				m.Walk(func(path []string, key string, value any) error {
					visited[key] = value
					return nil
				})
				return Map(visited)
			},
			want: Map{"a": "1", "b": Map{"c": "2"}, "c": "2"},
		},
		{
			name: "earlyStop",
			walk: func(m Map) any {
				count := 0
				m.Walk(func(path []string, key string, value any) error {
					count++
					return errors.New("non-nil error")
				})
				return count
			},
			want: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.walk(m)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Walk() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
