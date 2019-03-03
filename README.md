stampzilla-go [![Build Status](https://travis-ci.org/stampzilla/stampzilla-go.svg?branch=master)](https://travis-ci.org/stampzilla/stampzilla-go) [![codecov](https://codecov.io/gh/stampzilla/stampzilla-go/branch/master/graph/badge.svg)](https://codecov.io/gh/stampzilla/stampzilla-go) [![Go Report Card](https://goreportcard.com/badge/github.com/stampzilla/stampzilla-go)](https://goreportcard.com/report/github.com/stampzilla/stampzilla-go)
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
go get -u github.com/stampzilla/stampzilla-go/stampzilla
sudo stampzilla install
```
This creates a stampzilla user. checksout the code in stampzilla user home folder and creates some required folders. 

### Updating

Update the cli with `stampzilla self-update`

Update nodes with
```
sudo stampzilla stop
sudo stampzilla install -u
sudo stampzilla start
```

### Documentation
Is work in progress and can be found here:
* [Docs](docs/README.md)
