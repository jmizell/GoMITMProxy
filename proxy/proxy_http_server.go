// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type HTTPProxyServer struct {
	server     *http.Server
	listenAddr string
	port       int
}

func (p *HTTPProxyServer) Serve(ready chan bool, handler http.Handler) error {

	p.server = &http.Server{}

	listenAddress := fmt.Sprintf("%s:", p.listenAddr)
	if p.port > 0 {
		listenAddress = fmt.Sprintf("%s:%d", p.listenAddr, p.port)
	}
	connection, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	p.server.Handler = handler

	p.port = connection.Addr().(*net.TCPAddr).Port
	ip := connection.Addr().(*net.TCPAddr).IP
	Log.WithField("addr", fmt.Sprintf("%s:%d", ip, p.port)).
		Info("http server started")

	ready <- true
	return p.server.Serve(connection)
}

func (p *HTTPProxyServer) GetPort() int {

	return p.port
}

func (p *HTTPProxyServer) Shutdown() error {

	if p.server != nil {
		return p.server.Shutdown(context.Background())
	}

	return nil
}
