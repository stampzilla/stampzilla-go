.PHONY:	test cover cover-html cover-test

test:
	go test `go list ./... | grep -v /vendor/`

# todo: use this when golang issue 23910 is resolved
# go test -v -coverpkg=./... -coverprofile=all `go list ./... | grep -v /vendor/ `
cover:
	@echo Running coverage
	go get github.com/wadey/gocovmerge
	$(eval PKGS := $(shell go list ./... | grep -v /vendor/ ))
	$(eval PKGS_DELIM := $(shell echo $(PKGS) | sed -e 's/ /,/g'))
	go list -f '{{if or (len .TestGoFiles) (len .XTestGoFiles)}}go test -test.v -test.timeout=120s -covermode=atomic -coverprofile={{.Name}}_{{len .Imports}}_{{len .Deps}}.coverprofile -coverpkg $(PKGS_DELIM) {{.ImportPath}}{{end}}' $(PKGS) | xargs -I {} bash -c {}
	gocovmerge `ls *.coverprofile` > coverage.txt
	rm *.coverprofile
cover-normal:
	bash coverage

cover-html: cover
	go tool cover -html coverage.txt

build-ui: 
	cd nodes/stampzilla-server/public && gulp
	cd nodes/stampzilla-server && go generate
