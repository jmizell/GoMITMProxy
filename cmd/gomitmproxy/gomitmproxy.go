// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/jmizell/GoMITMProxy/proxy"
)

func main() {
	genCAOnly := flag.Bool("generate_ca_only", false, "generate a certificate authority, and exit")
	KeyAge := flag.Int("key_age_hours", 0, "certificate authority expire time in hours, used only with generate_ca_only")
	CAKeyFile := flag.String("ca_key_file", "", "path to certificate authority key file")
	CACertFile := flag.String("ca_cert_file", "", "path to certificate authority cert file")
	ListenAddr := flag.String("listen_addr", "127.0.0.1", "network address bind to")
	HTTPSPort := flag.Int("https_port", 0, "port to listen for https requests")
	HTTPPort := flag.Int("http_port", 0, "port to listen for http requests")
	DNSPort := flag.Int("dns_port", 0, "port to listen for dns requests")
	DNSServer := flag.String("dns_server", "", "use the supplied dns resolver, instead of system defaults")
	DNSRegex := flag.String("dns_regex", "", "domains matching this regex pattern will return the proxy address")
	config := flag.String("config", "", "proxy config file path")
	requestLogFile := flag.String("request_log_file", "", "file to log http requests")
	logJSON := flag.Bool("json", false, "output json log format to standard out")
	logDebug := flag.Bool("debug", false, "enable debug logging")
	logLevel := flag.String("log_level", "info", "set logging to log level")
	flag.Usage = func() {
		path.Base(os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *genCAOnly {
		c := proxy.Certs{
			KeyAge: time.Duration(*KeyAge) * time.Hour,
		}
		key, cert, err := c.GenerateCAPair()
		if err != nil {
			log.Fatal(err.Error())
		}

		err = proxy.WriteCA(*CACertFile, *CAKeyFile, cert, key)
		if err != nil {
			log.Fatal(err.Error())
		}

		os.Exit(0)
	}

	p := proxy.Proxy{}
	if *config != "" {
		data, err := ioutil.ReadFile(*config)
		if err != nil {
			proxy.Log.WithError(err).WithField("file", *config).Fatal("read config")
		}

		err = json.Unmarshal(data, proxy.Log)
		if err != nil {
			proxy.Log.WithError(err).WithField("file", *config).Fatal("unmarshal log config")
		}

		err = json.Unmarshal(data, p)
		if err != nil {
			proxy.Log.WithError(err).WithField("file", *config).Fatal("unmarshal proxy config")
		}
	}
	p.CAKeyFile = *CAKeyFile
	p.CACertFile = *CACertFile
	p.HTTPSPort = *HTTPSPort
	p.HTTPPort = *HTTPPort
	p.ListenAddr = *ListenAddr
	p.DNSServer = *DNSServer
	p.DNSPort = *DNSPort
	p.DNSRegex = *DNSRegex

	proxy.Log.Level.Parse(*logLevel)
	proxy.Log.RequestLogFile = *requestLogFile

	if *logJSON {
		proxy.Log.Format = proxy.LogJSON
	}

	if *logDebug {
		proxy.Log.Level = proxy.LogDEBUG
	}

	proxy.Log.WithField("log_level", proxy.Log.Level).Debug("")
	proxy.Log.WithField("log_format", proxy.Log.Format).Debug("")
	proxy.Log.WithField("request_log_file", proxy.Log.RequestLogFile).Debug("")
	proxy.Log.WithField("ca_key_file", p.CAKeyFile).Debug("")
	proxy.Log.WithField("ca_cert_file", p.CACertFile).Debug("")
	proxy.Log.WithField("listen_addr", p.ListenAddr).Debug("")
	proxy.Log.WithField("https_port", p.HTTPSPort).Debug("")
	proxy.Log.WithField("http_port", p.HTTPPort).Debug("")
	proxy.Log.WithField("dns_port", p.DNSPort).Debug("")
	proxy.Log.WithField("dns_server", p.DNSServer).Debug("")
	proxy.Log.WithField("dns_regex", p.DNSRegex).Debug("")

	err := p.Run()
	if err != nil {
		proxy.Log.WithError(err).Fatal("proxy server failed")
	}
}
