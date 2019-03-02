// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/benburkert/dns"
)

const HTTPServerFatal = 129
const TLSServerFatal = 130
const DNSServerFatal = 132

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
}

func (p *Proxy) Run() (err error) {

	p.certs = &Certs{}
	if p.CACertFile == "" || p.CAKeyFile == "" {
		p.certs.caKey, p.certs.caCert, err = p.certs.GenerateCAPair()
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
				log.Printf("fatal dns server: %s", err.Error())
				os.Exit(DNSServerFatal)
			}
		}()
	}

	go func() {
		if err := p.serve(); err != nil {
			log.Printf("fatal http server: %s", err.Error())
			os.Exit(HTTPServerFatal)
		}
	}()

	if err := p.serveTLS(); err != nil {
		log.Printf("fatal https server: %s", err.Error())
		os.Exit(TLSServerFatal)
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
	log.Printf("tls server listening on %s:%d", ip, p.HTTPSPort)

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
	log.Printf("http server listening on %s:%d", ip, p.HTTPPort)

	return p.httpServer.Serve(connection)
}

func (p *Proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	//body, err := ioutil.ReadAll(req.Body)
	//if err != nil {
	//	log.Printf("error reading request body: %s", err.Error())
	//	return
	//}
	//req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	req.URL.Host = req.Host
	req.URL.Scheme = "http"
	if req.TLS != nil {
		req.URL.Scheme = "https"
	}

	proxy := httputil.NewSingleHostReverseProxy(req.URL)
	proxy.ServeHTTP(resp, req)
	log.Printf("proxy: url=%s", req.URL.String())
}

func (p *Proxy) sniLookup(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return p.certs.Get(clientHello.ServerName)
}

func WriteCA(certFileName, keyFileName string, cert *x509.Certificate, key *rsa.PrivateKey) error {
	if certFileName == "" || keyFileName == "" {
		startTime := time.Now().Unix()
		certFileName = fmt.Sprintf("gomitmproxy_ca_%d.crt", startTime)
		keyFileName = fmt.Sprintf("gomitmproxy_ca_%d.key", startTime)
	}

	keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	err := ioutil.WriteFile(keyFileName, keyBytes, 0600)
	if err != nil {
		return err
	}
	log.Printf("wrote certificate authority key to %s", keyFileName)

	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	err = ioutil.WriteFile(certFileName, certBytes, 0600)
	if err != nil {
		return err
	}
	log.Printf("wrote certificate authority cert to %s", certFileName)

	return nil
}
