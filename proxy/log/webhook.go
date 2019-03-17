// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type WebHookWriter struct {
	level      Level
	WebHookURL string `json:"webhook_url"`
}

func (w *WebHookWriter) Write(msg *MSG) error {

	if w.WebHookURL == "" {
		return nil
	}

	// Only log requests, responses, and dns
	if msg.Request == nil && msg.Response == nil && msg.DNS == nil {
		return nil
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	err := json.NewEncoder(buf).Encode(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(w.WebHookURL, "application/json", buf)
	if err != nil {
		return err
	}

	if !(resp.StatusCode > 199 && resp.StatusCode < 300) {
		return fmt.Errorf("webhook returned non-200 response, %d", resp.StatusCode)
	}

	return nil
}

func (w *WebHookWriter) SetLevel(level Level) {}
