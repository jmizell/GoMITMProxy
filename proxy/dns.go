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

// DefaultDNSServer is the default forward dns server DNSServer will use if ForwardDNSServer is left unset
const DefaultDNSServer = "8.8.8.8"

// DNSServer is a forwarding dns server, that can redirect arbitrary A record requests back to the MITMProxy listening
// address. DNSServer only returns requests for valid dns entries, any request that cannot be answered by the forward
// dns server is returned nxdomain.
type DNSServer struct {
	server    *dns.Server
	dnsRegex  *regexp.Regexp
	record    *dns.A
	dnsClient *dns.Client

	ListenAddr       string `json:"listen_addr"`        // UDP address to listen for dns requests
	Port             int    `json:"port"`               // UDP Port to listen for dns requests
	ForwardDNSServer string `json:"forward_dns_server"` // Forward DNS server to query for each request
	DNSRegex         string `json:"dns_regex"`          // A record requests that match this pattern will return the proxy ip
}

// ListenAndServe starts a forwarding DNS server on both the TCP and UDP network address ListenAddr.
func (d *DNSServer) ListenAndServe() (err error) {

	d.record = &dns.A{A: net.ParseIP(d.ListenAddr).To4()}
	d.dnsRegex, err = regexp.Compile(d.DNSRegex)

	if d.ForwardDNSServer == "" {
		d.ForwardDNSServer = DefaultDNSServer
	}

	d.dnsClient = &dns.Client{
		Transport: &dns.Transport{
			Proxy: dns.NameServers{
				&net.TCPAddr{IP: net.ParseIP(d.ForwardDNSServer), Port: 53},
				&net.UDPAddr{IP: net.ParseIP(d.ForwardDNSServer), Port: 53},
			}.RoundRobin(),
		},
		Resolver: new(dns.Cache),
	}

	d.server = &dns.Server{
		Addr:    fmt.Sprintf("%s:%d", d.ListenAddr, d.Port),
		Handler: d,
	}
	log.WithField("addr", fmt.Sprintf("%s:%d", d.ListenAddr, d.Port)).
		Info("dns server started")

	return d.server.ListenAndServe(context.Background())
}

// ServeDNS handles incoming dns requests, forwarding to an upstream server, and overwriting A record answers
// that match the pattern in DNSRegex.
func (d *DNSServer) ServeDNS(ctx context.Context, w dns.MessageWriter, r *dns.Query) {

	var found bool

	logMsg := log.WithDNSQuestions(r.Questions)

	res, err := d.dnsClient.Do(context.Background(), r)
	if err != nil {
		logMsg.WithDNSNXDomain().WithError(err).Error("dns client forwarding failed")
		w.Status(dns.NXDomain)
		return
	}

	for _, upstreamDNS := range res.Answers {

		var matchRegex bool
		if d.dnsRegex.MatchString(upstreamDNS.Name) {
			matchRegex = true
		}

		if upstreamDNS.Record.Type() == dns.TypeA && matchRegex {
			logMsg.WithDNSAnswer(upstreamDNS.Name, time.Minute, d.record)
			w.Answer(upstreamDNS.Name, time.Minute, d.record)
			found = true
		} else if upstreamDNS.Record.Type() == dns.TypeAAAA && matchRegex {
			logMsg.WithField("ignored_aaaa", true)
		} else {
			logMsg.WithDNSAnswer(upstreamDNS.Name, upstreamDNS.TTL, upstreamDNS.Record)
			w.Answer(upstreamDNS.Name, upstreamDNS.TTL, upstreamDNS.Record)
			found = true
		}
	}

	if !found {
		logMsg.WithDNSNXDomain()
		w.Status(dns.NXDomain)
	}

	logMsg.Info("")
}
