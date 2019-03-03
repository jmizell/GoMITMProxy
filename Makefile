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
	go test -cover -v -timeout=15m ./...

clean:
	rm -f *.crt *.key gomitmproxy

.EXPORT_ALL_VARIABLES:
