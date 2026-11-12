package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stampzilla/stampzilla-go/v2/pkg/backup"
	"github.com/stampzilla/stampzilla-go/v2/pkg/build"
	"github.com/stampzilla/stampzilla-go/v2/pkg/installer"
	"github.com/stampzilla/stampzilla-go/v2/pkg/installer/binary"
	"github.com/stampzilla/stampzilla-go/v2/pkg/runner"
	"github.com/urfave/cli/v2"
)

var defaultBackupPaths = backup.Paths{
	ConfigDir: "/etc/stampzilla",
	BinDir:    "/home/stampzilla/go/bin",
}

func (t *cliHandler) Backup(c *cli.Context) error {
	requireRoot()

	out := c.String("output")
	if out == "" {
		hostname, _ := os.Hostname()
		out = fmt.Sprintf("stampzilla-backup-%s-%s.tar.gz", hostname, time.Now().Format("20060102-150405"))
	}
	if inside, err := isInside(out, defaultBackupPaths.ConfigDir); err != nil {
		return err
	} else if inside {
		return fmt.Errorf("output %s is inside %s which is being backed up, choose another location with -o", out, defaultBackupPaths.ConfigDir)
	}

	if c.Bool("stop") {
		if !c.Bool("systemd") {
			return fmt.Errorf("--stop is only supported with systemd")
		}
		r := &runner.Systemd{}
		defer r.Close()
		running, err := r.Running()
		if err != nil {
			return err
		}
		if len(running) > 0 { // Stop/Start without arguments acts on all units
			// Registered before stopping so nodes stopped before a failure are started again.
			defer func() {
				if err := r.Start(running...); err != nil {
					logrus.Errorf("failed to start nodes again: %s", err)
				}
			}()
			logrus.Info("Stopping nodes to get a consistent backup")
			if err := r.Stop(running...); err != nil {
				return err
			}
		}
	}

	tmp := out + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // contains private keys
	if err != nil {
		return err
	}
	m, err := backup.Create(f, defaultBackupPaths, backup.Options{
		IncludeBinaries: c.Bool("binaries"),
		Version:         build.Version,
		Warn:            func(msg string) { logrus.Warn(msg) },
	})
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, out); err != nil {
		return err
	}

	logrus.Infof("Backup of %d nodes written to %s", len(m.Nodes), out)
	return nil
}

func (t *cliHandler) Restore(c *cli.Context) error {
	requireRoot()

	file := c.Args().First()
	if file == "" {
		return fmt.Errorf("usage: stampzilla restore <backup.tar.gz>")
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	m, err := backup.ReadManifest(f)
	if err != nil {
		return err
	}
	logrus.Infof("Restoring backup from %s created %s (version %s, %s/%s) with nodes: %v",
		m.Hostname, m.CreatedAt.Local().Format(time.RFC3339), m.Version, m.OS, m.Arch, m.Nodes)

	skipBinaries := !m.Binaries || c.Bool("skip-install")
	if m.Binaries && (m.OS != runtime.GOOS || m.Arch != runtime.GOARCH) {
		logrus.Warnf("Bundled binaries are for %s/%s, will download %s/%s binaries instead", m.OS, m.Arch, runtime.GOOS, runtime.GOARCH)
		skipBinaries = true
	}

	// Creates the stampzilla user and the base directories.
	if err := installer.Prepare(); err != nil {
		return err
	}

	var sd *runner.Systemd
	if c.Bool("systemd") {
		sd = &runner.Systemd{}
		defer sd.Close()
		logrus.Info("Stopping running nodes")
		if err := sd.Stop(); err != nil {
			return err
		}
	} else {
		r := &runner.Bare{}
		if err := r.Stop(); err != nil {
			return err
		}
	}

	// Keep the current config so nothing is lost if the restore was a mistake.
	cfgDir := defaultBackupPaths.ConfigDir
	old := fmt.Sprintf("%s.pre-restore-%s", cfgDir, time.Now().Format("20060102-150405"))
	if err := os.Rename(cfgDir, old); err != nil {
		return fmt.Errorf("move existing config aside: %w", err)
	}
	logrus.Infof("Existing config moved to %s", old)

	chown, err := chownTo("stampzilla")
	if err != nil {
		return err
	}
	installer.CreateDirAsUser(cfgDir, "stampzilla")

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := backup.Extract(f, defaultBackupPaths, backup.RestoreOptions{
		SkipBinaries: skipBinaries,
		Chown:        chown,
	}); err != nil {
		if rbErr := rollback(cfgDir, old); rbErr != nil {
			return fmt.Errorf("restore failed: %w (rollback also failed: %s, previous config is in %s)", err, rbErr, old)
		}
		return fmt.Errorf("restore failed, previous config was put back: %w", err)
	}
	installer.CreateDirAsUser(filepath.Join(cfgDir, "nodes"), "stampzilla")
	logrus.Infof("Config restored to %s", cfgDir)

	if skipBinaries && !c.Bool("skip-install") && len(m.Nodes) > 0 {
		logrus.Info("Downloading binaries")
		// Install skips binaries that already exist.
		if err := binary.NewInstaller().Install(m.Nodes...); err != nil {
			return fmt.Errorf("install binaries: %w", err)
		}
	}

	if sd != nil {
		for _, node := range m.Nodes {
			if err := sd.GenerateUnit(node); err != nil {
				return err
			}
		}
	}

	if c.Bool("start") {
		if sd != nil {
			if len(m.Nodes) == 0 {
				return nil
			}
			return sd.Start(m.Nodes...)
		}
		r := &runner.Bare{}
		return r.Start()
	}

	logrus.Info("Restore done. Start the nodes with: stampzilla start")
	return nil
}

func rollback(cfgDir, old string) error {
	if err := os.RemoveAll(cfgDir); err != nil {
		return err
	}
	return os.Rename(old, cfgDir)
}

// isInside reports whether p is dir or is located below dir.
func isInside(p, dir string) (bool, error) {
	absP, err := filepath.Abs(p)
	if err != nil {
		return false, err
	}
	if resolved, err := filepath.EvalSymlinks(filepath.Dir(absP)); err == nil {
		absP = filepath.Join(resolved, filepath.Base(absP))
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}
	if resolved, err := filepath.EvalSymlinks(absDir); err == nil {
		absDir = resolved
	}
	rel, err := filepath.Rel(absDir, absP)
	if err != nil {
		return false, err
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))), nil
}

func chownTo(username string) (func(string) error, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return nil, err
	}
	return func(p string) error {
		return os.Lchown(p, uid, gid)
	}, nil
}
