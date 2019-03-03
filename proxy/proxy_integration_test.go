package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func testingCAPair() (certFilename string, keyFilename string, err error) {

	cert, err := ioutil.TempFile("", "*.crt")
	if err = cert.Close(); err != nil {
		return "", "", err
	}
	certFilename = cert.Name()

	key, err := ioutil.TempFile("", "*.key")
	if err = key.Close(); err != nil {
		return "", "", err
	}
	keyFilename = key.Name()

	certs := &Certs{}
	caKey, caCert, err := certs.GenerateCAPair()
	err = WriteCA(certFilename, keyFilename, caCert, caKey)
	if err != nil {
		return "", "", err
	}

	return certFilename, keyFilename, nil
}

func TestProxy_Run(t *testing.T) {

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

	if proxy.HTTPSPort == 0 {
		t.Fatalf("https port was not set")
	}

	if proxy.HTTPPort == 0 {
		t.Fatalf("http port was not set")
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", proxy.HTTPPort))
	if err != nil {
		t.Fatalf("failed to connect to test proxy, %s", err.Error())
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
