package wirepusher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type WirePusherSender struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Action string `json:"action"`
}

func New(parameters json.RawMessage) *WirePusherSender {
	wp := &WirePusherSender{}

	json.Unmarshal(parameters, wp)

	return wp
}

func (wp *WirePusherSender) Trigger(dest []string, body string) error {
	var failure error
	for _, d := range dest {
		err := wp.notify(true, d, body)
		if err != nil {
			failure = err
		}
	}

	return failure
}

func (wp *WirePusherSender) Release(dest []string, body string) error {
	var failure error
	for _, d := range dest {
		err := wp.notify(false, d, body)
		if err != nil {
			failure = err
		}
	}

	return failure
}

func (wp *WirePusherSender) notify(trigger bool, dest string, body string) error {
	u, err := url.Parse("https://wirepusher.com/send")
	if err != nil {
		return err
	}

	event := "Triggered"
	if !trigger {
		event = "Released"
	}

	q := u.Query()

	q.Set("id", dest)
	q.Set("title", fmt.Sprintf("%s - %s", wp.Title, event))
	q.Set("message", body)
	q.Set("type", wp.Type)
	q.Set("action", wp.Action)

	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	//b, err := ioutil.ReadAll(resp.Body)
	//spew.Dump(b)

	return err
}

func (wp *WirePusherSender) Destinations() (error, map[string]string) {
	return fmt.Errorf("Not implemented"), nil
}