// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"os"
	"sync"
)

type Writer interface {
	Write(*MSG) error
	SetLevel(Level)
}

type JSONWriter struct {
	level Level
}

func (w *JSONWriter) Write(msg *MSG) error {

	if msg.Level <= w.level {
		fmt.Println(string(msg.JSON()))
	}
	return nil
}

func (w *JSONWriter) SetLevel(level Level) {
	w.level = level
}

type TextWriter struct {
	level Level
}

func (w *TextWriter) Write(msg *MSG) error {

	if msg.Level <= w.level {
		fmt.Println(msg.String())
	}
	return nil
}

func (w *TextWriter) SetLevel(level Level) {
	w.level = level
}

type RequestWriter struct {
	lock sync.Mutex
	file *os.File

	RequestLogFile string `json:"request_log_file"`
}

func (w *RequestWriter) Write(msg *MSG) (err error) {

	if w.RequestLogFile == "" {
		return nil
	}

	// Only log requests, responses, and dns
	if msg.Request == nil && msg.Response == nil && msg.DNS == nil {
		return nil
	}

	w.lock.Lock()
	defer w.lock.Unlock()

	if w.file == nil {
		w.file, err = os.OpenFile(w.RequestLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
	}

	_, err = w.file.Write(append(msg.JSON(), []byte("\n")...))
	if err != nil {
		return err
	}
	return nil
}

func (w *RequestWriter) SetLevel(level Level) {}

func (w *RequestWriter) Close() error {

	if w.file != nil {
		err := w.file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
