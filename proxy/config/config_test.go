package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/jmizell/GoMITMProxy/proxy"
	"github.com/jmizell/GoMITMProxy/proxy/log"
)

var testConfigBytes []byte

func init() {

	var err error
	testConfigBytes, err = json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		log.WithError(err).Fatal("failed to unmarshal testConfig")
	}
	fmt.Println("DEBUG -- -- ", string(testConfigBytes))
}

func TestConfig_MITMProxy(t *testing.T) {

	t.Parallel()

	p := &proxy.MITMProxy{}
	err := json.Unmarshal(testConfigBytes, p)
	if err != nil {
		t.Fatalf("error unmarshalint testConfigBytes %s", err.Error())
	}

	if p.LogResponses != testConfig.LogResponses {
		t.Fatalf("expected %v, but found %v", testConfig.LogResponses, p.LogResponses)
	}

	if p.CAKeyFile != testConfig.CAKeyFile {
		t.Fatalf("expected %v, but found %v", testConfig.CAKeyFile, p.CAKeyFile)
	}

	if p.CACertFile != testConfig.CACertFile {
		t.Fatalf("expected %v, but found %v", testConfig.CACertFile, p.CACertFile)
	}

	if p.ListenAddr != testConfig.ListenAddr {
		t.Fatalf("expected %v, but found %v", testConfig.ListenAddr, p.ListenAddr)
	}

	if !reflect.DeepEqual(p.HTTPPorts, testConfig.HTTPPorts) {
		t.Fatalf("expected %v, but found %v", testConfig.HTTPPorts, p.HTTPPorts)
	}

	if !reflect.DeepEqual(p.HTTPSPorts, testConfig.HTTPSPorts) {
		t.Fatalf("expected %v, but found %v", testConfig.HTTPSPorts, p.HTTPSPorts)
	}

	if p.ForwardDNSServer != testConfig.ForwardDNSServer {
		t.Fatalf("expected %v, but found %v", testConfig.ForwardDNSServer, p.ForwardDNSServer)
	}

	if p.DNSPort != testConfig.DNSPort {
		t.Fatalf("expected %v, but found %v", testConfig.DNSPort, p.DNSPort)
	}

	if p.DNSRegex != testConfig.DNSRegex {
		t.Fatalf("expected %v, but found %v", testConfig.DNSRegex, p.DNSRegex)
	}

	if p.DNSResolverOverride != testConfig.DNSResolverOverride {
		t.Fatalf("expected %v, but found %v", testConfig.DNSResolverOverride, p.DNSResolverOverride)
	}
}

func TestConfig_Log(t *testing.T) {

	t.Parallel()

	l := &log.Config{}
	err := json.Unmarshal(testConfigBytes, l)
	if err != nil {
		t.Fatalf("error unmarshalint testConfigBytes %s", err.Error())
	}

	if l.Level != testConfig.Level {
		t.Fatalf("expected %v, but found %v", testConfig.Level, l.Level)
	}

	if l.Format != testConfig.Format {
		t.Fatalf("expected %v, but found %v", testConfig.Format, l.Format)
	}

	if l.RequestLogFile != testConfig.RequestLogFile {
		t.Fatalf("expected %v, but found %v", testConfig.RequestLogFile, l.RequestLogFile)
	}

	if l.WebHookURL != testConfig.WebHookURL {
		t.Fatalf("expected %v, but found %v", testConfig.WebHookURL, l.WebHookURL)
	}
}

var testConfig = &Config{
	LogResponses:        true,
	CAKeyFile:           "/path/to/file.key",
	CACertFile:          "/path/to/file.crt",
	ListenAddr:          "10.10.10.10",
	HTTPPorts:           []int{80, 8080},
	HTTPSPorts:          []int{443, 4443},
	ForwardDNSServer:    "8.8.8.8",
	DNSPort:             53,
	DNSRegex:            ".*example.com",
	DNSResolverOverride: "8.8.8.8",

	Level:          log.WARNING,
	Format:         log.JSON,
	RequestLogFile: "/path/to/log.json",
	WebHookURL:     "http://www.webhook.url/path",
}
