package vdf

import "fmt"

type KeyValue struct {
	Key   string
	Value any
}

// Children returns child KeyValues if Value is []*KeyValue, nil otherwise
func (k *KeyValue) Children() ([]*KeyValue, error) {
	if children, ok := k.Value.([]*KeyValue); ok {
		return children, nil
	}
	return nil, fmt.Errorf("has no children")
}

// IsLeaf returns true if Value is a string (terminal node), false otherwise
func (k *KeyValue) IsLeaf() bool {
	if _, ok := k.Value.(string); ok {
		return true
	}
	return false
}

func (k *KeyValue) String() string {
	if k.IsLeaf() {
		return k.Value.(string)
	}
	return ""
}

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
			if len(keys) == 1 {
				return child, nil
			}
			return k.Get(keys[1:]...)
		}
	}
	return nil, fmt.Errorf("keys not found")
}

func (k *KeyValue) GetAll(keys ...string) ([]*KeyValue, error) {
	if len(keys) == 0 {
		return []*KeyValue{k}, nil
	}

	children, err := k.Children()
	if err != nil {
		return nil, err
	}

	results := make([]*KeyValue, 0)
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

func (k *KeyValue) Each(fn func(child *KeyValue)) {
	children, err := k.Children()
	if err != nil {
		return
	}
	for _, child := range children {
		fn(child)
	}
}

type VdfMap map[string]any

type Document struct {
	Root *KeyValue
	Map  VdfMap
}
