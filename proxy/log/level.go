// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import "strings"

type Level uint8

const (
	FATAL Level = 1 << iota
	ERROR
	WARNING
	INFO
	DEBUG
)

func (l Level) String() string {

	switch l {
	case FATAL:
		return "FATAL"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	default:
		return "ERROR"
	}
}

func (l *Level) Parse(level string) {

	level = strings.ToUpper(level)

	switch level {
	case "FATAL":
		*l = FATAL
	case "DEBUG":
		*l = DEBUG
	case "INFO":
		*l = INFO
	case "WARNING":
		*l = WARNING
	default:
		*l = ERROR
	}
}
