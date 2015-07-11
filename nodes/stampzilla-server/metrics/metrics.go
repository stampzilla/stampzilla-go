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
	previous map[string]interface{}
	loggers  []Logger
	queue    chan UpdatePackage
}

type UpdatePackage struct {
	Node  *serverprotocol.Node
	State map[string]interface{}
}

/* ----[ startup ]-------------------------------------------------*/

func New() *Metrics {
	return &Metrics{
		make(map[string]interface{}),
		nil,
		make(chan UpdatePackage, 100),
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
		m.update(s.Node, s.State)
	}
}

/* ----[ handle updates ]------------------------------------------*/

func (m *Metrics) Update(node *serverprotocol.Node) {
	current := structToMetrics(node.Uuid+"_Node_State", node.State)
	//current := structToMetrics(node.Uuid, node) //DO we want this instead? It will also logg Node.Uuid, Node.Elements etc... i think not!

	data := UpdatePackage{
		Node:  node,
		State: current,
	}

	m.queue <- data
}

func (m *Metrics) update(node *serverprotocol.Node, current map[string]interface{}) {
	if len(m.loggers) == 0 {
		return
	}

	if len(m.previous) == 0 { // No previous values exists, then use this one and commit all values
		m.previous = current

		// Force commit the first set off values
		for k, v := range current {
			m.log(k, v)
		}
		for _, l := range m.loggers {
			l.Commit(current) // TODO: is currently returning wrong value, should be the full struct
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
			l.Commit(current) // TODO: is currently returning wrong value, should be the full struct
		}
	}
	m.updatePrevious(current)
}

func (m *Metrics) log(key string, value interface{}) {
	for _, l := range m.loggers {
		l.Log(key, value)
	}
}

func (m *Metrics) updatePrevious(s map[string]interface{}) {
	for k, v := range s {
		m.previous[k] = v
	}
}

func (m *Metrics) isDiff(k string, v interface{}) bool {
	if oldValue, ok := m.previous[k]; ok {
		if oldValue == v {
			return false // Previous value found and is equal. No difference
		}
	}
	return true // Previous value not found, difference
}

/* ----[ converters ]----------------------------------------------*/

func structToMetrics(baseName string, s interface{}) map[string]interface{} {
	flattened := make(map[string]interface{})
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
			value = structs.Map(value)
		}
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Type().Kind() == reflect.Map {
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
