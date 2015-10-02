// Package main provides ...
package basenode

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"

	log "github.com/cihub/seelog"
	"github.com/pborman/uuid"
)

var config = &Config{}
var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "Server host/ip")
	flag.StringVar(&port, "port", "8282", "Server port")
}

func NewConfig() *Config {
	var config = &Config{}

	config.Host = host
	config.Port = port

	return config
}

func SetConfig(c *Config) {

	configFromFile := readConfigFromFile()
	c.Merge(configFromFile)

	if c.Uuid == "" {
		c.Uuid = uuid.New()
	}

	// Load logger
	logger, err := log.LoggerFromConfigAsFile("../logconfig.xml")
	if err != nil {
		testConfig := `
			<seelog type="sync" asyncinterval="1000" minlevel="trace">
				<outputs>
					<filter levels="trace">
						<console formatid="colored-trace"/>
					</filter>
					<filter levels="debug">
						<console formatid="colored-debug"/>
					</filter>
					<filter levels="info">
						<console formatid="colored-info"/>
					</filter>
					<filter levels="warn">
						<console formatid="colored-warn"/>
					</filter>
					<filter levels="error">
						<console formatid="colored-error"/>
					</filter>
					<filter levels="critical">
						<console formatid="colored-critical"/>
					</filter>
				</outputs>
				<formats>
					<format id="colored-trace"  format="%Date %Time %EscM(40)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-debug"  format="%%Date Time %EscM(45)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-info"  format="%Date %Time %EscM(46)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-warn"  format="%Date %Time %EscM(43)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-error"  format="%Date %Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
					<format id="colored-critical"  format="%Date %Time %EscM(41)%Level%EscM(49) - %File:%Line - %Msg%n%EscM(0)"/>
				</formats>
			</seelog>`

		logger, _ = log.LoggerFromConfigAsBytes([]byte(testConfig))
	}
	log.ReplaceLogger(logger)

	config = c

	saveConfigToFile()
}

func saveConfigToFile() {
	configFile, err := os.Create("config.json")
	if err != nil {
		log.Error("creating config file", err.Error())
	}

	log.Info("Save config: ", config)
	var out bytes.Buffer
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		log.Error("error marshal json", err)
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)
}

func readConfigFromFile() *Config {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Error("opening config file", err.Error())
	}

	config := &Config{}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Error("parsing config file", err.Error())
	}

	return config
}

type Config struct {
	Host string
	Port string
	Uuid string
	Node *json.RawMessage `json:"Node,omitempty"`
}

func (c *Config) NodeSpecific(i interface{}) error {
	return json.Unmarshal(*c.Node, &i)
}

func (c *Config) GetUuid() string {
	return c.Uuid
}

func (c *Config) Merge(c2 *Config) {

	if c.Host != c2.Host && c.Host == "localhost" && c2.Host != "" {
		c.Host = c2.Host
	}
	if c.Port != c2.Port && c.Port == "8282" && c2.Port != "" {
		c.Port = c2.Port
	}

	c.Uuid = c2.Uuid
	c.Node = c2.Node
}
