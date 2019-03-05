// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import "strings"

type Format uint8

const (
	JSON Format = 1 << iota
	TEXT
)

func (l Format) String() string {

	switch l {
	case JSON:
		return "json"
	default:
		return "text"
	}
}

func (l *Format) Parse(level string) {

	level = strings.ToLower(level)

	switch level {
	case "json":
		*l = JSON
	default:
		*l = TEXT
	}
}
