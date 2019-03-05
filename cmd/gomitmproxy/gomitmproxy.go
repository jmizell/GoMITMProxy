// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jmizell/GoMITMProxy/proxy"
	"github.com/jmizell/GoMITMProxy/proxy/log"
)

func main() {
	genCAOnly := flag.Bool("generate_ca_only", false, "generate a certificate authority, and exit")
	KeyAge := flag.Int("key_age_hours", 0, "certificate authority expire time in hours, used only with generate_ca_only")
	CAKeyFile := flag.String("ca_key_file", "", "path to certificate authority key file")
	CACertFile := flag.String("ca_cert_file", "", "path to certificate authority cert file")
	ListenAddr := flag.String("listen_addr", "127.0.0.1", "network address bind to")
	HTTPSPorts := flag.String("https_ports", "0", "ports to listen for https requests")
	HTTPPorts := flag.String("http_ports", "0", "ports to listen for http requests")
	DNSPort := flag.Int("dns_port", 0, "port to listen for dns requests")
	DNSServer := flag.String("dns_server", "", "use the supplied dns resolver, instead of system defaults")
	DNSRegex := flag.String("dns_regex", "", "domains matching this regex pattern will return the proxy address")
	config := flag.String("config", "", "proxy config file path")
	requestLogFile := flag.String("request_log_file", "", "file to log http requests")
	logJSON := flag.Bool("json", false, "output json log format to standard out")
	logDebug := flag.Bool("debug", false, "enable debug logging")
	logLevel := flag.String("log_level", log.INFO.String(), "set logging to log level")
	flag.Usage = func() {
		path.Base(os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	// Generate certificate authority only, and exit
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

	// Parse config file to values
	p := proxy.Proxy{}
	logConfig := log.NewDefaultConfig()
	if *config != "" {

		data, err := ioutil.ReadFile(*config)
		if err != nil {
			log.WithError(err).WithField("file", *config).Fatal("read config")
		}

		err = json.Unmarshal(data, logConfig)
		if err != nil {
			log.WithError(err).WithField("file", *config).Fatal("unmarshal log config")
		}

		err = json.Unmarshal(data, p)
		if err != nil {
			log.WithError(err).WithField("file", *config).Fatal("unmarshal proxy config")
		}
	}

	// Setting proxy command line arguments, values supersede config values
	var err error
	p.HTTPSPorts, err = listToInts(*HTTPSPorts, ",")
	if err != nil {
		log.WithError(err).WithField("https_ports", *HTTPSPorts).Fatal("flag parse failure")
	}
	p.HTTPPorts, err = listToInts(*HTTPPorts, ",")
	if err != nil {
		log.WithError(err).WithField("http_ports", *HTTPPorts).Fatal("flag parse failure")
	}
	p.CAKeyFile = *CAKeyFile
	p.CACertFile = *CACertFile
	p.ListenAddr = *ListenAddr
	p.DNSServer = *DNSServer
	p.DNSPort = *DNSPort
	p.DNSRegex = *DNSRegex

	// Setting logger command line arguments, values supersede config values
	logConfig.Level.Parse(*logLevel)
	if *logJSON {
		logConfig.Format = log.JSON
	}
	if *requestLogFile != "" {
		logConfig.RequestLogFile = *requestLogFile
	}
	if *logDebug {
		logConfig.Level = log.DEBUG
	}
	logger, requestWriter := logConfig.GetLogger()
	log.DefaultLogger = logger
	if requestWriter != nil {
		defer requestWriter.Close()
	}

	// Output config values for debugging
	log.WithField("log_level", logConfig.Level).Debug("")
	log.WithField("log_format", logConfig.Format).Debug("")
	log.WithField("request_log_file", logConfig.RequestLogFile).Debug("")
	log.WithField("ca_key_file", p.CAKeyFile).Debug("")
	log.WithField("ca_cert_file", p.CACertFile).Debug("")
	log.WithField("listen_addr", p.ListenAddr).Debug("")
	log.WithField("https_ports", p.HTTPSPorts).Debug("")
	log.WithField("http_ports", p.HTTPPorts).Debug("")
	log.WithField("dns_port", p.DNSPort).Debug("")
	log.WithField("dns_server", p.DNSServer).Debug("")
	log.WithField("dns_regex", p.DNSRegex).Debug("")

	// Start the proxy
	if err = p.Run(); err != nil {
		log.WithError(err).Fatal("proxy server failed")
	}
}

func listToInts(s, sep string) (ints []int, err error) {

	for _, v := range strings.Split(s, sep) {

		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}

		ints = append(ints, i)
	}

	return ints, nil
}
