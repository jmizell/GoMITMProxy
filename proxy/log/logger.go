// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"net/http"
	"os"
)

type Handler struct {
	Writers []Writer `json:"-"`
	Level   Level    `json:"log_level"`
}

func NewHandler(level Level) *Handler {
	return &Handler{Level: level}
}

func (l *Handler) SetWriter(w Writer) {
	l.Writers = []Writer{w}
}

func (l *Handler) AddWriter(w Writer) {

	if l.Writers == nil {
		l.Writers = []Writer{}
	}

	l.Writers = append(l.Writers, w)
}

func (l *Handler) Write(msg *MSG) {

	if l.Writers == nil {
		l.Writers = []Writer{&TextWriter{}}
	}

	if msg.Level <= l.Level {
		for _, w := range l.Writers {
			err := w.Write(msg)
			if err != nil {
				fmt.Printf("error writing message to log writer, %s\n", err.Error())
				os.Exit(1)
			}
		}
	}

	if msg.Level == FATAL {

		if msg.exitCode > 0 {
			os.Exit(msg.exitCode)
		}

		os.Exit(1)
	}
}

func (l *Handler) NewMSG() *MSG {

	return &MSG{
		logger: l,
		Fields: map[string]interface{}{},
	}
}

func (l *Handler) WithExitCode(exitCode int) *MSG {

	return l.NewMSG().WithExitCode(exitCode)
}

func (l *Handler) WithError(err error) *MSG {

	return l.NewMSG().WithError(err)
}

func (l *Handler) WithField(key string, value interface{}) *MSG {

	return l.NewMSG().WithField(key, value)
}

func (l *Handler) WithRequest(req *http.Request) *MSG {

	return l.NewMSG().WithRequest(req)
}

func (l *Handler) Info(format string, a ...interface{}) {

	l.NewMSG().Info(format, a...)
}

func (l *Handler) Debug(format string, a ...interface{}) {

	l.NewMSG().Debug(format, a...)
}

func (l *Handler) Fatal(format string, a ...interface{}) {

	l.NewMSG().Fatal(format, a...)
}

func (l *Handler) Warning(format string, a ...interface{}) {

	l.NewMSG().Warning(format, a...)
}

func (l *Handler) Error(format string, a ...interface{}) {

	l.NewMSG().Error(format, a...)
}