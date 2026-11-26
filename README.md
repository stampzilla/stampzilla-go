stampzilla-go [![CI](https://github.com/stampzilla/stampzilla-go/actions/workflows/go.yml/badge.svg)](https://github.com/stampzilla/stampzilla-go/actions/workflows/go.yml) [![codecov](https://codecov.io/gh/stampzilla/stampzilla-go/branch/master/graph/badge.svg)](https://codecov.io/gh/stampzilla/stampzilla-go) [![Go Report Card](https://goreportcard.com/badge/github.com/stampzilla/stampzilla-go)](https://goreportcard.com/report/github.com/stampzilla/stampzilla-go)
=============

Awesome homeautomation software written in Go and React

### Installing

Installation from precompiled binaries
```bash
sudo curl -fsSL -o /usr/local/bin/stampzilla https://github.com/stampzilla/stampzilla-go/releases/latest/download/stampzilla-linux-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/armv.*/arm/') && sudo chmod +x /usr/local/bin/stampzilla && sudo stampzilla install server
```
This downloads the cli for your architecture (amd64, arm64 or arm) and installs the server. Add nodes later with for example `sudo stampzilla install deconz`.

Installation from source
```bash
go install github.com/stampzilla/stampzilla-go/v2/cmd/stampzilla@latest
sudo stampzilla install
```
This creates a stampzilla user. checksout the code in stampzilla user home folder and creates some required folders. 

### Updating

Always update the cli first, then the nodes:
```
sudo stampzilla self-update
sudo stampzilla update
```
`stampzilla update` updates all installed nodes and restarts the ones that are running. Update specific nodes with for example `sudo stampzilla update server tibber`.

Add `@version` to install, upgrade or downgrade a node to a specific release, for example `sudo stampzilla update tibber@v2.9.1`. A later `stampzilla update` without a version moves it back to the latest release.

Updating the cli first matters because newer releases can change how the binaries are published. Since the switch to gzip compressed node binaries, a cli older than that release can't find the nodes and fails with `Failed to install ...`.

### Documentation
Is work in progress and can be found here:
* [Docs](docs/README.md)
* [Screenshots](docs/screenshots.md)
