SHELL  := /bin/bash
GOPATH :=
BUILDID := $(shell if [ "${TRAVIS_JOB_NUMBER}" == "" ]; then echo 0; else echo ${TRAVIS_JOB_NUMBER}; fi)

all: test gomitmproxy

build: gomitmproxy

gomitmproxy:
	CGO_ENABLED=0 GOOS=linux go build \
		-a \
		-installsuffix cgo \
		-o gomitmproxy \
		cmd/gomitmproxy/gomitmproxy.go

docker: docker_app docker_selenium

docker_app:
	docker build --no-cache --force-rm -t jmizell/gomitmproxy:test-app-$(BUILDID) -f Dockerfile .

docker_selenium:
	docker build --no-cache --force-rm -t jmizell/gomitmproxy:test-selenium-$(BUILDID) -f selenium/DockerfileChrome .

test:
	go test -cover -coverprofile=cover.out -v -timeout=15m ./... \
	&& go tool cover -html=cover.out -o cover.html

test_integration:
	docker build --no-cache --force-rm -t jmizell/gomitmproxy:selenium_integration_test  -f selenium/DockerfileChrome . \
	&& go test -v -timeout=15m -tags integration_test github.com/jmizell/GoMITMProxy/selenium

clean:
	rm -f *.crt *.key cover.out cover.html gomitmproxy

.EXPORT_ALL_VARIABLES:
