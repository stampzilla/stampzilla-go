.PHONY:	test cover cover-html build-ui

# stampzilla-telldus needs telldus-core headers
PKGS = $(shell go list ./... | grep -v stampzilla-telldus)

test:
	go test $(PKGS)

cover:
	go test -coverprofile=coverage.txt -coverpkg=./... $(PKGS)

cover-html: cover
	go tool cover -html coverage.txt

build-ui:
	cd nodes/stampzilla-server/public && gulp
	cd nodes/stampzilla-server && go generate
