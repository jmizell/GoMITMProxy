// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package config

import (
	"encoding/json"
	"os"

	"github.com/jmizell/GoMITMProxy/proxy/log"
)

type Config struct {

	// MITMProxy Config
	LogResponses        bool   `json:"log_responses"`
	CAKeyFile           string `json:"ca_key_file"`
	CACertFile          string `json:"ca_cert_file"`
	ListenAddr          string `json:"listen_addr"`
	HTTPSPorts          []int  `json:"https_ports"`
	HTTPPorts           []int  `json:"http_ports"`
	ForwardDNSServer    string `json:"forward_dns_server"`
	DNSPort             int    `json:"dns_port"`
	DNSRegex            string `json:"dns_regex"`
	DNSResolverOverride string `json:"dns_resolver_override"`

	// Log Config
	Level          log.Level  `json:"log_level"`
	Format         log.Format `json:"log_format"`
	RequestLogFile string     `json:"request_log_file"`
	WebHookURL     string     `json:"webhook_url"`
}

func (c *Config) Write(filepath string) (err error) {

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return err
}
