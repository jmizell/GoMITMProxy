// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import "testing"

func TestNewDefaultConfig(t *testing.T) {

	t.Parallel()

	c := NewDefaultConfig()

	if c.Format != TEXT {
		t.Fatalf("new config should have log_format text, but found %s", c.Format)
	}

	if c.Level != INFO {
		t.Fatalf("new config should have log_level info, but found %s", c.Level)
	}

	if c.RequestLogFile != "" {
		t.Fatalf("new config should have request_log_file=\"\", but found %s", c.RequestLogFile)
	}
}

func TestConfig_GetLogger(t *testing.T) {

	t.Parallel()

	t.Run("default", func(subTest *testing.T) {
		config := NewDefaultConfig()
		handler, requestWriter := config.GetLogger()

		if len(handler.Writers) != 1 {
			subTest.Fatalf("expected to have 1 writers, but found %d", len(handler.Writers))
		}
		if _, ok := handler.Writers[0].(*TextWriter); !ok {
			subTest.Fatalf("expected writer to be text writer")
		}

		if handler.Level != INFO {
			subTest.Fatalf("should have log_level info, but found %s", handler.Level)
		}

		if requestWriter != nil {
			subTest.Fatalf("default config should not return a response writer")
		}
	})

	t.Run("warning_level_json_formatter", func(subTest *testing.T) {
		config := NewDefaultConfig()
		config.Format = JSON
		config.Level = WARNING
		handler, requestWriter := config.GetLogger()

		if len(handler.Writers) != 1 {
			subTest.Fatalf("expected to have 1 writers, but found %d", len(handler.Writers))
		}
		if _, ok := handler.Writers[0].(*JSONWriter); !ok {
			subTest.Fatalf("expected writer to be json writer")
		}

		if handler.Level != WARNING {
			subTest.Fatalf("should have log_level warning, but found %s", handler.Level)
		}

		if requestWriter != nil {
			subTest.Fatalf("default config should not return a response writer")
		}
	})

	t.Run("request_writer", func(subTest *testing.T) {
		config := NewDefaultConfig()
		config.RequestLogFile = "file_path"
		handler, requestWriter := config.GetLogger()

		if len(handler.Writers) != 2 {
			subTest.Fatalf("expected to have 2 writers, but found %d", len(handler.Writers))
		}
		var found bool
		for _, w := range handler.Writers {
			if _, ok := w.(*TextWriter); ok {
				found = true
			}
		}
		if !found {
			subTest.Fatalf("expected writers to contain text writer")
		}

		if handler.Level != INFO {
			subTest.Fatalf("should have log_level info, but found %s", handler.Level)
		}

		if requestWriter == nil {
			subTest.Fatalf("expected response writer")
		}
	})
}
