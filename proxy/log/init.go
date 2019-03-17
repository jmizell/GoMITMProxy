// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"github.com/benburkert/dns"
	"net/http"
	"time"
)

var DefaultLogger, _ = NewDefaultConfig().GetLogger()

func WithExitCode(exitCode int) *MSG {

	return DefaultLogger.WithExitCode(exitCode)
}

func WithError(err error) *MSG {

	return DefaultLogger.WithError(err)
}

func WithField(key string, value interface{}) *MSG {

	return DefaultLogger.WithField(key, value)
}

func WithRequest(req *http.Request) *MSG {

	return DefaultLogger.WithRequest(req)
}

func WithResponse(res *http.Response) *MSG {

	return DefaultLogger.WithResponse(res)
}

func WithDNSQuestions(questions []dns.Question) *MSG {

	return DefaultLogger.WithDNSQuestions(questions)
}

func WithDNSAnswer(name string, ttl time.Duration, record dns.Record) *MSG {

	return DefaultLogger.WithDNSAnswer(name, ttl, record)
}

func WithDNSNXDomain() *MSG {

	return DefaultLogger.WithDNSNXDomain()
}

func Info(format string, a ...interface{}) {

	DefaultLogger.Info(format, a...)
}

func Debug(format string, a ...interface{}) {

	DefaultLogger.Debug(format, a...)
}

func Fatal(format string, a ...interface{}) {

	DefaultLogger.Fatal(format, a...)
}

func Warning(format string, a ...interface{}) {

	DefaultLogger.Warning(format, a...)
}

func Error(format string, a ...interface{}) {

	DefaultLogger.Error(format, a...)
}
