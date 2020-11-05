package nibe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Parameter struct {
	Register string            `json:"register"`
	Factor   int               `json:"factor"`
	Size     string            `json:"size"`
	Mode     string            `json:"mode"`
	Title    string            `json:"titel"`
	Info     string            `json:"info"`
	Unit     string            `json:"unit"`
	Min      string            `json:"min"`
	Max      string            `json:"max"`
	Map      map[string]string `json:"map"`
}

var parameters = make(map[int]Parameter)

func (n *Nibe) LoadDefinitions(statikFS http.FileSystem, filename string) error {
	jsonFile, err := statikFS.Open(filename)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var decodedParams []Parameter
	json.Unmarshal(byteValue, &decodedParams)

	for _, param := range decodedParams {
		id, _ := strconv.Atoi(param.Register)
		parameters[id] = param
	}

	return nil
}

func (n *Nibe) Describe(reg uint16) (*Parameter, error) {
	if p, ok := parameters[int(reg)]; ok {
		return &p, nil
	}

	return nil, fmt.Errorf("Not found")
}
