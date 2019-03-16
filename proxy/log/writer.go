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
}

type JSONWriter struct{}

func (w *JSONWriter) Write(msg *MSG) error {

	fmt.Println(string(msg.JSON()))
	return nil
}

type TextWriter struct{}

func (w *TextWriter) Write(msg *MSG) error {

	fmt.Println(msg.String())
	return nil
}

type RequestWriter struct {
	lock sync.Mutex
	file *os.File

	RequestLogFile string `json:"request_log_file"`
}

func (w *RequestWriter) Write(msg *MSG) error {
	var err error

	if msg.Request != nil && w.RequestLogFile != "" {
		w.lock.Lock()
		defer w.lock.Unlock()

		if w.file == nil {
			w.file, err = os.OpenFile(w.RequestLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
		}

		_, err := w.file.Write(append(msg.JSON(), []byte("\n")...))
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *RequestWriter) Close() error {

	if w.file != nil {
		err := w.file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
