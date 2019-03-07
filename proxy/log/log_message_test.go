// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMSG_WithRequest(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	testRequest := httptest.NewRequest("GET", "http://localhost", nil)
	msg.WithRequest(testRequest)

	if msg.Request.URL != testRequest.URL {
		t.Fatalf("expected url %s to match original request %s", msg.Request.URL, testRequest.URL)
	}

	if msg.Request.Method != testRequest.Method {
		t.Fatalf("expected method %s to match original request %s", msg.Request.Method, testRequest.Method)
	}

	if msg.Request.Proto != testRequest.Proto {
		t.Fatalf("expected proto %s to match original request %s", msg.Request.Proto, testRequest.Proto)
	}
}

func TestMSG_WithError(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	testError := fmt.Errorf("test_error")
	msg.WithError(testError)

	if msg.ErrorMessage != testError.Error() {
		t.Fatalf("expected error %s to match original value %s", msg.ErrorMessage, testError.Error())
	}
}

func TestMSG_WithExitCode(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	msg.WithExitCode(1010)

	if msg.exitCode != 1010 {
		t.Fatalf("expected exit code %d to match original value 1010", msg.exitCode)
	}
}

func TestMSG_WithField(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	msg.WithField("test_key", "test_value")

	if v, ok := msg.Fields["test_key"]; !ok {
		t.Fatalf("expected to find key test_key in fields")
	} else if v != "test_value" {
		t.Fatalf("expected field value %s to be test_value", v)
	}
}

func TestMSG_String(t *testing.T) {

	t.Parallel()

	testTime := time.Now()
	testHandler := &TestHandler{}
	msg := &MSG{
		logger:       testHandler,
		Message:      "test_message",
		Timestamp:    testTime,
		Level:        WARNING,
		ErrorMessage: "test_error",
		Fields:       map[string]interface{}{"test_key": "test_value"},
	}
	expectedMsg := fmt.Sprintf("%s WARNING: test_message err=\"test_error\" test_key=\"test_value\"", testTime.Format(time.RFC3339))

	if msg.String() != expectedMsg {
		t.Fatalf("expected message string \n%s\nbut resefound\n%s", expectedMsg, msg.String())
	}
}

func TestMSG_JSON(t *testing.T) {

	t.Parallel()

	testTime := time.Now()
	testHandler := &TestHandler{}
	msg := &MSG{
		logger:       testHandler,
		Message:      "test_message",
		Timestamp:    testTime,
		Level:        WARNING,
		ErrorMessage: "test_error",
		Fields:       map[string]interface{}{"test_key": "test_value"},
	}
	msgFormat := `{"timestamp":%s,"message":"test_message","fields":{"test_key":"test_value"},"error":"test_error","level":"WARNING"}`
	jsonTime, _ := testTime.MarshalJSON()
	expectedMsg := fmt.Sprintf(msgFormat, string(jsonTime))

	if string(msg.JSON()) != expectedMsg {
		t.Fatalf("expected message json \n%s\nbut found\n%s", expectedMsg, msg.JSON())
	}
}

func TestMSG_Debug(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	msg.Debug("test_message")

	if len(testHandler.messages) != 1 {
		t.Fatalf("expected 1 message to be sent, but found %d", len(testHandler.messages))
	}

	if testHandler.messages[0] != msg {
		t.Fatalf("expected set message to be %s, but found %s", msg, testHandler.messages[0])
	}
}

func TestMSG_Info(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	msg.Info("test_message")

	if len(testHandler.messages) != 1 {
		t.Fatalf("expected 1 message to be sent, but found %d", len(testHandler.messages))
	}

	if testHandler.messages[0] != msg {
		t.Fatalf("expected set message to be %s, but found %s", msg, testHandler.messages[0])
	}
}

func TestMSG_Warning(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	msg.Warning("test_message")

	if len(testHandler.messages) != 1 {
		t.Fatalf("expected 1 message to be sent, but found %d", len(testHandler.messages))
	}

	if testHandler.messages[0] != msg {
		t.Fatalf("expected set message to be %s, but found %s", msg, testHandler.messages[0])
	}
}

func TestMSG_Error(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	msg.Error("test_message")

	if len(testHandler.messages) != 1 {
		t.Fatalf("expected 1 message to be sent, but found %d", len(testHandler.messages))
	}

	if testHandler.messages[0] != msg {
		t.Fatalf("expected set message to be %s, but found %s", msg, testHandler.messages[0])
	}
}

func TestMSG_Fatal(t *testing.T) {

	t.Parallel()

	testHandler := &TestHandler{}
	msg := NewMSG(testHandler)
	msg.Fatal("test_message")

	if len(testHandler.messages) != 1 {
		t.Fatalf("expected 1 message to be sent, but found %d", len(testHandler.messages))
	}

	if testHandler.messages[0] != msg {
		t.Fatalf("expected set message to be %s, but found %s", msg, testHandler.messages[0])
	}
}

type TestHandler struct {
	messages []*MSG
}

func (t *TestHandler) SetWriter(Writer) {}

func (t *TestHandler) AddWriter(Writer) {}

func (t *TestHandler) Write(m *MSG) {
	t.messages = append(t.messages, m)
}

func (t *TestHandler) WithExitCode(int) *MSG {
	return nil
}

func (t *TestHandler) WithError(error) *MSG {
	return nil
}

func (t *TestHandler) WithField(string, interface{}) *MSG {
	return nil
}

func (t *TestHandler) WithRequest(*http.Request) *MSG {
	return nil
}

func (t *TestHandler) Info(string, ...interface{}) {}

func (t *TestHandler) Debug(string, ...interface{}) {}

func (t *TestHandler) Fatal(string, ...interface{}) {}

func (t *TestHandler) Warning(string, ...interface{}) {}

func (t *TestHandler) Error(string, ...interface{}) {}
