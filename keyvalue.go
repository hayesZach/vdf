package vdf

type Document struct {
	Root *KeyValue
	Map  map[string]any
}

type KeyValue struct {
	Key   string
	Value any
}

// Children returns child KeyValues if Value is []*KeyValue, nil otherwise
func (k *KeyValue) Children() []*KeyValue {
	if children, ok := k.Value.([]*KeyValue); ok {
		return children
	}
	return nil
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

func iterateToKey(subvalues any, key string) *KeyValue {
	children, ok := subvalues.([]*KeyValue)
	if !ok {
		return nil
	}

	for _, child := range children {
		if child.Key == key {
			return child
		}
	}
	return nil
}

func (k *KeyValue) Get(key string, next ...string) *KeyValue {
	if key == "" {
		return nil
	}

	keys := []string{key}
	keys = append(keys, next...)

	kv := k
	for _, key := range keys {
		kv = iterateToKey(kv.Value, key)
		if kv == nil {
			return nil
		}
	}
	return kv
}

func (k *KeyValue) iterateToKeys(keys ...string) *KeyValue {
	if len(keys) == 0 {
		return k
	}

	children, ok := k.Value.([]*KeyValue)
	if !ok {
		return nil
	}

	for _, child := range children {
		if child.Key == keys[0] {
			if child.IsLeaf() {
				return child
			}
			return child.iterateToKeys(keys[1:]...)
		}
	}
	return nil
}

func (k *KeyValue) GetAll(key string, next ...string) []*KeyValue {
	if key == "" {
		return nil
	}

	children, ok := k.Value.([]*KeyValue)
	if !ok {
		return nil
	}

	keys := []string{key}
	keys = append(keys, next...)

	kvs := make([]*KeyValue, 0)
	for _, child := range children {
		var kv *KeyValue
		if child.Key == key {
			kv = child.iterateToKeys(next...)
			if kv != nil {
				kvs = append(kvs, kv)
			}
		}
	}
	return kvs
}

func (k *KeyValue) ToMap() map[string]any {
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
			val = child.ToMap()
		}

		if existing, exists := m[child.Key]; exists {
			existingMap, ok1 := existing.(map[string]any)
			valMap, ok2 := val.(map[string]any)

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
