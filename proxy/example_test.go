// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy_test

import (
	"fmt"
	"github.com/jmizell/GoMITMProxy/proxy"
	"log"
)

func ExampleDNSServer() {

	// DNS Server listening on TCP and UDP port 53, network address 127.0.0.1,
	// forwarding requests to 1.1.1.1, except in the case of A record requests
	// for example.com, which are rewritten to point to 127.0.0.1.
	dnsServer := &proxy.DNSServer{
		ListenAddr:       "127.0.0.1",
		Port:             53,
		ForwardDNSServer: "1.1.1.1",
		DNSRegex:         ".*example.com",
	}

	err := dnsServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleCerts_GenerateCAPair() {

	// Create the cert store
	certs := &proxy.Certs{}

	// Generate a RSA private key, and x509 certificate for the cert authority
	_, _, err := certs.GenerateCAPair()
	if err != nil {
		log.Fatal(err)
	}

	// Write the key and certs as PEM encoded files to disk
	err = certs.WriteCA("/path/to/cert.crt", "/path/to/key.key")
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleMITMProxy() {

	// This creates a MITM proxy listening on TCP ports 80 and 443 network address
	// 127.0.0.1, and a dns server on port 53 forwarding requests to 1.1.1.1,
	// redirecting requests for example.com to 127.0.0.1.
	//
	// When CAKeyFile and CACertFile are empty, and a proxy.Certs struct isn't
	// supplied, the proxy will generate the keys, and write them to the current
	// working directory
	p := &proxy.MITMProxy{
		ListenAddr: "127.0.0.1",
		DNSServer:  "1.1.1.1",
		DNSPort:    53,
		DNSRegex:   ".*example.com",
		HTTPSPorts: []int{443},
		HTTPPorts:  []int{80},
	}

	// Run generates the CA, and starts listening for client connections
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func ExampleProxyError_castProxyError() {

	// Create a proxy error constant. This is the immutable representation of our errors that
	// we can later compare against to check error types.
	const customProxyErr = proxy.ErrorStr("custom error")

	// Create a proxy error with a embedded go error.
	newProxyError := customProxyErr.Err().WithError(fmt.Errorf("go error"))

	// Use the proxy error as a go error.
	var goError error
	goError = newProxyError

	// Now cast our go error back to a proxy error.
	var receivingProxyError proxy.ProxyError
	if err := receivingProxyError.UnmarshalError(goError); err != nil {
		fmt.Printf("couldn't unmarshal go error to proxy.ProxyError, %s", err.Error())
		return
	}

	// Compare the proxy error against the constant to see if they are the same type.
	if receivingProxyError.Match(customProxyErr.Err()) {
		fmt.Println("receivingProxyError matches customProxyErr type")
	}

	// Retrieve the embedded go error from the proxy error.
	fmt.Printf("receivingProxyError has error %s\n", receivingProxyError.GetError())
}

func ExampleProxyError_nestedErrors() {

	// Create a couple proxy error constants.
	const customProxyErr1 = proxy.ErrorStr("custom error one")
	const customProxyErr2 = proxy.ErrorStr("custom error two")

	// Create a new proxy error with a embedded proxy error.
	newProxyError := customProxyErr1.Err().WithError(customProxyErr2.Err())

	// Retrieve the embedded error and cast back to a proxy error.
	var embeddedProxyErr proxy.ProxyError
	if err := embeddedProxyErr.UnmarshalError(newProxyError.GetError()); err != nil {
		fmt.Printf("couldn't unmarshal go error to proxy.ProxyError, %s", err.Error())
		return
	}

	// Compare the embedded proxy error against the constant.
	if embeddedProxyErr.Match(customProxyErr2.Err()) {
		fmt.Println("receivingProxyError matches customProxyErr2 type")
	}
}
