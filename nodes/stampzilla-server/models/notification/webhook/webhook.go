package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WebhookSender struct {
	Method string `json:"method"`
}

func New(parameters json.RawMessage) *WebhookSender {
	ws := &WebhookSender{}

	json.Unmarshal(parameters, ws)

	return ws
}

func (ws *WebhookSender) Trigger(dest []string, body string) error {
	var failure error
	for _, url := range dest {
		err := ws.notify(true, url, body)
		if err != nil {
			failure = err
		}
	}

	return failure
}

func (ws *WebhookSender) Release(dest []string, body string) error {
	var failure error
	for _, url := range dest {
		err := ws.notify(false, url, body)
		if err != nil {
			failure = err
		}
	}

	return failure
}

func (ws *WebhookSender) notify(trigger bool, url string, body string) error {
	req, err := http.NewRequest(ws.Method, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// b, err := ioutil.ReadAll(resp.Body)
	// spew.Dump(b)

	return err
}

func (ws *WebhookSender) Destinations() (map[string]string, error) {
	return nil, fmt.Errorf("not implemented")
}
