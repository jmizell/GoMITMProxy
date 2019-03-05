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

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

const ProxyServerFatal = 129
const DNSServerFatal = 132

var DefaultProxyHandler = func(url *url.URL) http.Handler {
	return httputil.NewSingleHostReverseProxy(url)
}

type proxyServer interface {
	ListenAndServe(chan bool, http.Handler) error
	GetPort() int
	Shutdown() error
}

type Proxy struct {
	servers      []proxyServer
	serverErrors chan error

	Certs      *Certs `json:"-"`
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

	p.serverErrors = make(chan error, len(p.HTTPSPorts)+len(p.HTTPPorts))

	if p.Certs == nil {
		p.Certs = &Certs{}
		if p.CACertFile == "" || p.CAKeyFile == "" {
			p.Certs.caKey, p.Certs.caCert, err = p.Certs.GenerateCAPair()
			if err != nil {
				return err
			}
			err = WriteCA(p.CACertFile, p.CAKeyFile, p.Certs.caCert, p.Certs.caKey)
		} else {
			err = p.Certs.LoadCAPair(p.CAKeyFile, p.CACertFile)
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
		go p.runDNSServer()
	}

	if err := p.runProxyServers(); err != nil {
		log.WithError(err).WithExitCode(ProxyServerFatal).Fatal("proxy server start failed")
	}

	err = <-p.serverErrors
	if err != http.ErrServerClosed && err != nil {
		log.WithError(err).Fatal("server exited with unexpected error")
		return p.Shutdown()
	}

	return nil
}

func (p *Proxy) runDNSServer() {

	dnsServer := DNSServer{
		ListenAddr: p.ListenAddr,
		DNSPort:    p.DNSPort,
		DNSServer:  p.DNSServer,
		DNSRegex:   p.DNSRegex,
	}

	if err := dnsServer.ListenAndServe(); err != nil {
		log.WithError(err).WithExitCode(DNSServerFatal).Fatal("dns server failed")
	}
}

func (p *Proxy) runProxyServers() error {

	var httpsPorts []int
	for _, port := range p.HTTPSPorts {

		srv, err := p.runTLSServer(port)
		if err != nil {
			return err
		}

		httpsPorts = append(httpsPorts, srv.GetPort())
		p.servers = append(p.servers, srv)
	}
	p.HTTPSPorts = httpsPorts

	var httpPorts []int
	for _, port := range p.HTTPPorts {

		srv, err := p.runHTTPServer(port)
		if err != nil {
			return err
		}

		httpPorts = append(httpPorts, srv.GetPort())
		p.servers = append(p.servers, srv)
	}
	p.HTTPPorts = httpPorts

	return nil
}

func (p *Proxy) runTLSServer(port int) (*TLSProxyServer, error) {

	ready := make(chan bool, 1)
	srv := &TLSProxyServer{
		listenAddr: p.ListenAddr,
		port:       port,
		certs:      p.Certs,
	}

	go func() {
		p.serverErrors <- srv.ListenAndServe(ready, p)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		return nil, fmt.Errorf("timed out waiting for tls server %s:%d to be ready", p.ListenAddr, srv.GetPort())
	}

	return srv, nil
}

func (p *Proxy) runHTTPServer(port int) (*HTTPProxyServer, error) {

	ready := make(chan bool, 1)
	srv := &HTTPProxyServer{
		listenAddr: p.ListenAddr,
		port:       port,
	}

	go func() {
		p.serverErrors <- srv.ListenAndServe(ready, p)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		return nil, fmt.Errorf("timed out waiting for http server %s:%d to be ready", p.ListenAddr, srv.GetPort())
	}

	return srv, nil
}

func (p *Proxy) Shutdown() (err error) {

	done := make(chan error, len(p.servers))
	for _, srv := range p.servers {
		go func(s proxyServer) {
			done <- s.Shutdown()
		}(srv)
	}

	for i := 0; i <= len(p.servers); i++ {
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
		p.newProxy = DefaultProxyHandler
	}

	req.URL.Host = req.Host
	req.URL.Scheme = "http"
	if req.TLS != nil {
		req.URL.Scheme = "https"
	}

	p.newProxy(req.URL).ServeHTTP(resp, req)
	log.WithRequest(req).Info("")
}
