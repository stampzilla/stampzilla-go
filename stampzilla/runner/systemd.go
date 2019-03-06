package runner

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/coreos/go-systemd/dbus"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/stampzilla/installer"
)

type Systemd struct {
	dbusConn *dbus.Conn
}

func (sd *Systemd) Close() {
	if sd.dbusConn != nil {
		sd.dbusConn.Close()
	}
}

func (sd *Systemd) dbus() *dbus.Conn {
	if sd.dbusConn != nil {
		return sd.dbusConn
	}

	var err error
	sd.dbusConn, err = dbus.NewSystemdConnection()
	if err != nil {
		logrus.Fatal("systemd init error: ", err)
	}

	return sd.dbusConn
}

func (sd *Systemd) Start(nodes ...string) error {

	conn := sd.dbus()

	if len(nodes) > 0 {
		for _, name := range nodes {
			name = getUnitName(name)
			logrus.Info("Starting ", name)
			ch := make(chan string)
			_, err := conn.StartUnit(name, "replace", ch)
			if err != nil {
				return err
			}
			<-ch
		}
		return nil
	}

	units, err := conn.ListUnitsByPatterns([]string{"inactive", "failed", "deactivating"}, []string{"stampzilla-*"})
	if err != nil {
		return err
	}
	for _, u := range units {
		logrus.Info("Starting ", u.Name)
		ch := make(chan string)
		_, err := conn.StartUnit(u.Name, "replace", ch)
		if err != nil {
			return err
		}
		<-ch
	}
	return nil
}

func (sd *Systemd) Stop(nodes ...string) error {
	conn := sd.dbus()
	if len(nodes) > 0 {
		for _, node := range nodes {
			name := getUnitName(node)
			logrus.Info("Stopping ", name)
			ch := make(chan string)
			_, err := conn.StopUnit(name, "replace", ch)
			if err != nil {
				return err
			}
			<-ch
		}
		return nil
	}

	units, err := conn.ListUnitsByPatterns([]string{"active", "activating", "failed"}, []string{"stampzilla-*"})
	if err != nil {
		return err
	}

	//stop all running stampzilla processes
	for _, p := range units {
		logrus.Info("Stopping ", p.Name)
		ch := make(chan string)
		_, err := conn.StopUnit(p.Name, "replace", ch)
		if err != nil {
			return err
		}
		<-ch
	}
	return nil
}

// Restart restarts currencly running nodes
func (sd *Systemd) Restart(nodes ...string) error {
	conn := sd.dbus()
	units, err := conn.ListUnitsByPatterns([]string{"active"}, []string{"stampzilla-*"})
	if err != nil {
		return err
	}
	ch := make(chan string)
	for _, p := range units {
		logrus.Info("Stopping ", p.Name)
		_, err := conn.StopUnit(p.Name, "replace", ch)
		if err != nil {
			return err
		}
		<-ch

		logrus.Info("Starting ", p.Name)
		_, err = conn.StartUnit(p.Name, "replace", ch)
		if err != nil {
			return err
		}
		<-ch
	}
	return nil
}

func (sd *Systemd) Status() error {
	conn := sd.dbus()
	units, err := conn.ListUnitsByPatterns([]string{"active", "activating", "failed"}, []string{"stampzilla-*"})
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "Name", "Active", "Pid", "CPU%", "Memory%", "Memory")
	for _, v := range units {
		prop, err := conn.GetServiceProperty(v.Name, "MainPID")
		if err != nil {
			logrus.Error(err)
			continue
		}
		pid := prop.Value.Value().(uint32)

		cpu, memPercent, mem := sd.getStats(pid)
		fmt.Fprintf(w, "%s\t%s\t%d\t%.2f%%\t%.2f%%\t%s\n", v.Name, v.ActiveState, pid, cpu, memPercent, humanize.Bytes(mem))
	}
	return w.Flush()
}

func (sd *Systemd) getStats(pid uint32) (cpu float64, memPercent float32, mem uint64) {
	if pid == 0 {
		return
	}

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		logrus.Error(err)
	}
	cpu, err = p.CPUPercent()
	if err != nil {
		logrus.Error(err)
	}
	memPercent, err = p.MemoryPercent()
	if err != nil {
		logrus.Error(err)
	}
	m, err := p.MemoryInfo()
	if err != nil {
		logrus.Error(err)
	}
	mem = m.RSS
	return
}

func (sd *Systemd) GenerateUnit(name string) error {
	u := `[Unit]
Description={{.Name}}

[Service]
Type=simple

User=stampzilla
Group=stampzilla

Restart=always
RestartSec=5s

ExecStart=/home/stampzilla/go/bin/{{.Name}}
WorkingDirectory=/etc/stampzilla/nodes/{{.Name}}

[Install]
WantedBy=multi-user.target
`

	type td struct {
		Name string
	}
	tmplData := td{
		Name: GetProcessName(name),
	}
	t, err := template.New("t1").Parse(u)
	if err != nil {
		return err
	}

	unitName := getUnitName(name)
	installer.CreateDirAsUser(filepath.Join("/", "etc", "stampzilla", "nodes", strings.TrimSuffix(unitName, filepath.Ext(unitName))), "stampzilla")
	p := filepath.Join("/", "etc", "systemd", "system", unitName)
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("create file: %s", err)
	}
	defer f.Close()
	logrus.Infof("Generating %s", p)
	err = t.Execute(f, tmplData)
	if err != nil {
		return err
	}

	conn := sd.dbus()

	err = conn.Reload()
	if err != nil {
		return err
	}

	_, _, err = conn.EnableUnitFiles([]string{unitName}, false, false)
	return err
}

func (sd *Systemd) Disable(names ...string) error {
	conn := sd.dbus()
	ns := []string{}

	for _, name := range names {
		un := getUnitName(name)
		logrus.Infof("Disabling %s", un)
		ns = append(ns, un)
	}
	_, err := conn.DisableUnitFiles(ns, false)
	return err
}

func getUnitName(s string) string {
	return GetProcessName(s) + ".service"
}
