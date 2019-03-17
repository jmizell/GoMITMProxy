// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

type Config struct {
	Level          Level  `json:"log_level"`
	Format         Format `json:"log_format"`
	RequestLogFile string `json:"request_log_file"`
	WebHookURL     string `json:"webhook_url"`
}

func (c Config) GetLogger() (handler *DefaultHandler, requestWriter *RequestWriter) {

	handler = NewHandler(c.Level)

	if c.Format == JSON {
		handler.SetWriter(&JSONWriter{})
	} else if c.Format == TEXT {
		handler.SetWriter(&TextWriter{})
	}

	if c.RequestLogFile != "" {
		requestWriter = &RequestWriter{RequestLogFile: c.RequestLogFile}
		handler.AddWriter(requestWriter)
	}

	if c.WebHookURL != "" {
		handler.AddWriter(&WebHookWriter{WebHookURL: c.WebHookURL})
	}

	return handler, requestWriter
}

func NewDefaultConfig() *Config {
	return &Config{
		Level:          INFO,
		Format:         TEXT,
		RequestLogFile: "",
		WebHookURL:     "",
	}
}
