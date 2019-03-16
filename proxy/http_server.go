// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	standardLogger "log"
	"net"
	"net/http"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

// HTTPServer handles incoming http requests for MITMProxy
type HTTPServer struct {
	server *http.Server

	ListenAddr string // TCP address for the server to listen on
	Port       int    // TCP Port of the server to listen on
}

// ListenAndServe creates the server process, and blocks until an error occurs. A ready channel is used to signal
// when the server has started listening.
func (p *HTTPServer) ListenAndServe(ready chan bool, handler http.Handler) error {

	reader, writer := io.Pipe()
	defer func() {
		_ = reader.Close()
		_ = writer.Close()
	}()
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			log.WithField("server", "http").
				WithField("port", p.Port).
				Error(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.WithError(err).Error("exception reading from log writer")
		}
	}()

	p.server = &http.Server{
		Handler:  handler,
		ErrorLog: standardLogger.New(writer, "", 0),
	}

	listenAddress := fmt.Sprintf("%s:", p.ListenAddr)
	if p.Port > 0 {
		listenAddress = fmt.Sprintf("%s:%d", p.ListenAddr, p.Port)
	}
	connection, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	p.Port = connection.Addr().(*net.TCPAddr).Port
	ip := connection.Addr().(*net.TCPAddr).IP
	log.WithField("addr", fmt.Sprintf("%s:%d", ip, p.Port)).
		Info("http server started")

	ready <- true
	return p.server.Serve(connection)
}

// GetPort returns the Port that TLSServer will listen to. In the case that Port is a nil value, this value will change
// to a randomly selected Port after calling ListenAndServe
func (p *HTTPServer) GetPort() int {

	return p.Port
}

// Shutdown calls Shutdown on the listening server with a background context
func (p *HTTPServer) Shutdown() error {

	if p.server != nil {
		return p.server.Shutdown(context.Background())
	}

	return nil
}
