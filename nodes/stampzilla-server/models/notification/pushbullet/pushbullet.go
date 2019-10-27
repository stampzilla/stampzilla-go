package pushbullet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PushbulletSender struct {
	Token  string `json:"token"`
	server string
}

func New(parameters json.RawMessage) *PushbulletSender {
	pb := &PushbulletSender{
		server: "https://api.pushbullet.com",
	}

	json.Unmarshal(parameters, pb)

	return pb
}

func (pb *PushbulletSender) Trigger(dest []string, body string) error {
	values := map[string]string{
		"type":        "note",
		"body":        body,
		"title":       "stampzilla",
		"device_iden": dest[0],
	}
	postBody, _ := json.Marshal(values)

	req, err := http.NewRequest("POST", pb.server+"/v2/pushes", bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}

	req.Header.Add("Access-Token", pb.Token)
	req.Header.Add("Content-Type", `application/json`)

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

func (pb *PushbulletSender) Release(dest []string, body string) error {
	return fmt.Errorf("not implemented")
}

func (pb *PushbulletSender) Destinations() (map[string]string, error) {
	req, err := http.NewRequest("GET", pb.server+"/v2/devices", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Access-Token", pb.Token)
	req.Header.Add("Content-Type", `application/json`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	type Response struct {
		Devices []struct {
			Active         bool    `json:"active"`
			Iden           string  `json:"iden"`
			Created        float64 `json:"created"`
			Modified       float64 `json:"modified"`
			Type           string  `json:"type"`
			Kind           string  `json:"kind"`
			Nickname       string  `json:"nickname"`
			Manufacturer   string  `json:"manufacturer"`
			Model          string  `json:"model"`
			AppVersion     int     `json:"app_version"`
			Pushable       bool    `json:"pushable"`
			Icon           string  `json:"icon"`
			KeyFingerprint string  `json:"key_fingerprint"`
			Fingerprint    string  `json:"fingerprint,omitempty"`
			PushToken      string  `json:"push_token,omitempty"`
			HasSms         bool    `json:"has_sms,omitempty"`
			HasMms         bool    `json:"has_mms,omitempty"`
			RemoteFiles    string  `json:"remote_files,omitempty"`
		} `json:"devices"`
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response = Response{}
	err = json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}

	//spew.Dump(b)
	dest := make(map[string]string)
	for _, dev := range response.Devices {
		dest[dev.Iden] = dev.Nickname
	}

	return dest, nil
}
