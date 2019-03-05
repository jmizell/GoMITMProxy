// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/benburkert/dns"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

const DefaultDNSServer = "8.8.8.8"

type DNSServer struct {
	server    *dns.Server
	dnsRegex  *regexp.Regexp
	record    *dns.A
	dnsClient *dns.Client

	ListenAddr string `json:"listen_addr"`
	DNSPort    int    `json:"dns_port"`
	DNSServer  string `json:"dns_server"`
	DNSRegex   string `json:"dns_regex"`
}

func (d *DNSServer) ListenAndServe() (err error) {

	d.record = &dns.A{A: net.ParseIP(d.ListenAddr).To4()}
	d.dnsRegex, err = regexp.Compile(d.DNSRegex)

	if d.DNSServer == "" {
		d.DNSServer = DefaultDNSServer
	}

	d.dnsClient = &dns.Client{
		Transport: &dns.Transport{
			Proxy: dns.NameServers{
				&net.TCPAddr{IP: net.ParseIP(d.DNSServer), Port: 53},
				&net.UDPAddr{IP: net.ParseIP(d.DNSServer), Port: 53},
			}.RoundRobin(),
		},
		Resolver: new(dns.Cache),
	}

	d.server = &dns.Server{
		Addr:    fmt.Sprintf("%s:%d", d.ListenAddr, d.DNSPort),
		Handler: d,
	}
	log.WithField("addr", fmt.Sprintf("%s:%d", d.ListenAddr, d.DNSPort)).
		Info("dns server started")

	return d.server.ListenAndServe(context.Background())
}

func (d *DNSServer) ServeDNS(ctx context.Context, w dns.MessageWriter, r *dns.Query) {

	var found bool
	var matchRegex bool
	for _, q := range r.Questions {

		if !d.dnsRegex.MatchString(q.Name) {
			continue
		}

		matchRegex = true

		if q.Type == dns.TypeA {
			w.Answer(q.Name, time.Minute, d.record)
			found = true
		}
	}

	if !found && !matchRegex {
		res, err := d.dnsClient.Do(context.Background(), r)
		if err != nil {
			log.WithError(err).Error("dns client forwarding failed")
		}

		for _, r := range res.Answers {
			w.Answer(r.Name, r.TTL, r.Record)
			found = true
		}
	}

	if !found {
		w.Status(dns.NXDomain)
	}
}
