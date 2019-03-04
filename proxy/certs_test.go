// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

func TestCerts_Get(t *testing.T) {

	t.Parallel()

	t.Run("single_query", func(subTest *testing.T) {

		certStore, err := getTestCertStore()
		if err != nil {
			t.Fatalf("expected GenerateCAPair to not return an error, received %s", err.Error())
		}

		if certStore.certStore != nil {
			t.Fatalf("expected cert store to have 0 entries")
		}

		hostKey, err := certStore.Get("example.com")
		if err != nil {
			t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
		}
		_ = validateCert(t, hostKey, "example.com")

		if len(certStore.certStore) != 1 {
			t.Fatalf("expected cert store to have 1 entries, found %d", len(certStore.certStore))
		}
	})

	t.Run("double_query", func(subTest *testing.T) {

		certStore, err := getTestCertStore()
		if err != nil {
			t.Fatalf("expected GenerateCAPair to not return an error, received %s", err.Error())
		}

		if certStore.certStore != nil {
			t.Fatalf("expected cert store to have 0 entries")
		}

		hostKey1, err := certStore.Get("example.com")
		if err != nil {
			t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
		}
		parsedKey1 := validateCert(t, hostKey1, "example.com")

		hostKey2, err := certStore.Get("example.com")
		if err != nil {
			t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
		}
		parsedKey2 := validateCert(t, hostKey2, "example.com")

		if parsedKey1.SerialNumber.String() != parsedKey2.SerialNumber.String() {
			t.Fatal("key serial numbers do not match")
		}

		if len(certStore.certStore) != 1 {
			t.Fatalf("expected cert store to have 1 entries, found %d", len(certStore.certStore))
		}
	})

	t.Run("concurrent_query", func(subTest *testing.T) {

		certStore, err := getTestCertStore()
		if err != nil {
			t.Fatalf("expected GenerateCAPair to not return an error, received %s", err.Error())
		}

		if certStore.certStore != nil {
			t.Fatalf("expected cert store to have 0 entries")
		}

		var hostKey1 *tls.Certificate
		var hostKey2 *tls.Certificate
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			hostKey1, err = certStore.Get("example.com")
			if err != nil {
				t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			hostKey2, err = certStore.Get("example.com")
			if err != nil {
				t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
			}
		}()

		wg.Wait()
		parsedKey1 := validateCert(t, hostKey1, "example.com")
		parsedKey2 := validateCert(t, hostKey2, "example.com")

		if parsedKey1.SerialNumber.String() != parsedKey2.SerialNumber.String() {
			t.Fatal("key serial numbers do not match")
		}

		if len(certStore.certStore) != 1 {
			t.Fatalf("expected cert store to have 1 entries, found %d", len(certStore.certStore))
		}
	})

	t.Run("mixed_queries", func(subTest *testing.T) {

		certStore, err := getTestCertStore()
		if err != nil {
			t.Fatalf("expected GenerateCAPair to not return an error, received %s", err.Error())
		}

		if certStore.certStore != nil {
			t.Fatalf("expected cert store to have 0 entries")
		}

		hostKey1, err := certStore.Get("example.com")
		if err != nil {
			t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
		}
		parsedKey1 := validateCert(t, hostKey1, "example.com")

		if len(certStore.certStore) != 1 {
			t.Fatalf("expected cert store to have 1 entries, found %d", len(certStore.certStore))
		}

		hostKeya, err := certStore.Get("example.net")
		if err != nil {
			t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
		}
		_ = validateCert(t, hostKeya, "example.net")

		hostKey2, err := certStore.Get("example.com")
		if err != nil {
			t.Fatalf("expected Certs.Get to not return an error, received %s", err.Error())
		}
		parsedKey2 := validateCert(t, hostKey2, "example.com")

		if parsedKey1.SerialNumber.String() != parsedKey2.SerialNumber.String() {
			t.Fatal("key serial numbers do not match")
		}

		if len(certStore.certStore) != 2 {
			t.Fatalf("expected cert store to have 2 entries, found %d", len(certStore.certStore))
		}
	})
}

func TestCerts_LoadCAPair(t *testing.T) {

	t.Parallel()

	certStore := &Certs{}
	certFile, keyFile, err := testingCAPair()
	if err != nil {
		t.Fatalf("failed to create testing ca key pair, %s", err.Error())
	}

	err = certStore.LoadCAPair(keyFile, certFile)
	if err != nil {
		t.Fatalf("failed to load testing ca key pair, %s", err.Error())
	}
}

func getTestCertStore() (*Certs, error) {

	certStore := &Certs{}
	key, cert, err := certStore.GenerateCAPair()
	certStore.caCert = cert
	certStore.caKey = key

	return certStore, err
}

func validateCert(t *testing.T, cert *tls.Certificate, domain string) *x509.Certificate {

	if len(cert.Certificate) != 1 {
		t.Fatalf("expected certificate to contain one public cert, found %d", len(cert.Certificate))
	}

	parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("expected x509.ParseCertificate to not return an error, received %s", err.Error())
	}

	var foundDomain bool
	for _, dnsName := range parsedCert.DNSNames {
		if dnsName == domain {
			foundDomain = true
		}
	}
	if !foundDomain {
		t.Fatalf("returned certificate was not valid for domain requested")
	}

	if time.Now().After(parsedCert.NotAfter) {
		t.Fatalf("returned certificate is expired")
	}

	return parsedCert
}

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
