// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReverseProxy_ServeHTTP(t *testing.T) {

	receivedMessages := make([]*http.Request, 0)
	receivedMessageBodies := make([][]byte, 0)
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("unexpected error in target server while reading reques")
		}
		receivedMessages = append(receivedMessages, request)
		receivedMessageBodies = append(receivedMessageBodies, body)

		writer.WriteHeader(http.StatusOK)
		_, err = writer.Write([]byte("OKAY"))
		if err != nil {
			t.Fatalf("unexpected error in target server while writing response")
		}
	})
	connection, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("failed to open a port of for the test server, %s", err.Error())
	}
	port := connection.Addr().(*net.TCPAddr).Port

	go server.Serve(connection)
	defer server.Shutdown(context.Background())
	t.Logf("server listen address 127.0.0.1:%d\n", port)

	rp := ReverseProxy{}

	t.Run("method_without_body", func(subTest *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/test_path", port), nil)
		resp := httptest.NewRecorder()
		rp.ServeHTTP(resp, req)

		if len(receivedMessages) != 1 {
			subTest.Fatalf("expected one message to be received by target server, but found %d", len(receivedMessages))
		}

		if receivedMessages[0].Method != http.MethodGet {
			subTest.Fatalf("expected target received method to be %s, but received %s", http.MethodGet, receivedMessages[0].Method)
		}

		if receivedMessages[0].URL.Path != "/test_path" {
			subTest.Fatalf("expected target received method to be /test_path, but received %s", receivedMessages[0].URL.Path)
		}

		if len(receivedMessageBodies) != 1 {
			subTest.Fatalf("expected one message to be received by target server, but found %d", len(receivedMessages))
		}
		if string(receivedMessageBodies[0]) != "" {
			subTest.Fatalf("expected received body to be empty, but received %s", string(receivedMessageBodies[0]))
		}

		if resp.Code != http.StatusOK {
			subTest.Fatalf("expected response status code %d, but received %d", http.StatusOK, resp.Code)
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			subTest.Fatalf("error reading from response body %s", err.Error())
		}
		if string(respBody) != "OKAY" {
			subTest.Fatalf("expected response body to be OKAY, but received %s", string(respBody))
		}
	})

	t.Run("method_with_body", func(subTest *testing.T) {
		reqBody := bytes.NewBufferString("test body")
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/test_path2", port), reqBody)
		resp := httptest.NewRecorder()
		rp.ServeHTTP(resp, req)

		if len(receivedMessages) != 2 {
			subTest.Fatalf("expected one message to be received by target server, but found %d", len(receivedMessages))
		}

		if receivedMessages[1].Method != http.MethodPost {
			subTest.Fatalf("expected target received method to be %s, but received %s", http.MethodGet, receivedMessages[1].Method)
		}

		if receivedMessages[1].URL.Path != "/test_path2" {
			subTest.Fatalf("expected target received method to be /test_path, but received %s", receivedMessages[1].URL.Path)
		}

		if len(receivedMessageBodies) != 2 {
			subTest.Fatalf("expected one message to be received by target server, but found %d", len(receivedMessages))
		}
		if string(receivedMessageBodies[1]) != "test body" {
			subTest.Fatalf("expected received body to be \"test body\", but received %s", string(receivedMessageBodies[1]))
		}

		if resp.Code != http.StatusOK {
			subTest.Fatalf("expected response status code %d, but received %d", http.StatusOK, resp.Code)
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			subTest.Fatalf("error reading from response body %s", err.Error())
		}
		if string(respBody) != "OKAY" {
			subTest.Fatalf("expected response body to be OKAY, but received %s", string(respBody))
		}
	})
}