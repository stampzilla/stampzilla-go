package metrics

import (
	"reflect"
	"strconv"

	log "github.com/cihub/seelog"
	"github.com/fatih/structs"
	serverprotocol "github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type Logger interface {
	Log(key string, value interface{})
	Commit(node interface{})
}

type Metrics struct {
	current  map[string]interface{}
	previous map[string]interface{}
	loggers  []Logger
}

func New() *Metrics {
	return &Metrics{
		make(map[string]interface{}),
		make(map[string]interface{}),
		nil,
	}
}
func (m *Metrics) AddLogger(l Logger) {
	log.Infof("Adding logger: %T\n", l)
	m.loggers = append(m.loggers, l)
}

func (m *Metrics) Update(s interface{}) {
	if len(m.loggers) == 0 {
		return
	}

	m.current = structToMetrics(s)
	if len(m.previous) == 0 {
		m.previous = m.current
		return
	}

	changed := false
	for k, v := range m.current {
		if m.isDiff(k) {
			log.Info("found diff. logging!")
			m.log(k, v)
			changed = true
		}
	}

	if changed {
		for _, l := range m.loggers {
			l.Commit(s)
		}
	}
	m.previous = m.current
}

func (m *Metrics) log(key string, value interface{}) {
	for _, l := range m.loggers {
		l.Log(key, value)
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

func structToMetrics(s interface{}) map[string]interface{} {
	flattened := make(map[string]interface{})
	baseName := ""
	if node, ok := s.(serverprotocol.Node); ok {
		//st := structs.New(node)
		baseName = node.Uuid
	}
	flatten(structs.Map(s), baseName, &flattened)
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
