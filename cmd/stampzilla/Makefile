LDFLAGS=-ldflags "-X main.buildstamp \"`date -u '+%Y-%m-%d %H:%M:%S %Z'`\" -X main.githash `git rev-parse --short HEAD`"

all:
	go build ${LDFLAGS} $(filter-out $@,$(MAKECMDGOALS))

install:
	go install ${LDFLAGS} $(filter-out $@,$(MAKECMDGOALS))

run:
	go build ${LDFLAGS}
	./stampzilla $(filter-out $@,$(MAKECMDGOALS))

%:
	@:
