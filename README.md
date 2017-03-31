stampzilla-go [![Build Status](https://travis-ci.org/stampzilla/stampzilla-go.svg?branch=master)](https://travis-ci.org/stampzilla/stampzilla-go) [![codecov](https://codecov.io/gh/stampzilla/stampzilla-go/branch/master/graph/badge.svg)](https://codecov.io/gh/stampzilla/stampzilla-go)
=============

New version of this awesome homeautomation software written in Go and javascript with Polymer and webcomponents. 

Installation from precompiled binaries
```bash
curl -s https://api.github.com/repos/stampzilla/stampzilla-go/releases | \
grep "browser_download_url.*stampzilla-amd64" | \
head -n 1 | cut -d '"' -f 4 | xargs curl -L -o stampzilla && \
chmod +x stampzilla
```
Move stampzilla to /usr/local/bin or ~/bin depending on your setup. 
```mv stampzilla /usr/local/bin```

Installation from source
```bash
go get -u github.com/stampzilla/stampzilla-go/stampzilla
sudo stampzilla install
```
This creates a stampzilla user. checksout the code in stampzilla user home folder and creates some required folders. 

### Documentation
Is work in progress and can be found here:
* [Docs](docs/README.md)
