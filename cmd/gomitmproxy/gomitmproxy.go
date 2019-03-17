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

	// Get a default proxy server, and log config
	p := proxy.NewProxyWithDefaults()
	logConfig := log.NewDefaultConfig()

	// parse some flags
	genCAOnly := flag.Bool("generate_ca_only", false, "generate a certificate authority, and exit")
	KeyAge := flag.Int("key_age_hours", 0, "certificate authority expire time in hours, used only with generate_ca_only")
	config := flag.String("config", "", "proxy config file path")
	CAKeyFile := flag.String("ca_key_file", p.CAKeyFile, "path to certificate authority key file")
	CACertFile := flag.String("ca_cert_file", p.CACertFile, "path to certificate authority cert file")
	ListenAddr := flag.String("listen_addr", p.ListenAddr, "network address bind to")
	HTTPSPorts := flag.String("https_ports", intsToString(p.HTTPSPorts), "ports to listen for https requests")
	HTTPPorts := flag.String("http_ports", intsToString(p.HTTPPorts), "ports to listen for http requests")
	DNSPort := flag.Int("dns_port", p.DNSPort, "port to listen for dns requests")
	DNSResolverOverride := flag.String("dns_resolver_override", p.DNSResolverOverride, "use the supplied dns resolver, instead of system defaults")
	ForwardDNSServer := flag.String("forward_dns_server", p.ForwardDNSServer, "use the supplied dns resolver, instead of system defaults")
	DNSRegex := flag.String("dns_regex", p.DNSRegex, "domains matching this regex pattern will return the proxy address")
	logResponses := flag.Bool("log_responses", p.LogResponses, "enable logging upstream server responses")
	requestLogFile := flag.String("request_log_file", logConfig.RequestLogFile, "file to log dns, and http requests")
	webHookURL := flag.String("webhook_url", logConfig.WebHookURL, "url to post request, and dns logs")
	logJSON := flag.Bool("json", false, "output json log format to standard out")
	logDebug := flag.Bool("debug", false, "enable debug logging")
	logLevel := flag.String("log_level", log.DefaultLogger.Level.String(), "set logging to log level")
	version := flag.Bool("version", false, "output version")
	flag.Usage = func() {
		path.Base(os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *version {
		fmt.Println("Version ", proxy.Version)
		os.Exit(0)
	}

	// Generate certificate authority only, and exit
	if *genCAOnly {
		GenerateCA(*KeyAge, *CACertFile, *CAKeyFile)
	}

	// Parse config file to values
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
	flag.Visit(func(i *flag.Flag) {
		switch i.Name {
		case "ca_key_file":
			p.CAKeyFile = *CAKeyFile
		case "ca_cert_file":
			p.CACertFile = *CACertFile
		case "listen_addr":
			p.ListenAddr = *ListenAddr
		case "https_ports":
			p.HTTPSPorts, err = listToInts(*HTTPSPorts, ",")
			if err != nil {
				log.WithError(err).WithField("https_ports", *HTTPSPorts).Fatal("flag parse failure")
			}
		case "http_ports":
			p.HTTPPorts, err = listToInts(*HTTPPorts, ",")
			if err != nil {
				log.WithError(err).WithField("http_ports", *HTTPPorts).Fatal("flag parse failure")
			}
		case "dns_port":
			p.DNSPort = *DNSPort
		case "forward_dns_server":
			p.ForwardDNSServer = *ForwardDNSServer
		case "dns_resolver_override":
			p.DNSResolverOverride = *DNSResolverOverride
		case "dns_regex":
			p.DNSRegex = *DNSRegex
		case "log_responses":
			p.LogResponses = *logResponses
		case "webhook_url":
			logConfig.WebHookURL = *webHookURL
		case "log_level":
			logConfig.Level.Parse(*logLevel)
		}
	})

	// Setting logger command line arguments, values supersede config values
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
		defer func() { _ = requestWriter.Close() }()
	}

	// Output config values for debugging
	log.WithField("log_level", logConfig.Level).Debug("")
	log.WithField("log_format", logConfig.Format).Debug("")
	log.WithField("request_log_file", logConfig.RequestLogFile).Debug("")
	log.WithField("webhook_url", logConfig.WebHookURL).Debug("")
	log.WithField("ca_key_file", p.CAKeyFile).Debug("")
	log.WithField("ca_cert_file", p.CACertFile).Debug("")
	log.WithField("listen_addr", p.ListenAddr).Debug("")
	log.WithField("https_ports", p.HTTPSPorts).Debug("")
	log.WithField("http_ports", p.HTTPPorts).Debug("")
	log.WithField("dns_port", p.DNSPort).Debug("")
	log.WithField("forward_dns_server", p.ForwardDNSServer).Debug("")
	log.WithField("dns_resolver_override", p.DNSResolverOverride).Debug("")
	log.WithField("dns_regex", p.DNSRegex).Debug("")
	log.WithField("log_responses", p.LogResponses).Debug("")

	// Start the proxy
	if err = p.Run(); err != nil {
		log.WithError(err).Fatal("proxy server failed")
	}
}

func GenerateCA(KeyAge int, CACertFile, CAKeyFile string) {

	c := proxy.Certs{
		KeyAge: time.Duration(KeyAge) * time.Hour,
	}
	_, _, err := c.GenerateCAPair()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = c.WriteCA(CACertFile, CAKeyFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	os.Exit(0)
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

func intsToString(ints []int) (intString string) {

	strInts := make([]string, len(ints))
	for _, i := range ints {
		strInts = append(strInts, fmt.Sprintf("%d", i))
	}

	return strings.Join(strInts, ",")
}
