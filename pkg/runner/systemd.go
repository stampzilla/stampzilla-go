package runner

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/installer"
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

	// stop all running stampzilla processes
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

// Restart restarts currencly running nodes.
func (sd *Systemd) Restart(nodes ...string) error {
	conn := sd.dbus()
	units, err := conn.ListUnitsByPatterns([]string{"active", "activating", "failed"}, []string{"stampzilla-*"})
	if err != nil {
		return err
	}
	ch := make(chan string)

	toRestart := []string{}

	if len(nodes) > 0 {
		// restart requested nodes
		for _, unit := range units {
			for _, node := range nodes {
				if strings.HasPrefix(unit.Name, getUnitName(node)) {
					toRestart = append(toRestart, unit.Name)
				}
			}
		}
	} else {
		// restart all running
		for _, unit := range units {
			toRestart = append(toRestart, unit.Name)
		}
	}

	pids := make(map[string]uint32)
	for _, unit := range toRestart {
		logrus.Info("Restarting ", unit)
		_, err := conn.RestartUnit(unit, "replace", ch)
		if err != nil {
			return err
		}
		if err := jobResult(unit, <-ch); err != nil {
			return err
		}
		pids[unit], err = sd.mainPID(unit)
		if err != nil {
			return err
		}
	}

	if len(toRestart) == 0 {
		return nil
	}

	// A crashing node still gives a successful restart job since the units are
	// Type=simple. Wait a moment and verify that the units are still running.
	logrus.Infof("Waiting %s to verify that restarted nodes keep running", restartCheckDelay)
	time.Sleep(restartCheckDelay)

	failed := []string{}
	for _, unit := range toRestart {
		state, err := sd.unitState(unit)
		if err != nil {
			return err
		}
		if err := checkRunning(unit, pids[unit], state); err != nil {
			logrus.Error(err)
			failed = append(failed, unit)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("not running after restart: %s (see stampzilla log <node>)", strings.Join(failed, ", "))
	}
	return nil
}

// restartCheckDelay is how long Restart waits before verifying that restarted units keep running.
var restartCheckDelay = 3 * time.Second

type unitState struct {
	ActiveState string
	SubState    string
	MainPID     uint32
}

func (sd *Systemd) mainPID(unit string) (uint32, error) {
	p, err := sd.dbus().GetUnitTypeProperty(unit, "Service", "MainPID")
	if err != nil {
		return 0, err
	}
	pid, _ := p.Value.Value().(uint32)
	return pid, nil
}

func (sd *Systemd) unitState(unit string) (unitState, error) {
	props, err := sd.dbus().GetUnitProperties(unit)
	if err != nil {
		return unitState{}, err
	}
	pid, err := sd.mainPID(unit)
	if err != nil {
		return unitState{}, err
	}
	active, _ := props["ActiveState"].(string)
	sub, _ := props["SubState"].(string)
	return unitState{ActiveState: active, SubState: sub, MainPID: pid}, nil
}

// checkRunning returns an error if the unit is not running, or has been
// restarted by systemd (new main pid) since startedPID was read.
func checkRunning(unit string, startedPID uint32, s unitState) error {
	if s.ActiveState != "active" || s.SubState != "running" {
		return fmt.Errorf("%s is %s (%s) after restart", unit, s.ActiveState, s.SubState)
	}
	if startedPID != 0 && s.MainPID != startedPID {
		return fmt.Errorf("%s crashed and was restarted by systemd (pid %d -> %d)", unit, startedPID, s.MainPID)
	}
	return nil
}

// jobResult converts the result of a systemd job to an error. The result is
// "done" on success, otherwise ex "failed", "timeout", "canceled" or "dependency".
func jobResult(unit, result string) error {
	if result != "done" {
		return fmt.Errorf("%s: systemd job %s", unit, result)
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

// GenerateUnit creates the node's config dir, owned by stampzilla, and writes
// and enables its systemd unit.
func (sd *Systemd) GenerateUnit(name string) error {
	installer.CreateDirAsUser(nodeConfigDir(name), "stampzilla")
	return sd.WriteUnit(name)
}

// nodeConfigDir returns the working directory a node's unit runs in.
func nodeConfigDir(name string) string {
	name = strings.Split(name, "@")[0]
	return filepath.Join("/", "etc", "stampzilla", "nodes", GetProcessName(name))
}

// WriteUnit writes and enables the node's systemd unit without touching its
// config dir. Use it when the config dir already is in place and might be
// owned by someone else than root.
func (sd *Systemd) WriteUnit(name string) error {
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

	tmp := strings.Split(name, "@")
	name = tmp[0]

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

// Running returns the node names (unit name without .service) of all active stampzilla units.
func (sd *Systemd) Running() ([]string, error) {
	units, err := sd.dbus().ListUnitsByPatterns([]string{"active", "activating"}, []string{"stampzilla-*.service"})
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(units))
	for _, u := range units {
		ret = append(ret, strings.TrimSuffix(u.Name, ".service"))
	}
	return ret, nil
}
