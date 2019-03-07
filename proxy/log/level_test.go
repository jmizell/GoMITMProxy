// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"testing"
)

func TestLevel_Parse(t *testing.T) {

	t.Parallel()

	tests := map[string]Level{
		"FATAL":   FATAL,
		"":        FATAL,
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"WARNING": WARNING,
		"ERROR":   ERROR,
	}

	for levelString, levelObject := range tests {
		t.Run(levelString, func(subTest *testing.T) {

			var l Level
			l.Parse(levelString)

			if l != levelObject {
				subTest.Fatalf("%s != %s", levelString, l.String())
			}
		})
	}
}

func TestLevel_String(t *testing.T) {

	t.Parallel()

	tests := map[string]Level{
		"FATAL":   FATAL,
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"WARNING": WARNING,
		"ERROR":   ERROR,
	}

	for levelString, levelObject := range tests {
		t.Run(levelString, func(subTest *testing.T) {

			if levelString != levelObject.String() {
				subTest.Fatalf("%s != %s", levelString, levelObject.String())
			}
		})
	}
}

func TestLevel_UnmarshalJSON(t *testing.T) {

	t.Parallel()

	tests := map[string]Level{
		"\"FATAL\"":   FATAL,
		"\"\"":        FATAL,
		"\"DEBUG\"":   DEBUG,
		"\"INFO\"":    INFO,
		"\"WARNING\"": WARNING,
		"\"ERROR\"":   ERROR,
	}

	for levelString, levelObject := range tests {
		t.Run(levelString, func(subTest *testing.T) {

			var l Level
			err := l.UnmarshalJSON([]byte(levelString))
			if err != nil {
				subTest.Fatalf("UnmarshalJSON failed with %s", err.Error())
			}

			if l != levelObject {
				subTest.Fatalf("%s != %s", levelString, l.String())
			}
		})
	}
}

func TestLevel_MarshalJSON(t *testing.T) {

	t.Parallel()

	tests := map[string]Level{
		"\"FATAL\"":   FATAL,
		"\"DEBUG\"":   DEBUG,
		"\"INFO\"":    INFO,
		"\"WARNING\"": WARNING,
		"\"ERROR\"":   ERROR,
	}

	for levelString, levelObject := range tests {
		t.Run(levelString, func(subTest *testing.T) {

			l, err := levelObject.MarshalJSON()
			if err != nil {
				subTest.Fatalf("MarshalJSON failed with %s", err.Error())
			}

			if string(l) != levelString {
				subTest.Fatalf("%s != %s", levelString, string(l))
			}
		})
	}
}
