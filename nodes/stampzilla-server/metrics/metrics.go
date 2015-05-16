package metrics

import (
	"reflect"
	"strconv"

	log "github.com/cihub/seelog"
	"github.com/fatih/structs"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/protocol"
)

type Logger interface {
	Log(key, value string)
}

type Metrics struct {
	current  map[string]string
	previous map[string]string
	loggers  []Logger
}

func New() *Metrics {
	return &Metrics{
		make(map[string]string),
		make(map[string]string),
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

	for k, v := range m.current {
		if m.isDiff(k) {
			log.Info("found diff. logging!")
			m.log(k, v)
		}
	}
	m.previous = m.current
}

func (m *Metrics) log(key, value string) {
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

func structToMetrics(s interface{}) map[string]string {
	log.Infof("%T", s)
	if node, ok := s.(protocol.Node); ok {
		st := structs.New(node)
		flattened := make(map[string]string)
		flatten(st.Map(), node.Uuid, &flattened)
		return flattened
	}
	return nil
}

func flatten(inputJSON map[string]interface{}, lkey string, flattened *map[string]string) {
	for rkey, value := range inputJSON {
		key := lkey + "_" + rkey

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
