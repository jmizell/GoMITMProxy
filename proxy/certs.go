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
	"sync"
	"time"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

const CertOrg = "GoMITMProxy"
const KeyLength = 1024
const DefaultKeyAge = time.Hour * 24

const ERRCertNoCA = ErrorStr("no certificate authority set")
const ERRCertCARead = ErrorStr("read ca file failed")
const ERRCertCAParse = ErrorStr("parse ca failed")
const ERRCertCAExpired = ErrorStr("ca expired")
const ERRCertGenCA = ErrorStr("generate ca failed")
const ERRCertGenHostKey = ErrorStr("generate host key failed")
const ERRCertWriteCA = ErrorStr("writing ca to disk failed")
const ERRCertGenerateKey = ErrorStr("generate key failed")
const ERRCertx509Create = ErrorStr("create x509 cert failed")
const ERRCertx509Parse = ErrorStr("parse x509 cert failed")

const EXITCODECertFatal = 131

type Certs struct {
	certStore map[string]*tls.Certificate
	caKey     *rsa.PrivateKey
	caCert    *x509.Certificate
	KeyAge    time.Duration `json:"key_age"`
	lock      sync.Mutex
}

func (c *Certs) Get(vhost string) (*tls.Certificate, error) {

	if vhost == "" {
		return nil, nil
	}

	if c.certStore == nil {
		c.certStore = map[string]*tls.Certificate{}
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	key, ok := c.certStore[vhost]
	if ok {
		return key, nil
	}

	key, err := c.GenerateHostKey(vhost)
	if err != nil {
		return nil, err
	}
	c.certStore[vhost] = key

	return key, nil
}

func (c *Certs) LoadCAPair(keyFile, certFile string) error {

	keyBytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return ERRCertCARead.Err().WithReason("%s - %s", keyFile, err.Error())
	}

	keyDecoded, _ := pem.Decode(keyBytes)
	if c.caKey, err = x509.ParsePKCS1PrivateKey(keyDecoded.Bytes); err != nil {
		return ERRCertCAParse.Err().WithReason("%s - %s", keyFile, err.Error())
	}

	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return ERRCertCARead.Err().WithReason("%s - %s", certFile, err.Error())
	}

	certDecoded, _ := pem.Decode(certBytes)
	if c.caCert, err = x509.ParseCertificate(certDecoded.Bytes); err != nil {
		return ERRCertCAParse.Err().WithReason("%s - %s", certFile, err.Error())
	}

	if c.caCert.NotAfter.Before(time.Now()) {
		return ERRCertCAExpired.Err()
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
		return nil, nil, ERRCertGenCA.Err().WithError(err)
	}

	return key, cert, nil
}

func (c *Certs) GenerateHostKey(vhost string) (*tls.Certificate, error) {

	if c.caKey == nil || c.caCert == nil {
		return nil, ERRCertNoCA.Err()
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
	if err != nil {
		return nil, ERRCertGenHostKey.Err().WithError(err)
	}

	tlsCert, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
	)
	if err != nil {
		return nil, ERRCertGenHostKey.Err().WithError(err)
	}

	return &tlsCert, nil
}

func genCerts(certTemplate *x509.Certificate, signingCert *x509.Certificate, signingKey *rsa.PrivateKey, KeyAge time.Duration) (
	key *rsa.PrivateKey, cert *x509.Certificate, err error) {

	if KeyAge == 0 {
		KeyAge = DefaultKeyAge
	}

	key, err = rsa.GenerateKey(rand.Reader, KeyLength)
	if err != nil {
		return nil, nil, ERRCertGenerateKey.Err().WithError(err)
	}

	if signingCert == nil || signingKey == nil {
		signingCert = certTemplate
		signingKey = key
	}

	signedCertBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, signingCert, &key.PublicKey, signingKey)
	if err != nil {
		return nil, nil, ERRCertx509Create.Err().WithError(err)
	}

	cert, err = x509.ParseCertificate(signedCertBytes)
	if err != nil {
		return nil, nil, ERRCertx509Parse.Err().WithError(err)
	}

	return key, cert, err
}

func genSerial() *big.Int {

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.WithError(err).WithExitCode(EXITCODECertFatal).Fatal("generate serial")
	}

	return serialNumber
}

func WriteCA(certFileName, keyFileName string, cert *x509.Certificate, key *rsa.PrivateKey) error {

	if certFileName == "" || keyFileName == "" {
		startTime := time.Now().Unix()
		certFileName = fmt.Sprintf("gomitmproxy_ca_%d.crt", startTime)
		keyFileName = fmt.Sprintf("gomitmproxy_ca_%d.key", startTime)
	}

	keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	err := ioutil.WriteFile(keyFileName, keyBytes, 0600)
	if err != nil {
		return ERRCertWriteCA.Err().WithError(err)
	}
	log.WithField("key_file", keyFileName).Info("wrote certificate authority key")

	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	err = ioutil.WriteFile(certFileName, certBytes, 0600)
	if err != nil {
		return ERRCertWriteCA.Err().WithError(err)
	}
	log.WithField("cert_file", certFileName).Info("wrote certificate authority certificate")

	return nil
}
