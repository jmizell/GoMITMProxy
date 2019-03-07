// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/benburkert/dns"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

const DefaultDNSServer = "8.8.8.8"

var dnsTypes = map[dns.Type]string {
	dns.TypeA: "A",
	dns.TypeNS: "NS",
	dns.TypeCNAME: "CNAME",
	dns.TypeSOA: "SOA",
	dns.TypeWKS: "WKS",
	dns.TypePTR: "PTR",
	dns.TypeHINFO: "HINFO",
	dns.TypeMINFO: "MINFO",
	dns.TypeMX: "MX",
	dns.TypeTXT: "TXT",
	dns.TypeAAAA: "AAAA",
	dns.TypeSRV: "SRV",
	dns.TypeDNAME: "DNAME",
	dns.TypeOPT: "OPT",
	dns.TypeAXFR: "AXFR",
	dns.TypeALL: "ALL",
	dns.TypeCAA: "CAA",
	dns.TypeANY: "ANY",
}

var dnsClass = map[dns.Class]string {
	dns.ClassIN: "IN",
	dns.ClassCH: "CH",
	dns.ClassHS: "HS",
	dns.ClassANY: "ANY",
}

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

	res, err := d.dnsClient.Do(context.Background(), r)
	if err != nil {
		log.WithError(err).Error("dns client forwarding failed")
	}

	for _, upstreamDNS := range res.Answers {

		var matchRegex bool
		if d.dnsRegex.MatchString(upstreamDNS.Name) {
			matchRegex = true
		}

		if upstreamDNS.Record.Type() == dns.TypeA && matchRegex {
			log.WithField("req_name", upstreamDNS.Name).
				WithField("req_type", DNSTypeString(upstreamDNS.Type())).
				WithField("answer_record", DNSRecordString(d.record)).
				WithField("answer_ttl", time.Minute).
				Debug("[DNS]")

			w.Answer(upstreamDNS.Name, time.Minute, d.record)
			found = true
		} else if upstreamDNS.Record.Type() == dns.TypeAAAA && matchRegex {
				log.WithField("req_name", upstreamDNS.Name).
					WithField("req_type", DNSTypeString(upstreamDNS.Type())).
					Debug("[DNS] Ignoring IPV6 AAAA")
		} else {
			log.WithField("req_name", upstreamDNS.Name).
				WithField("req_type", DNSTypeString(upstreamDNS.Type())).
				WithField("answer_record", DNSRecordString(upstreamDNS.Record)).
				WithField("answer_ttl", upstreamDNS.TTL).
				Debug("[DNS]")

			w.Answer(upstreamDNS.Name, upstreamDNS.TTL, upstreamDNS.Record)
			found = true
		}
	}

	if !found {
		log.WithField("req_questions", DNSQuestionsString(r.Questions)).
			Debug("[DNS] NXDomain")

		w.Status(dns.NXDomain)
	}
}

func DNSTypeString(t dns.Type) string {

	if name, ok := dnsTypes[t]; ok {
		return name
	}

	return "UNKNOWN"
}

func DNSClassString(c dns.Class) string {

	if name, ok := dnsClass[c]; ok {
		return name
	}

	return "UNKNOWN"
}

func DNSRecordString(r dns.Record) string {

	data, _ := json.Marshal(r)
	return strings.Replace(string(data), `"`, "", -1)
}

func DNSQuestionsString(questions []dns.Question) string {

	var qSlice []string
	for _, q := range questions {
		questionStr := fmt.Sprintf(
			"{Name:%s,Type:%s,Class:%v}",
			q.Name,
			DNSTypeString(q.Type),
			DNSClassString(q.Class))
		qSlice = append(qSlice, questionStr)
	}

	return fmt.Sprintf("[%s]", strings.Join(qSlice, ","))
}