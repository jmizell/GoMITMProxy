
# GoMITMProxy
[![Build Status](https://travis-ci.com/jmizell/GoMITMProxy.svg?branch=master)](https://travis-ci.com/jmizell/GoMITMProxy)
![GitHub](https://img.shields.io/github/license/jmizell/GoMITMProxy.svg?color=00ff00)
[![GoDoc](https://godoc.org/github.com/jmizell/GoMITMProxy/proxy?status.svg)](https://godoc.org/github.com/jmizell/GoMITMProxy/proxy)

Golang Man in the Middle Proxy

## Docs
https://godoc.org/github.com/jmizell/GoMITMProxy/proxy

## Build
```make build```

## Test
```make test```

## Install

### Local Install

To use, make sure you have go >= 1.12.x installed, and your 
[GOBIN is in your path](https://golang.org/cmd/go/). Then run 

```
go get github.com/jmizell/GoMITMProxy
go install github.com/jmizell/GoMITMProxy/cmd/gomitmproxy
```

### Docker
[![](https://images.microbadger.com/badges/version/jmizell/gomitmproxy.svg)](https://microbadger.com/images/jmizell/gomitmproxy)
[![](https://images.microbadger.com/badges/image/jmizell/gomitmproxy.svg)](https://microbadger.com/images/jmizell/gomitmproxy)

Copies of each realease can be found in [jmizell/gomitmproxy](https://hub.docker.com/r/jmizell/gomitmproxy).

```
docker pull jmizell/gomitmproxy:latest
```

## Usage

```
Usage: gomitmproxy [options]

  -ca_cert_file string
    	path to certificate authority cert file
  -ca_key_file string
    	path to certificate authority key file
  -config string
    	proxy config file path
  -debug
    	enable debug logging
  -dns_port int
    	port to listen for dns requests
  -dns_regex string
    	domains matching this regex pattern will return the proxy address
  -dns_server string
    	use the supplied dns resolver, instead of system defaults
  -generate_ca_only
    	generate a certificate authority, and exit
  -http_ports string
    	ports to listen for http requests (default ",0")
  -https_ports string
    	ports to listen for https requests (default ",0")
  -json
    	output json log format to standard out
  -key_age_hours int
    	certificate authority expire time in hours, used only with generate_ca_only
  -listen_addr string
    	network address bind to (default "127.0.0.1")
  -log_level string
    	set logging to log level (default "INFO")
  -log_responses
    	enable logging upstream server responses
  -request_log_file string
    	file to log http requests
  -version
    	output version

```