stampzilla-go [![CI](https://github.com/stampzilla/stampzilla-go/actions/workflows/go.yml/badge.svg)](https://github.com/stampzilla/stampzilla-go/actions/workflows/go.yml) [![codecov](https://codecov.io/gh/stampzilla/stampzilla-go/branch/master/graph/badge.svg)](https://codecov.io/gh/stampzilla/stampzilla-go) [![Go Report Card](https://goreportcard.com/badge/github.com/stampzilla/stampzilla-go)](https://goreportcard.com/report/github.com/stampzilla/stampzilla-go)
=============

Awesome homeautomation software written in Go and React

### Installing

Installation from precompiled binaries
```bash
curl -s https://api.github.com/repos/stampzilla/stampzilla-go/releases/latest | grep "browser_download_url.*stampzilla-linux-amd64" | cut -d : -f 2,3 | tr -d \" | xargs curl -L -s -o stampzilla && chmod +x stampzilla
sudo mv stampzilla /usr/local/bin #or ~/bin if you use that
sudo stampzilla install server deconz #or whatever nodes you want to use.
```

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
sudo stampzilla install -u
sudo stampzilla restart
```

Updating the cli first matters because newer releases can change how the binaries are published. Since the switch to gzip compressed node binaries, a cli older than that release can't find the nodes and fails with `Failed to install ...`.

### Documentation
Is work in progress and can be found here:
* [Docs](docs/README.md)
* [Screenshots](docs/screenshots.md)
