LDFLAGS=-ldflags "-X main.buildstamp \"`date -u '+%Y-%m-%d %H:%M:%S'`\" -X main.githash `git rev-parse --short HEAD`"

all:
	go build ${LDFLAGS}
