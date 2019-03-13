// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"net"
	"net/http"
	"time"

	"github.com/benburkert/dns"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

// Exit code returned when the proxy server fails
const EXITCODEProxyServer = 129

// Exit code returns when the dns server fails
const EXITCODEDNSServer = 132

// Error returned when https server fails to start
const ERRTLSProxyStart = ErrorStr("https server failed to start")

// Error returned when http server fails to start
const ERRHTTPProxyStart = ErrorStr("http server failed to start")

// Error returned when shutting down the http and https servers fails
const ERRProxyShutdown = ErrorStr("proxy shutdown failed")

// server represents a tls or http server to be used in MITMProxy
type Server interface {
	ListenAndServe(chan bool, http.Handler) error
	GetPort() int
	Shutdown() error
}

// MITMProxy handles the creation of servers, and the transport of requests to their targets. A zero value is a no-op,
// a valid config that starts servers is minimally at least one Port specification in HTTPSPorts, or HTTPPorts.
type MITMProxy struct {
	servers      []Server
	serverErrors chan error

	// LogResponses enabled logging the the response with the request.
	LogResponses bool `json:"log_responses"`

	// Certs is the store for the root CA, and all automatically generated host keys.
	// Uninitialized a new Certs struct will be created either from the certificate files
	// provided in CAKeyFile, and CACertFile, or if those are empty, new ones will be
	// generated and written to disk.
	Certs      *Certs `json:"-"`
	CAKeyFile  string `json:"ca_key_file"`  // File path to the CA key file
	CACertFile string `json:"ca_cert_file"` // File path to the CA certificate file

	ListenAddr string `json:"listen_addr"` // TCP address to listen on
	HTTPSPorts []int  `json:"https_ports"` // List of ports to start a tls server on
	HTTPPorts  []int  `json:"http_ports"`  // List of ports to start http server on

	DNSPort  int    `json:"dns_port"`  // Port to start listening for dns requests on, a zero value disables the server
	DNSRegex string `json:"dns_regex"` // A regex pattern representing the vhosts to redirect to the proxy

	// ForwardDNSServer overrides net.DefaultResolver with this dns server address.
	// If the proxy is also serving dns, this value will be used for forwarded dns requests.
	DNSServer string `json:"dns_server"` // Override net.DefaultResolver with this dns server address

	// ProxyTransport is the http.Handler that receives requests, and performs the round trip.
	// If this value is nil, then ReverseProxy is used.
	//
	// When a custom handler is set, logging will also need to be implemented, as LogResponses is not
	// passed to the custom handler.
	ProxyTransport http.Handler `json:"-"`
}

// NewProxyWithDefaults returns a proxy with the default values used in gomitmproxy
func NewProxyWithDefaults() *MITMProxy {
	return &MITMProxy{
		CAKeyFile:  "",
		CACertFile: "",
		ListenAddr: "127.0.0.1",
		DNSServer:  "",
		DNSPort:    0,
		DNSRegex:   "",
		HTTPSPorts: []int{0},
		HTTPPorts:  []int{0},
	}
}

// Run creates the certificate store if not present, starts the dns server if configured, updates the
// net.DefaultResolver if ForwardDNSServer is set, and starts the tls and http servers. It blocks until the first
// server error is received, or the servers exit.
func (p *MITMProxy) Run() (err error) {

	if p.ProxyTransport == nil {
		p.ProxyTransport = &ReverseProxy{LogResponses: p.LogResponses}
	}

	if p.ListenAddr == "" {
		p.ListenAddr = "127.0.0.1"
	}

	p.serverErrors = make(chan error, len(p.HTTPSPorts)+len(p.HTTPPorts))

	if p.Certs == nil {
		p.Certs = &Certs{}
		if p.CACertFile == "" || p.CAKeyFile == "" {
			_, _, err := p.Certs.GenerateCAPair()
			if err != nil {
				return err
			}
			err = p.Certs.WriteCA(p.CACertFile, p.CAKeyFile)
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
		log.WithError(err).WithExitCode(EXITCODEProxyServer).Fatal("proxy server start failed")
	}

	err = <-p.serverErrors
	if err != http.ErrServerClosed && err != nil {
		log.WithError(err).Fatal("server exited with unexpected error")
		return p.Shutdown()
	}

	return nil
}

func (p *MITMProxy) runDNSServer() {

	dnsServer := DNSServer{
		ListenAddr:       p.ListenAddr,
		Port:             p.DNSPort,
		ForwardDNSServer: p.DNSServer,
		DNSRegex:         p.DNSRegex,
	}

	if err := dnsServer.ListenAndServe(); err != nil {
		log.WithError(err).WithExitCode(EXITCODEDNSServer).Fatal("dns server failed")
	}
}

func (p *MITMProxy) runProxyServers() error {

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

func (p *MITMProxy) runTLSServer(port int) (*TLSServer, error) {

	ready := make(chan bool, 1)
	srv := &TLSServer{
		ListenAddr: p.ListenAddr,
		Port:       port,
		Certs:      p.Certs,
	}

	go func() {
		p.serverErrors <- srv.ListenAndServe(ready, p.ProxyTransport)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		return nil, ERRTLSProxyStart.Err().WithReason("timed out waiting %s:%d to be ready", p.ListenAddr, srv.GetPort())
	}

	return srv, nil
}

func (p *MITMProxy) runHTTPServer(port int) (*HTTPServer, error) {

	ready := make(chan bool, 1)
	srv := &HTTPServer{
		ListenAddr: p.ListenAddr,
		Port:       port,
	}

	go func() {
		p.serverErrors <- srv.ListenAndServe(ready, p.ProxyTransport)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		return nil, ERRHTTPProxyStart.Err().WithReason("timed out waiting %s:%d to be ready", p.ListenAddr, srv.GetPort())
	}

	return srv, nil
}

// Shutdown signals all servers to exit, and waits for them to return, or errors after a 10 second timeout.
func (p *MITMProxy) Shutdown() (err error) {

	done := make(chan error, len(p.servers))
	for _, srv := range p.servers {
		go func(s Server) {
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
			return ERRProxyShutdown.Err().WithReason("timeout")
		}
	}

	return nil
}
