// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"testing"
)

func TestFormat_Parse(t *testing.T) {

	t.Parallel()

	tests := map[string]Format{
		"TEXT": TEXT,
		"":     TEXT,
		"JSON": JSON,
	}

	for formatString, formatObject := range tests {
		t.Run(formatString, func(subTest *testing.T) {

			var l Format
			l.Parse(formatString)

			if l != formatObject {
				subTest.Fatalf("%s != %s", formatString, l.String())
			}
		})
	}
}

func TestFormat_String(t *testing.T) {

	t.Parallel()

	tests := map[string]Format{
		"TEXT": TEXT,
		"JSON": JSON,
	}

	for formatString, formatObject := range tests {
		t.Run(formatString, func(subTest *testing.T) {

			if formatString != formatObject.String() {
				subTest.Fatalf("%s != %s", formatString, formatObject.String())
			}
		})
	}
}

func TestFormat_UnmarshalJSON(t *testing.T) {

	t.Parallel()

	tests := map[string]Format{
		"\"TEXT\"": TEXT,
		"\"\"":     TEXT,
		"\"JSON\"": JSON,
	}

	for formatString, formatObject := range tests {
		t.Run(formatString, func(subTest *testing.T) {

			var l Format
			err := l.UnmarshalJSON([]byte(formatString))
			if err != nil {
				subTest.Fatalf("UnmarshalJSON failed with %s", err.Error())
			}

			if l != formatObject {
				subTest.Fatalf("%s != %s", formatString, l.String())
			}
		})
	}
}

func TestFormat_MarshalJSON(t *testing.T) {

	t.Parallel()

	tests := map[string]Format{
		"\"TEXT\"": TEXT,
		"\"JSON\"": JSON,
	}

	for formatString, formatObject := range tests {
		t.Run(formatString, func(subTest *testing.T) {

			l, err := formatObject.MarshalJSON()
			if err != nil {
				subTest.Fatalf("MarshalJSON failed with %s", err.Error())
			}

			if string(l) != formatString {
				subTest.Fatalf("%s != %s", formatString, string(l))
			}
		})
	}
}
