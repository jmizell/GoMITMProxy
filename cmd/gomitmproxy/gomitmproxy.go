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

	var p proxy.Proxy

	if *config == "" {
		p = proxy.Proxy{
			CAKeyFile:  *CAKeyFile,
			CACertFile: *CACertFile,
			HTTPSPort:  *HTTPSPort,
			HTTPPort:   *HTTPPort,
			ListenAddr: *ListenAddr,
			DNSServer:  *DNSServer,
			DNSPort:    *DNSPort,
			DNSRegex:   *DNSRegex,
		}
	} else {
		data, err := ioutil.ReadFile(*config)
		if err != nil {
			log.Fatal(err.Error())
		}

		err = json.Unmarshal(data, p)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	err := p.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
