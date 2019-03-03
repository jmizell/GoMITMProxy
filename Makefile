SHELL  := /bin/bash
GOPATH :=

all: test gomitmproxy

build: gomitmproxy

gomitmproxy:
	CGO_ENABLED=0 GOOS=linux go build \
		-a \
		-installsuffix cgo \
		-o gomitmproxy \
		github.com/jmizell/GoMITMProxy/cmd/gomitmproxy

test:
	go test -cover -coverprofile=cover.out -v -timeout=15m ./... \
	&& go tool cover -html=cover.out -o cover.html

clean:
	rm -f *.crt *.key cover.out cover.html gomitmproxy

.EXPORT_ALL_VARIABLES:
