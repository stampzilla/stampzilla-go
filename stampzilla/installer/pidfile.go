package installer

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type PidFile string

func (f *PidFile) String() string {
	return string(*f)
}

// Read the pidfile.
func (f *PidFile) Read() int {
	data, err := ioutil.ReadFile(string(*f))
	if err != nil {
		return 0
	}
	pidString := strings.Trim(string(data), "\n")
	pid, err := strconv.ParseInt(string(pidString), 0, 32)
	if err != nil {
		return 0
	}
	return int(pid)
}

// Write the pidfile.
func (f *PidFile) write(data int) error {
	err := ioutil.WriteFile(string(*f), []byte(strconv.Itoa(data)), 0660)
	if err != nil {
		return err
	}
	return nil
}

//Delete the pidfile.
func (f *PidFile) delete() bool {
	_, err := os.Stat(string(*f))
	if err != nil {
		return true
	}
	err = os.Remove(string(*f))
	if err == nil {
		return true
	}
	return false
}
