// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"encoding/json"
	"strings"
)

type Level uint8

const (
	FATAL Level = 0
	ERROR Level = 1 << iota
	WARNING
	INFO
	DEBUG
)

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
	case "ERROR":
		*l = ERROR
	default:
		*l = FATAL
	}
}

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
	case ERROR:
		return "ERROR"
	default:
		return "FATAL"
	}
}

func (l *Level) UnmarshalJSON(data []byte) error {

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	l.Parse(s)

	return nil
}

func (l *Level) MarshalJSON() ([]byte, error) {

	return json.Marshal(l.String())
}
