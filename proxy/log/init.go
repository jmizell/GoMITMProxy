// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import "net/http"

var DefaultLogger = NewHandler(INFO)

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
