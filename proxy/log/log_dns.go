// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"time"

	"github.com/benburkert/dns"
)

var dnsTypes = map[dns.Type]string{
	dns.TypeA:     "A",
	dns.TypeNS:    "NS",
	dns.TypeCNAME: "CNAME",
	dns.TypeSOA:   "SOA",
	dns.TypeWKS:   "WKS",
	dns.TypePTR:   "PTR",
	dns.TypeHINFO: "HINFO",
	dns.TypeMINFO: "MINFO",
	dns.TypeMX:    "MX",
	dns.TypeTXT:   "TXT",
	dns.TypeAAAA:  "AAAA",
	dns.TypeSRV:   "SRV",
	dns.TypeDNAME: "DNAME",
	dns.TypeOPT:   "OPT",
	dns.TypeAXFR:  "AXFR",
	dns.TypeALL:   "ALL",
	dns.TypeCAA:   "CAA",
	dns.TypeANY:   "ANY",
}

var dnsClass = map[dns.Class]string{
	dns.ClassIN:  "IN",
	dns.ClassCH:  "CH",
	dns.ClassHS:  "HS",
	dns.ClassANY: "ANY",
}

func dnsTypeString(t dns.Type) string {

	if name, ok := dnsTypes[t]; ok {
		return name
	}

	return "UNKNOWN"
}

func dnsClassString(c dns.Class) string {

	if name, ok := dnsClass[c]; ok {
		return name
	}

	return "UNKNOWN"
}

type DNSRecord struct {
	Questions []*DNSQuestion `json:"questions"`
	Answers   []*DNSAnswer   `json:"answers"`
}

func (d *DNSRecord) AddQuestions(questions []dns.Question) {

	for _, question := range questions {
		q := &DNSQuestion{}
		q.Load(question)
		d.Questions = append(d.Questions, q)
	}
}

func (d *DNSRecord) AddAnswer(name string, ttl time.Duration, record dns.Record) {

	d.Answers = append(d.Answers, &DNSAnswer{Name: name, TTL: ttl, Record: record})
}

func (d *DNSRecord) AddNXDomain() {

	d.Answers = append(d.Answers, &DNSAnswer{NXDomain: "NXDomain"})
}

type DNSAnswer struct {
	Name     string        `json:"name,omitempty"`
	Record   interface{}   `json:"record,omitempty"`
	TTL      time.Duration `json:"ttl,omitempty"`
	NXDomain string        `json:"nx_domain,omitempty"`
}

type DNSQuestion struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Class string `json:"class"`
}

func (q *DNSQuestion) Load(question dns.Question) {

	q.Name = question.Name
	q.Type = dnsTypeString(question.Type)
	q.Class = dnsClassString(question.Class)
}
