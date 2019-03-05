// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProxy_Run(t *testing.T) {

	t.Parallel()

	certFilename, keyFilename, err := testingCAPair()
	if err != nil {
		t.Fatalf("failed to create testing ca key pair, %s", err.Error())
	}
	defer func() {
		if err := os.Remove(certFilename); err != nil {
			t.Fatalf("error deleting %s, %s", certFilename, err.Error())
		}
		if err := os.Remove(keyFilename); err != nil {
			t.Fatalf("error deleting %s, %s", keyFilename, err.Error())
		}
	}()

	var newProxyCalls []*url.URL
	proxyHandler := &testProxyHandler{response: []byte("okay")}
	proxy := &Proxy{
		CAKeyFile:  keyFilename,
		CACertFile: certFilename,
		newProxy: func(url *url.URL) http.Handler {
			newProxyCalls = append(newProxyCalls, url)
			return proxyHandler
		},
		ListenAddr: "localhost",
		HTTPPorts:  []int{0, 0},
		HTTPSPorts: []int{0, 0},
	}

	go func() {
		if err := proxy.Run(); err != nil {
			t.Fatalf("server exited non-nil error, %s", err.Error())
		}
	}()
	defer func() {
		_ = proxy.Shutdown()
	}()
	time.Sleep(time.Millisecond * 500)

	if len(proxy.HTTPSPorts) != 2 {
		t.Fatalf("https port was not set")
	}

	if len(proxy.HTTPPorts) != 2 {
		t.Fatalf("http port was not set")
	}

	for _, proxyPort := range proxy.HTTPPorts {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", proxyPort))
		if err != nil {
			t.Fatalf("failed to make a http connection to test proxy, %s", err.Error())
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected response code %d, but received %d", http.StatusOK, resp.StatusCode)
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("response body read returned error, %s", err.Error())
		}
		if string(respBody) != "okay" {
			t.Fatalf("expected response body to contain \"okay\", but received %s", string(respBody))
		}
	}

	for _, proxyPort := range proxy.HTTPSPorts {
		rootCAs := x509.NewCertPool()
		rootCAs.AddCert(proxy.Certs.caCert)
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: rootCAs,
				},
			},
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://localhost:%d", proxyPort), nil)
		if err != nil {
			t.Fatalf("failed to create http request, %s", err.Error())
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make a tls connection to test proxy, %s", err.Error())
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected response code %d, but received %d", http.StatusOK, resp.StatusCode)
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("response body read returned error, %s", err.Error())
		}
		if string(respBody) != "okay" {
			t.Fatalf("expected response body to contain \"okay\", but received %s", string(respBody))
		}

		var foundValidCert bool
		for _, chain := range resp.TLS.VerifiedChains {
			for _, cert := range chain {
				if !cert.IsCA {
					for _, domain := range cert.DNSNames {
						if domain == "localhost" {
							foundValidCert = true
						}
					}
					if time.Now().After(cert.NotAfter) {
						t.Fatalf("returned cert expired %s", cert.NotAfter.Local().String())
					}
				}
			}
		}
		if !foundValidCert {
			t.Fatal("no returned certificate contained a valid domain")
		}
	}
}

func TestProxy_ServeHTTP(t *testing.T) {

	t.Parallel()

	type ServeTests struct {
		name   string
		method string
		url    string
		body   string
	}
	tests := []ServeTests{
		{name: "http_get", method: "GET", url: "http://127.0.0.1:8081/test1", body: ""},
		{name: "https_get", method: "GET", url: "https://127.0.0.1:8082/test2", body: ""},
		{name: "http_post", method: "POST", url: "http://127.0.0.1:8083/test3", body: "test1"},
		{name: "https_post", method: "POST", url: "https://127.0.0.1:8084/test4", body: "test2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(subTest *testing.T) {

			var newProxyCalls []*url.URL
			proxyHandler := &testProxyHandler{response: []byte(test.name)}
			proxy := &Proxy{
				newProxy: func(url *url.URL) http.Handler {
					newProxyCalls = append(newProxyCalls, url)
					return proxyHandler
				},
			}

			req := httptest.NewRequest(test.method, test.url, strings.NewReader(test.body))
			resp := httptest.NewRecorder()
			proxy.ServeHTTP(resp, req)

			if len(newProxyCalls) != 1 {
				subTest.Fatalf("expected 1 call to newProxy, but recorded %d", len(newProxyCalls))
			}

			if len(proxyHandler.requests) != 1 {
				subTest.Fatalf("expected 1 request passed to proxy handler, but recorded %d", len(proxyHandler.requests))
			}

			if len(proxyHandler.responses) != 1 {
				subTest.Fatalf("expected 1 response passed to proxy handler, but recorded %d", len(proxyHandler.responses))
			}

			for _, e := range proxyHandler.writeErrors {
				if e != nil {
					subTest.Fatalf("expected no proxy handler write errors, but received %s", e.Error())
				}
			}

			if proxyHandler.requests[0].URL.String() != test.url {
				subTest.Fatalf("expected url %s, but recorded %s", test.url, proxyHandler.requests[0].URL.String())
			}

			if proxyHandler.requests[0].Method != test.method {
				subTest.Fatalf("expected method %s, but recorded %s", test.method, proxyHandler.requests[0].Method)
			}

			reqBody, err := ioutil.ReadAll(proxyHandler.requests[0].Body)
			if err != nil {
				subTest.Fatalf("reading from request body resulted in error, %s", err.Error())
			}
			if string(reqBody) != test.body {
				subTest.Fatalf("expected request body %s, but recorded %s", test.body, string(reqBody))
			}

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				subTest.Fatalf("reading from response body resulted in error, %s", err.Error())
			}
			if string(respBody) != test.name {
				subTest.Fatalf("expected response body %s, but recorded %s", test.name, string(respBody))
			}

		})
	}
}

type testProxyHandler struct {
	requests    []*http.Request
	responses   []http.ResponseWriter
	response    []byte
	writeErrors []error
}

func (t *testProxyHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	t.requests = append(t.requests, req)
	t.responses = append(t.responses, resp)
	_, err := resp.Write(t.response)
	t.writeErrors = append(t.writeErrors, err)
}
