// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"time"
)

const CertOrg = "GoMITMProxy"
const KeyLength = 1024
const DefaultKeyAge = time.Hour * 24

const CertErrNoCA = "no certificate authority set"
const CertErrLoadCA = "error loading ca"

const GenCertFatal = 131

type Certs struct {
	certStore map[string]*tls.Certificate
	caKey     *rsa.PrivateKey
	caCert    *x509.Certificate
	KeyAge    time.Duration `json:"key_age"`
	lock      sync.Mutex
}

func (c *Certs) Get(vhost string) (*tls.Certificate, error) {

	if c.certStore == nil {
		c.certStore = map[string]*tls.Certificate{}
	}

	c.lock.Lock()
	key, ok := c.certStore[vhost]
	c.lock.Unlock()
	if ok {
		return key, nil
	}

	key, err := c.GenerateHostKey(vhost)
	if err != nil {
		return nil, err
	}
	c.lock.Lock()
	c.certStore[vhost] = key
	c.lock.Unlock()

	return key, nil
}

func (c *Certs) LoadCAPair(keyFile, certFile string) error {

	keyBytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("%s %s: %v", CertErrLoadCA, keyFile, err.Error())
	}

	keyDecoded, _ := pem.Decode(keyBytes)
	if c.caKey, err = x509.ParsePKCS1PrivateKey(keyDecoded.Bytes); err != nil {
		return fmt.Errorf("%s key parse error: %v", CertErrLoadCA, err.Error())
	}

	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("%s %s: %v", CertErrLoadCA, certFile, err.Error())
	}

	certDecoded, _ := pem.Decode(certBytes)
	if c.caCert, err = x509.ParseCertificate(certDecoded.Bytes); err != nil {
		return fmt.Errorf("%s cert parse error: %v", CertErrLoadCA, err.Error())
	}

	if c.caCert.NotAfter.Before(time.Now()) {
		return fmt.Errorf("%s cert expired", CertErrLoadCA)
	}

	c.KeyAge = c.caCert.NotAfter.Sub(time.Now())

	return nil
}

func (c *Certs) GenerateCAPair() (key *rsa.PrivateKey, cert *x509.Certificate, err error) {

	if c.KeyAge == 0 {
		c.KeyAge = DefaultKeyAge
	}

	CACertTemplate := &x509.Certificate{
		SerialNumber: genSerial(),
		Subject: pkix.Name{
			Organization:       []string{CertOrg},
			OrganizationalUnit: []string{CertOrg},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(c.KeyAge),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
	}

	key, cert, err = genCerts(CACertTemplate, c.caCert, c.caKey, DefaultKeyAge)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, err
}

func (c *Certs) GenerateHostKey(vhost string) (*tls.Certificate, error) {

	if c.caKey == nil || c.caCert == nil {
		return nil, fmt.Errorf(CertErrNoCA)
	}

	hostCertTemplate := &x509.Certificate{
		SerialNumber: genSerial(),
		Subject: pkix.Name{
			CommonName: vhost,
		},
		DNSNames:              []string{vhost},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(DefaultKeyAge),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  false,
	}

	key, cert, err := genCerts(hostCertTemplate, c.caCert, c.caKey, DefaultKeyAge)

	tlsCert, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
	)
	return &tlsCert, err
}

func genCerts(certTemplate *x509.Certificate, signingCert *x509.Certificate, signingKey *rsa.PrivateKey, KeyAge time.Duration) (
	key *rsa.PrivateKey, cert *x509.Certificate, err error) {

	if KeyAge == 0 {
		KeyAge = DefaultKeyAge
	}

	key, err = rsa.GenerateKey(rand.Reader, KeyLength)
	if err != nil {
		return nil, nil, err
	}

	if signingCert == nil || signingKey == nil {
		signingCert = certTemplate
		signingKey = key
	}

	signedCertBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, signingCert, &key.PublicKey, signingKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err = x509.ParseCertificate(signedCertBytes)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, err
}

func genSerial() *big.Int {

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		os.Exit(GenCertFatal)
	}

	return serialNumber
}
