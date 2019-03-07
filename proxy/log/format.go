// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"encoding/json"
	"strings"
)

type Format uint8

const (
	TEXT Format = 0
	JSON Format = 1 << iota
)

func (f *Format) Parse(level string) {

	level = strings.ToUpper(level)

	switch level {
	case "JSON":
		*f = JSON
	case "TEXT":
		*f = TEXT
	default:
		*f = TEXT
	}
}

func (f Format) String() string {

	switch f {
	case JSON:
		return "JSON"
	case TEXT:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (f *Format) UnmarshalJSON(data []byte) error {

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	f.Parse(s)

	return nil
}

func (f *Format) MarshalJSON() ([]byte, error) {

	return json.Marshal(f.String())
}
