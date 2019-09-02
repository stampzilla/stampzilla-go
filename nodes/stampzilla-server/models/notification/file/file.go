package file

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type FileSender struct {
	Append    bool `json:"append"`
	Timestamp bool `json:"timestamp"`
}

func New(parameters json.RawMessage) *FileSender {
	f := &FileSender{}

	json.Unmarshal(parameters, f)

	return f
}

func (f *FileSender) Trigger(dest []string, body string) error {
	var failure error
	for _, d := range dest {
		err := f.notify(true, d, body)
		if err != nil {
			failure = err
		}
	}

	return failure
}

func (f *FileSender) Release(dest []string, body string) error {
	var failure error
	for _, d := range dest {
		err := f.notify(false, d, body)
		if err != nil {
			failure = err
		}
	}

	return failure
}

func (f *FileSender) notify(trigger bool, filename string, body string) error {
	mode := os.O_CREATE | os.O_WRONLY
	if f.Append {
		mode = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	}

	tf, err := os.OpenFile(filename, mode, 0644)
	if err != nil {
		return err
	}

	defer tf.Close()

	line := body
	if f.Timestamp {
		line = fmt.Sprintf("%s\t%s", time.Now().Format("2006-01-02 15:04:05"), line)
	}

	event := "Triggered"
	if !trigger {
		event = "Released"
	}
	line = fmt.Sprintf("%s\t%s\r\n", line, event)

	_, err = tf.WriteString(line)
	return err
}

func (f *FileSender) Destinations() (error, map[string]string) {
	return fmt.Errorf("Not implemented"), nil
}
