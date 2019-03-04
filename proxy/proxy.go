// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/benburkert/dns"
)

const HTTPServerFatal = 129
const TLSServerFatal = 130
const DNSServerFatal = 132

var defaultProxyHandler = func(url *url.URL) http.Handler {
	return httputil.NewSingleHostReverseProxy(url)
}

type proxyServer interface {
	Serve(chan bool, http.Handler) error
	GetPort() int
	Shutdown() error
}

type Proxy struct {
	tlsServers   []proxyServer
	httpServers  []proxyServer
	serverErrors chan error
	certs        *Certs

	ListenAddr string `json:"listen_addr"`
	HTTPSPorts []int  `json:"https_ports"`
	HTTPPorts  []int  `json:"http_ports"`
	CAKeyFile  string `json:"ca_key_file"`
	CACertFile string `json:"ca_cert_file"`
	DNSPort    int    `json:"dns_port"`
	DNSServer  string `json:"dns_server"`
	DNSRegex   string `json:"dns_regex"`

	newProxy func(*url.URL) http.Handler
}

func (p *Proxy) Run() (err error) {
	defer Log.Close()

	p.serverErrors = make(chan error, len(p.HTTPSPorts)+len(p.HTTPPorts))

	if p.certs == nil {
		p.certs = &Certs{}
		if p.CACertFile == "" || p.CAKeyFile == "" {
			p.certs.caKey, p.certs.caCert, err = p.certs.GenerateCAPair()
			if err != nil {
				return err
			}
			err = WriteCA(p.CACertFile, p.CAKeyFile, p.certs.caCert, p.certs.caKey)
		} else {
			err = p.certs.LoadCAPair(p.CAKeyFile, p.CACertFile)
		}

		if err != nil {
			return err
		}
	}

	if p.DNSServer != "" {
		net.DefaultResolver = &net.Resolver{
			PreferGo: true,

			Dial: (&dns.Client{
				Transport: &dns.Transport{
					Proxy: dns.NameServers{
						&net.UDPAddr{IP: net.ParseIP(p.DNSServer), Port: 53},
					}.RoundRobin(),
				},
			}).Dial,
		}
	}

	if p.DNSPort > 0 {
		go p.serveDNS()
	}

	if err := p.runHTTPServers(); err != nil {
		Log.WithError(err).WithExitCode(HTTPServerFatal).Fatal("http server failed")
	}

	if err := p.runTLSServers(); err != nil {
		Log.WithError(err).WithExitCode(TLSServerFatal).Fatal("https server failed")
	}

	err = <-p.serverErrors
	if err != http.ErrServerClosed && err != nil {
		Log.WithError(err).Fatal("server exited with unexpected error")
		return p.Shutdown()
	}

	return nil
}

func (p *Proxy) serveDNS() {

	dnsServer := DNSServer{
		ListenAddr: p.ListenAddr,
		DNSPort:    p.DNSPort,
		DNSServer:  p.DNSServer,
		DNSRegex:   p.DNSRegex,
	}

	if err := dnsServer.ListenAndServe(); err != nil {
		Log.WithError(err).WithExitCode(DNSServerFatal).Fatal("dns server failed")
	}
}

func (p *Proxy) runTLSServers() error {

	var ports []int
	for _, port := range p.HTTPSPorts {

		ready := make(chan bool, 1)
		srv := &TLSProxyServer{
			listenAddr: p.ListenAddr,
			port:       port,
			certs:      p.certs,
		}

		go func() {
			p.serverErrors <- srv.Serve(ready, p)
		}()

		select {
		case <-ready:
		case <-time.After(1 * time.Second):
			return fmt.Errorf("timed out waiting for tls server %s:%d to be ready", p.ListenAddr, port)
		}

		ports = append(ports, srv.GetPort())
		p.tlsServers = append(p.tlsServers, srv)
	}
	p.HTTPSPorts = ports

	return nil
}

func (p *Proxy) runHTTPServers() error {

	var ports []int
	for _, port := range p.HTTPPorts {

		ready := make(chan bool, 1)
		srv := &HTTPProxyServer{
			listenAddr: p.ListenAddr,
			port:       port,
		}

		go func() {
			p.serverErrors <- srv.Serve(ready, p)
		}()

		select {
		case <-ready:
		case <-time.After(1 * time.Second):
			return fmt.Errorf("timed out waiting for http server %s:%d to be ready", p.ListenAddr, port)
		}

		ports = append(ports, srv.GetPort())
		p.httpServers = append(p.httpServers, srv)
	}
	p.HTTPPorts = ports

	return nil
}

func (p *Proxy) Shutdown() (err error) {

	servers := append(p.httpServers, p.tlsServers...)
	done := make(chan error, len(servers))
	for _, srv := range servers {
		go func(s proxyServer) {
			done <- s.Shutdown()
		}(srv)
	}

	for i := 0; i <= len(servers); i++ {
		select {
		case err := <-done:
			if err != http.ErrServerClosed && err != nil {
				return err
			}
		case <-time.After(time.Second * 10):
			return fmt.Errorf("timed out waiting for server to shutdown")
		}
	}

	return nil
}

func (p *Proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	if p.newProxy == nil {
		p.newProxy = defaultProxyHandler
	}

	req.URL.Host = req.Host
	req.URL.Scheme = "http"
	if req.TLS != nil {
		req.URL.Scheme = "https"
	}

	p.newProxy(req.URL).ServeHTTP(resp, req)
	Log.WithRequest(req).Info("")
}
