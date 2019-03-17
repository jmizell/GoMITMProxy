// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestWebHookWriter_Write(t *testing.T) {

	t.Parallel()

	receivedMessages := make([]*MSG, 0)
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		if request.Method != http.MethodPost {
			t.Logf("incoming request should be %s, but received %s", http.MethodPost, request.Method)
			t.Fail()
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var msg MSG
		err := json.NewDecoder(request.Body).Decode(&msg)
		if err != nil {
			t.Logf("failed to decode webhook payload to message, %s", err.Error())
			t.Fail()
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		receivedMessages = append(receivedMessages, &msg)
		writer.WriteHeader(http.StatusOK)
	})

	connection, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("failed to open a port of for the test server, %s", err.Error())
	}
	port := connection.Addr().(*net.TCPAddr).Port

	go server.Serve(connection)
	defer server.Shutdown(context.Background())
	t.Logf("server listen address 127.0.0.1:%d\n", port)

	webhookWriter := &WebHookWriter{
		WebHookURL: fmt.Sprintf("http://127.0.0.1:%d", port),
	}

	newMsg := func(i int) *MSG {
		return &MSG{
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("%d", i),
			Fields:    map[string]interface{}{"id": i},
			Level:     INFO,
		}
	}

	for i := 0; i < 10; i++ {

		msg := newMsg(i)
		err = webhookWriter.Write(msg)
		if err != nil {
			t.Fatalf("error writing to webhook, %s", err.Error())
		}

		msg = newMsg(i)
		msg.Level = DEBUG
		err = webhookWriter.Write(msg)
		if err != nil {
			t.Fatalf("error writing to webhook, %s", err.Error())
		}

		msg = newMsg(i)
		msg.Request = &RequestRecord{Method: "TEST", Host: fmt.Sprintf("TESTHOST%d", i)}
		err = webhookWriter.Write(msg)
		if err != nil {
			t.Fatalf("error writing to webhook, %s", err.Error())
		}

		msg = newMsg(i)
		msg.DNS = &DNSRecord{Answers: []*DNSAnswer{{Name: fmt.Sprintf("DOMAIN%d", i)}}}
		err = webhookWriter.Write(msg)
		if err != nil {
			t.Fatalf("error writing to webhook, %s", err.Error())
		}

		msg = newMsg(i)
		msg.Response = &ResponseRecord{Body: fmt.Sprintf("TESTBODY%d", i)}
		err = webhookWriter.Write(msg)
		if err != nil {
			t.Fatalf("error writing to webhook, %s", err.Error())
		}
	}
	time.Sleep(time.Second * 3)

	if len(receivedMessages) != 30 {
		t.Fatalf("expected to have received 30 messages, but received %d", len(receivedMessages))
	}
}
