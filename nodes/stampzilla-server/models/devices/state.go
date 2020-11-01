package devices

type State map[string]interface{}

func (ds State) Clone() State {
	newState := make(State)
	for k, v := range ds {
		newState[k] = v
	}
	return newState
}

// Bool runs fn only if key is found in map and it is of type bool.
func (ds State) Bool(key string, fn func(bool)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(bool); ok {
			fn(v)
		}
	}
}

// Int runs fn only if key is found in map and it is of type int.
func (ds State) Int(key string, fn func(int64)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(int); ok {
			fn(int64(v))
		}
		if v, ok := v.(int64); ok {
			fn(v)
		}
	}
}

// Float runs fn only if key is found in map and it is of type int.
func (ds State) Float(key string, fn func(float64)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(float64); ok {
			fn(v)
		}
	}
}

// String runs fn only if key is found in map and it is of type int.
func (ds State) String(key string, fn func(string)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(string); ok {
			fn(v)
		}
	}
}

// Diff calculates the diff between 2 states. If key is missing in right but exists in left it will not be a diff.
func (ds State) Diff(right State) State {
	diff := make(State)
	for k, v := range ds {
		rv, ok := right[k]
		if !ok {
			// diff[k] = v
			continue
		}
		if ok && v != rv {
			diff[k] = rv
		}
	}

	for k, v := range right {
		if _, ok := ds[k]; !ok {
			diff[k] = v
		}
	}

	return diff
}

// Merge two states.
func (ds State) Merge(right State) State {
	diff := make(State)
	for k, v := range ds {
		diff[k] = v
	}
	for k, v := range right {
		diff[k] = v
	}
	return diff
}

func (ds State) MergeWith(right State) {
	for k, v := range right {
		ds[k] = v
	}
}

func (ds State) Equal(right State) bool {
	if (ds == nil) != (right == nil) {
		return false
	}

	if len(ds) != len(right) {
		return false
	}
	for k := range ds {
		if ds[k] != right[k] {
			return false
		}
	}
	return true
}
