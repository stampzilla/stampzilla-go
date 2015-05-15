package metrics

import (
	"reflect"
	"strconv"

	"github.com/fatih/structs"
)

type Metrics struct {
	current  map[string]string
	previous map[string]string
}

func New() *Metrics {
	return &Metrics{
		make(map[string]string),
		make(map[string]string),
	}
}

func (m *Metrics) Update(s interface{}) {
	m.current = structToMetrics(s)
	if len(m.previous) == 0 {
		return
	}

	for k, _ := range m.current {
		m.isDiff(k)
	}

}

func (m *Metrics) isDiff(s string) bool {
	if v, ok := m.previous[s]; ok {
		if v != m.current[s] {
			return true
		}
	}
	return false
}

func structToMetrics(s interface{}) map[string]string {
	st := structs.New(s)
	m := st.Map()
	flattened := make(map[string]string)
	flatten(m, st.Name(), &flattened)
	return flattened
}

func flatten(inputJSON map[string]interface{}, lkey string, flattened *map[string]string) {
	for rkey, value := range inputJSON {
		key := lkey + "." + rkey

		if structs.IsStruct(value) {
			//fmt.Println("Its a struct: ", value)
			value = structs.Map(value)
		}
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Type().Kind() == reflect.Map {
			//fmt.Println("its a map: ", value)
			out := make(map[string]interface{})
			for _, b := range reflectValue.MapKeys() {
				out[b.String()] = reflectValue.MapIndex(b).Interface()
			}
			value = out
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flatten(v, key, flattened)
		default:
			(*flattened)[key] = toString(v)
		}

	}
}

func toString(s interface{}) string {
	switch v := s.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	case bool:
		if v {
			return "1"
		}
		return "0"
	}
	return ""
}
