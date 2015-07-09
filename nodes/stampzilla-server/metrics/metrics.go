package metrics

import (
	"encoding/json"
	"reflect"
	"strconv"

	log "github.com/cihub/seelog"
)

type Logger interface {
	Log(key string, value interface{})
	Commit(node interface{})
}

type Metrics struct {
	previous map[string]interface{}
	loggers  []Logger
	queue    chan []byte
}

func New() *Metrics {
	return &Metrics{
		make(map[string]interface{}),
		nil,
		make(chan []byte, 100),
	}
}
func (m *Metrics) AddLogger(l Logger) {
	log.Infof("Adding logger: %T\n", l)
	m.loggers = append(m.loggers, l)
}

func (m *Metrics) Start() {
	go m.worker()
}

func (m *Metrics) worker() {
	for s := range m.queue {
		var data interface{}
		err := json.Unmarshal(s, &data)

		if err != nil {
			log.Warn("Failed to unmarshal metrics update: ", err)
			continue
		}

		m.update(data)
	}
}

func (m *Metrics) Update(s interface{}) {
	data, err := json.Marshal(s)
	if err != nil {
		log.Warn("Failed to marshal metrics update: ", err)
		return
	}

	m.queue <- data
}

func (m *Metrics) update(s interface{}) {
	if len(m.loggers) == 0 {
		return
	}

	current := structToMetrics(s)
	if len(m.previous) == 0 { // No previous values exists, then use this one and commit all values
		m.previous = current

		// Force commit the first set off values
		for k, v := range current {
			m.log(k, v)
		}
		for _, l := range m.loggers {
			l.Commit(s)
		}
		return
	}

	changed := false
	for k, v := range current {
		if m.isDiff(k, v) {
			m.log(k, v)
			changed = true
		}
	}

	if changed {
		for _, l := range m.loggers {
			l.Commit(s)
		}
	}
	m.updatePrevious(current)
}

func (m *Metrics) updatePrevious(s map[string]interface{}) {
	for k, v := range s {
		m.previous[k] = v
	}
}

func (m *Metrics) log(key string, value interface{}) {
	for _, l := range m.loggers {
		l.Log(key, value)
	}
}

func (m *Metrics) isDiff(k string, v interface{}) bool {
	if oldValue, ok := m.previous[k]; ok {
		if oldValue != v {
			return true
		}
	}
	return false
}

func structToMetrics(s interface{}) map[string]interface{} {
	flattened := make(map[string]interface{})
	baseName := ""

	if data, ok := s.(map[string]interface{}); ok {
		if uuid, ok := data["Uuid"].(string); ok {
			baseName = uuid
		}

		flatten(data, baseName, &flattened)
	}
	return flattened
}

func flatten(inputJSON map[string]interface{}, lkey string, flattened *map[string]interface{}) {
	for rkey, value := range inputJSON {
		key := lkey + "_" + rkey
		if lkey == "" {
			key = rkey
		}

		if value == nil {
			continue
		}

		//if structs.IsStruct(value) {
		//value = structs.Map(value)
		//}
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Type().Kind() == reflect.Map {
			out := make(map[string]interface{})
			for _, b := range reflectValue.MapKeys() {
				out[b.String()] = reflectValue.MapIndex(b).Interface()
			}
			value = out
		}

		switch v := value.(type) {
		case *map[string]interface{}:
			flatten(*v, key, flattened)
		case map[string]interface{}:
			flatten(v, key, flattened)
		default:
			(*flattened)[key] = cast(v)
		}

	}
}

func cast(s interface{}) interface{} {
	switch v := s.(type) {
	case int:
		return v
		//return strconv.Itoa(v)
	case float64:
		//return strconv.FormatFloat(v, 'f', -1, 64)
		return v
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		return v
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return ""
}
