package devices

type State map[string]interface{}

func (ds State) Clone() State {
	newState := make(State)
	for k, v := range ds {
		newState[k] = v
	}
	return newState
}

// Bool runs fn only if key is found in map and it is of type bool
func (ds State) Bool(key string, fn func(bool)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(bool); ok {
			fn(v)
		}
	}
}

// Int runs fn only if key is found in map and it is of type int
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

// Float runs fn only if key is found in map and it is of type int
func (ds State) Float(key string, fn func(float64)) {
	if v, ok := ds[key]; ok {
		if v, ok := v.(float64); ok {
			fn(v)
		}
	}
}

func (ds State) Diff(right State) State {
	diff := make(State)
	for k, v := range ds {
		if v != right[k] {
			diff[k] = right[k]
		}
	}

	for k, v := range right {
		if _, ok := ds[k]; !ok {
			diff[k] = v
		}
	}

	return diff
}

// Merge two states
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
