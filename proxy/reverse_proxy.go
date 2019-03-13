// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"io"
	"net/http"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

// ReverseProxy is an http.Handler that that receives requests, performs the round trip,
// and handles logging.
type ReverseProxy struct {

	// LogResponses enabled logging the the response with the request.
	LogResponses bool `json:"log_responses"`

	// The transport used to perform proxy requests. If nil, http.DefaultTransport is used.
	Transport http.RoundTripper `json:"-"`
}

func (p *ReverseProxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	req.URL.Host = req.Host
	req.URL.Scheme = "http"
	if req.TLS != nil {
		req.URL.Scheme = "https"
	}

	logMsg := log.WithRequest(req)

	outRequest := req.WithContext(req.Context())
	if req.ContentLength == 0 {
		outRequest.Body = nil
	}

	outRequest.Header = req.Header
	outRequest.Close = false

	transport := p.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	roundTripResponse, err := transport.RoundTrip(outRequest)
	if err != nil {
		logMsg.WithError(err).Error("failed round trip")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if p.LogResponses {
		logMsg.WithResponse(roundTripResponse)
	}

	// TODO implement trailers

	for k, vv := range roundTripResponse.Header {
		for _, v := range vv {
			resp.Header().Add(k, v)
		}
	}

	resp.WriteHeader(roundTripResponse.StatusCode)
	logMsg.WithField("status_code", roundTripResponse.StatusCode)

	byteCount, err := io.Copy(resp, roundTripResponse.Body)
	if err != nil {
		logMsg.WithError(err).Error("failed to write response")
		return
	}

	logMsg.WithField("response_bytes", byteCount).Info("")
}
