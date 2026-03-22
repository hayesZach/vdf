package vdf

import "fmt"

type KeyValue struct {
	Key   string
	Value any
}

// Children returns child KeyValues if Value is []*KeyValue.
func (k *KeyValue) Children() ([]*KeyValue, error) {
	if children, ok := k.Value.([]*KeyValue); ok {
		return children, nil
	}
	return nil, fmt.Errorf("has no children")
}

// IsLeaf returns true if Value is a string (terminal node).
func (k *KeyValue) IsLeaf() bool {
	_, ok := k.Value.(string)
	return ok
}

// String returns the string value if this is a leaf node, empty string otherwise.
func (k *KeyValue) String() string {
	if s, ok := k.Value.(string); ok {
		return s
	}
	return ""
}

// Len returns the number of direct children, 0 if leaf.
func (k *KeyValue) Len() int {
	children, err := k.Children()
	if err != nil {
		return 0
	}
	return len(children)
}

// At returns the child at the given index, or nil if out of bounds or leaf.
func (k *KeyValue) At(index int) *KeyValue {
	children, err := k.Children()
	if err != nil || index < 0 || index >= len(children) {
		return nil
	}
	return children[index]
}

// Get walks a key path, returning the first child matching each successive key.
func (k *KeyValue) Get(keys ...string) (*KeyValue, error) {
	if len(keys) == 0 {
		return k, nil
	}

	children, err := k.Children()
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		if child.Key == keys[0] {
			return child.Get(keys[1:]...)
		}
	}
	return nil, fmt.Errorf("keys not found")
}

// GetAll walks a key path, collecting all children that match at each level.
// Unlike Get, this returns every match for duplicate keys.
func (k *KeyValue) GetAll(keys ...string) ([]*KeyValue, error) {
	if len(keys) == 0 {
		return []*KeyValue{k}, nil
	}

	children, err := k.Children()
	if err != nil {
		return nil, err
	}

	var results []*KeyValue
	for _, child := range children {
		if child.Key == keys[0] {
			if len(keys) == 1 {
				results = append(results, child)
			} else {
				kvs, err := child.GetAll(keys[1:]...)
				if err != nil {
					return nil, err
				}
				results = append(results, kvs...)
			}
		}
	}

	return results, nil
}

// Has returns true if a direct child with the given key exists.
func (k *KeyValue) Has(key string) bool {
	children, err := k.Children()
	if err != nil {
		return false
	}
	for _, child := range children {
		if child.Key == key {
			return true
		}
	}
	return false
}

// Count returns the number of direct children with the given key.
func (k *KeyValue) Count(key string) int {
	children, err := k.Children()
	if err != nil {
		return 0
	}
	n := 0
	for _, child := range children {
		if child.Key == key {
			n++
		}
	}
	return n
}

// Keys returns the keys of all direct children. If unique is true,
// duplicates are omitted (first occurrence order is preserved).
func (k *KeyValue) Keys(unique bool) []string {
	children, err := k.Children()
	if err != nil {
		return nil
	}

	var keys []string
	if !unique {
		for _, child := range children {
			keys = append(keys, child.Key)
		}
		return keys
	}

	seen := make(map[string]struct{})
	for _, child := range children {
		if _, ok := seen[child.Key]; !ok {
			seen[child.Key] = struct{}{}
			keys = append(keys, child.Key)
		}
	}
	return keys
}

// Each iterates over direct children, calling fn for each one.
// Return false from fn to stop early.
func (k *KeyValue) Each(fn func(child *KeyValue) bool) {
	children, err := k.Children()
	if err != nil {
		return
	}
	for _, child := range children {
		if !fn(child) {
			return
		}
	}
}

// Walk visits every node in the tree depth-first (pre-order).
// The path slice contains the keys from the root down to (but not including) the current node.
// Return false from fn to stop the walk entirely.
func (k *KeyValue) Walk(fn func(path []string, node *KeyValue) bool) {
	k.walk(nil, fn)
}

func (k *KeyValue) walk(path []string, fn func([]string, *KeyValue) bool) bool {
	if !fn(path, k) {
		return false
	}
	children, err := k.Children()
	if err != nil {
		return true
	}
	childPath := append(path, k.Key)
	for _, child := range children {
		if !child.walk(childPath, fn) {
			return false
		}
	}
	return true
}

// Find returns the first node in a depth-first search that matches the predicate, or nil.
func (k *KeyValue) Find(predicate func(*KeyValue) bool) *KeyValue {
	var result *KeyValue
	k.Walk(func(_ []string, node *KeyValue) bool {
		if predicate(node) {
			result = node
			return false
		}
		return true
	})
	return result
}

// FindAll returns all nodes in a depth-first search that match the predicate.
func (k *KeyValue) FindAll(predicate func(*KeyValue) bool) []*KeyValue {
	var results []*KeyValue
	k.Walk(func(_ []string, node *KeyValue) bool {
		if predicate(node) {
			results = append(results, node)
		}
		return true
	})
	return results
}

// FindByKey does a deep search for all nodes with the given key.
func (k *KeyValue) FindByKey(key string) []*KeyValue {
	return k.FindAll(func(node *KeyValue) bool {
		return node.Key == key
	})
}

// GetSubMap converts the tree to a map representation.
// Duplicate keys with object values are merged; duplicate leaf keys keep the last value.
func (k *KeyValue) GetSubMap() VdfMap {
	m := make(map[string]any)

	children, ok := k.Value.([]*KeyValue)
	if !ok {
		return m
	}

	for _, child := range children {
		var val any
		if child.IsLeaf() {
			val = child.Value
		} else {
			val = child.GetSubMap()
		}

		if existing, exists := m[child.Key]; exists {
			existingMap, ok1 := existing.(VdfMap)
			valMap, ok2 := val.(VdfMap)

			if ok1 && ok2 {
				mergeMaps(existingMap, valMap)
			} else {
				m[child.Key] = val
			}
		} else {
			m[child.Key] = val
		}
	}
	return m
}

type VdfMap map[string]any

type Document struct {
	Root *KeyValue
	Map  VdfMap
}
