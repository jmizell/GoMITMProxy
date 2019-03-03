// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/benburkert/dns"
)

const HTTPServerFatal = 129
const TLSServerFatal = 130
const DNSServerFatal = 132

var defaultProxyHandler = func(url *url.URL) http.Handler {
	return httputil.NewSingleHostReverseProxy(url)
}

type Proxy struct {
	tlsServer  *http.Server
	tlsConfig  *tls.Config
	httpServer *http.Server
	certs      *Certs

	ListenAddr string `json:"listen_addr"`
	HTTPSPort  int    `json:"https_port"`
	HTTPPort   int    `json:"http_port"`
	CAKeyFile  string `json:"ca_key_file"`
	CACertFile string `json:"ca_cert_file"`
	DNSPort    int    `json:"dns_port"`
	DNSServer  string `json:"dns_server"`
	DNSRegex   string `json:"dns_regex"`

	newProxy func(*url.URL) http.Handler
}

func (p *Proxy) Run() (err error) {
	defer Log.Close()

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
		dnsServer := DNSServer{
			ListenAddr: p.ListenAddr,
			DNSPort:    p.DNSPort,
			DNSServer:  p.DNSServer,
			DNSRegex:   p.DNSRegex,
		}
		go func() {
			if err := dnsServer.ListenAndServe(); err != nil {
				Log.WithError(err).WithExitCode(DNSServerFatal).Fatal("dns server failed")
			}
		}()
	}

	go func() {
		if err := p.serve(); err != http.ErrServerClosed && err != nil {
			Log.WithError(err).WithExitCode(HTTPServerFatal).Fatal("http server failed")
		}
	}()

	if err := p.serveTLS(); err != http.ErrServerClosed && err != nil {
		Log.WithError(err).WithExitCode(TLSServerFatal).Fatal("https server failed")
	}

	return nil
}

func (p *Proxy) serveTLS() (err error) {

	p.tlsServer = &http.Server{}

	p.tlsConfig = &tls.Config{}
	p.tlsConfig.GetCertificate = p.sniLookup

	listenAddress := fmt.Sprintf("%s:", p.ListenAddr)
	if p.HTTPSPort > 0 {
		listenAddress = fmt.Sprintf("%s:%d", p.ListenAddr, p.HTTPSPort)
	}
	connection, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(connection, p.tlsConfig)

	p.tlsServer.Handler = p

	p.HTTPSPort = connection.Addr().(*net.TCPAddr).Port
	ip := connection.Addr().(*net.TCPAddr).IP
	Log.WithField("addr", fmt.Sprintf("%s:%d", ip, p.HTTPSPort)).
		Info("https server started")

	return p.tlsServer.Serve(tlsListener)
}

func (p *Proxy) serve() error {

	p.httpServer = &http.Server{}

	listenAddress := fmt.Sprintf("%s:", p.ListenAddr)
	if p.HTTPPort > 0 {
		listenAddress = fmt.Sprintf("%s:%d", p.ListenAddr, p.HTTPPort)
	}
	connection, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	p.httpServer.Handler = p

	p.HTTPPort = connection.Addr().(*net.TCPAddr).Port
	ip := connection.Addr().(*net.TCPAddr).IP
	Log.WithField("addr", fmt.Sprintf("%s:%d", ip, p.HTTPPort)).
		Info("http server started")

	return p.httpServer.Serve(connection)
}

func (p *Proxy) Shutdown() (err error) {

	if p.httpServer != nil {
		httpErr := p.httpServer.Shutdown(context.Background())
		if httpErr != nil {
			err = httpErr
		}
	}

	if p.tlsServer != nil {
		tlsErr := p.tlsServer.Shutdown(context.Background())
		if tlsErr != nil {
			err = tlsErr
		}
	}

	return err
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

func (p *Proxy) sniLookup(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return p.certs.Get(clientHello.ServerName)
}
