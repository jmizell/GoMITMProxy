// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golang.org/x/net/http2"
	"io/ioutil"
	"net/http"
	"os"
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

	proxyHandler := &testProxyHandler{response: []byte("okay")}
	proxy := &MITMProxy{
		CAKeyFile:      keyFilename,
		CACertFile:     certFilename,
		ProxyTransport: proxyHandler,
		ListenAddr:     "localhost",
		HTTPPorts:      []int{0, 0},
		HTTPSPorts:     []int{0, 0},
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
		t.Fatalf("https Port was not set")
	}

	if len(proxy.HTTPPorts) != 2 {
		t.Fatalf("http Port was not set")
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
		for _, cert := range resp.TLS.PeerCertificates {
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

func TestProxy_Run_http2(t *testing.T) {

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

	proxyHandler := &testProxyHandler{response: []byte("okay")}
	proxy := &MITMProxy{
		CAKeyFile:      keyFilename,
		CACertFile:     certFilename,
		ProxyTransport: proxyHandler,
		ListenAddr:     "localhost",
		HTTPSPorts:     []int{0, 0},
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
		t.Fatalf("https Port was not set")
	}

	for _, proxyPort := range proxy.HTTPSPorts {
		rootCAs := x509.NewCertPool()
		rootCAs.AddCert(proxy.Certs.caCert)
		client := &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
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
		for _, cert := range resp.TLS.PeerCertificates {
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
		for _, chain := range resp.TLS.VerifiedChains {
			for _, cert := range chain {
				fmt.Println("DEBUG -- --", cert.DNSNames)
				fmt.Println("DEBUG -- --", cert.IsCA)
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
