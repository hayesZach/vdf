package vdf

type Map map[string]any

// Get walks a key path and returns the value at the end.
func (m Map) Get(keys ...string) (any, bool) {
	if len(keys) == 0 {
		return m, true
	}

	val, ok := m[keys[0]]
	if !ok {
		return nil, false
	}
	if len(keys) == 1 {
		return val, true
	}

	sub, ok := val.(Map)
	if !ok {
		return nil, false
	}
	return sub.Get(keys[1:]...)
}

// GetString walks a key path and returns the string value, or empty string if not found or not a string.
func (m Map) GetString(keys ...string) string {
	val, ok := m.Get(keys...)
	if !ok {
		return ""
	}
	s, _ := val.(string)
	return s
}

// GetSubMap walks a key path and returns the Map value, or nil if not found or not a map.
func (m Map) GetSubMap(keys ...string) Map {
	val, ok := m.Get(keys...)
	if !ok {
		return nil
	}
	sub, _ := val.(Map)
	return sub
}

// Has returns true if the key exists as a direct entry.
func (m Map) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Keys returns all direct keys.
func (m Map) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// IsLeaf returns true if the value at key is a string.
func (m Map) IsLeaf(key string) bool {
	val, ok := m[key]
	if !ok {
		return false
	}

	_, isStr := val.(string)
	return isStr
}

// Each iterates over direct entries, calling fn for each one.
// Return false from fn to stop early.
func (m Map) Each(fn func(key string, value any) bool) {
	for k, v := range m {
		if !fn(k, v) {
			return
		}
	}
}

type WalkFunc func(path []string, key string, value any) error

func (m Map) walk(path []string, fn WalkFunc) error {
	for k, v := range m {
		if err := fn(path, k, v); err != nil {
			return err
		}

		if sub, ok := v.(Map); ok {
			if err := sub.walk(append(path, k), fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Walk visits every node in the tree depth-first.
// The path slice contains the keys from the root down to (but not including) the current key.
// Return an error from fn to stop the walk entirely.
func (m Map) Walk(fn WalkFunc) error {
	return m.walk(nil, fn)
}
