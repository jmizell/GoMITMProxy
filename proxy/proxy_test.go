package proxy

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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

func TestProxy_ServeHTTP(t *testing.T) {

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
