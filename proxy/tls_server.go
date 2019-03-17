// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	standardLogger "log"
	"net"
	"net/http"

	"golang.org/x/net/http2"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

// TLSServer handles incoming https and http/2 requests for MITMProxy
type TLSServer struct {
	server    *http.Server
	tlsConfig *tls.Config

	ListenAddr string // TCP address for the server to listen on
	Port       int    // TCP Port of the server to listen on
	Certs      *Certs // Certificate cache
}

// ListenAndServe creates the server process, and blocks until an error occurs. A ready channel is used to signal
// when the server has started listening.
func (p *TLSServer) ListenAndServe(ready chan bool, handler http.Handler) error {

	reader, writer := io.Pipe()
	defer func() {
		_ = reader.Close()
		_ = writer.Close()
	}()
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			log.WithField("server", "https").
				WithField("port", p.Port).
				Error("[SERVER] %s", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.WithError(err).Error("exception reading from log writer")
		}
	}()

	p.tlsConfig = &tls.Config{
		GetCertificate: p.sniLookup,
	}

	p.server = &http.Server{
		TLSConfig: p.tlsConfig,
		Handler:   handler,
		ErrorLog:  standardLogger.New(writer, "", 0),
	}

	err := http2.ConfigureServer(p.server, nil)
	if err != nil {
		return err
	}

	listenAddress := fmt.Sprintf("%s:", p.ListenAddr)
	if p.Port > 0 {
		listenAddress = fmt.Sprintf("%s:%d", p.ListenAddr, p.Port)
	}
	connection, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(connection, p.tlsConfig)

	p.Port = connection.Addr().(*net.TCPAddr).Port
	ip := connection.Addr().(*net.TCPAddr).IP
	log.WithField("addr", fmt.Sprintf("%s:%d", ip, p.Port)).
		Info("https server started")

	ready <- true
	return p.server.Serve(tlsListener)
}

// GetPort returns the Port that TLSServer will listen to. In the case that Port is a nil value, this value will change
// to a randomly selected Port after calling ListenAndServe
func (p *TLSServer) GetPort() int {

	return p.Port
}

// Shutdown calls Shutdown on the listening server with a background context
func (p *TLSServer) Shutdown() error {

	if p.server != nil {
		return p.server.Shutdown(context.Background())
	}

	return nil
}

func (p *TLSServer) sniLookup(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	log.WithField("server_name", clientHello.ServerName).Debug("[SNI] lookup with client hello")
	return p.Certs.Get(clientHello.ServerName)
}
