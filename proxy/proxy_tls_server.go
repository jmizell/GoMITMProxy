// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

type TLSProxyServer struct {
	server     *http.Server
	tlsConfig  *tls.Config
	listenAddr string
	port       int
	certs      *Certs
}

func (p *TLSProxyServer) ListenAndServe(ready chan bool, handler http.Handler) error {

	p.server = &http.Server{}

	p.tlsConfig = &tls.Config{}
	p.tlsConfig.GetCertificate = p.sniLookup

	listenAddress := fmt.Sprintf("%s:", p.listenAddr)
	if p.port > 0 {
		listenAddress = fmt.Sprintf("%s:%d", p.listenAddr, p.port)
	}
	connection, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(connection, p.tlsConfig)

	p.server.Handler = handler

	p.port = connection.Addr().(*net.TCPAddr).Port
	ip := connection.Addr().(*net.TCPAddr).IP
	log.WithField("addr", fmt.Sprintf("%s:%d", ip, p.port)).
		Info("https server started")

	ready <- true
	return p.server.Serve(tlsListener)
}

func (p *TLSProxyServer) GetPort() int {

	return p.port
}

func (p *TLSProxyServer) Shutdown() error {

	if p.server != nil {
		return p.server.Shutdown(context.Background())
	}

	return nil
}

func (p *TLSProxyServer) sniLookup(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return p.certs.Get(clientHello.ServerName)
}
