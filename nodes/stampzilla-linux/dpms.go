package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/models/devices"
)

func startMonitorDpms() {
	c1 := exec.Command("ls", "/tmp/.X11-unix")
	c2 := exec.Command("tr", "X", ":")
	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()
	result := strings.Split(strings.TrimSpace(b2.String()), "\n")

	for _, screen := range result {
		go monitorDpms(screen)
	}
}

func monitorDpms(screen string) {
	dev := &devices.Device{
		Name:   "Monitor " + screen,
		ID:     devices.ID{ID: "monitor" + screen},
		Online: true,
		Traits: []string{"OnOff"},
		State: devices.State{
			"on": false,
		},
	}

	re := regexp.MustCompile("Monitor is (in )?([^ \n]+)")

	for {
		cmd := exec.Command("xset", "q")
		cmd.Env = append(os.Environ(), "DISPLAY="+screen)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()

		if err != nil {
			logrus.Errorf("Failed to read monitor status: %s: %s", fmt.Sprint(err), stderr.String())
			return
		}

		status := re.FindStringSubmatch(string(out))
		if len(status) > 2 {
			dev.State["monitor_status"] = status[2]
			dev.State["on"] = status[2] == "On"
			n.AddOrUpdate(dev)
		}
		<-time.After(time.Second * 1)
	}
}

func changeDpmsState(screen string, state bool) error {
	if state {
		cmd := exec.Command("xset", "dpms", "force", "on")
		cmd.Env = append(os.Environ(), "DISPLAY="+screen)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
	}
	if !state {
		cmd := exec.Command("xset", "dpms", "force", "off")
		cmd.Env = append(os.Environ(), "DISPLAY="+screen)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
	}

	return nil
}
