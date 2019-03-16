// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"net/http"
	"os"
)

type Handler interface {
	SetWriter(Writer)
	AddWriter(Writer)
	Write(*MSG)
	WithExitCode(int) *MSG
	WithError(error) *MSG
	WithField(string, interface{}) *MSG
	WithRequest(*http.Request) *MSG
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Fatal(string, ...interface{})
	Warning(string, ...interface{})
	Error(string, ...interface{})
}

type DefaultHandler struct {
	Writers []Writer `json:"-"`
	Level   Level    `json:"log_level"`
}

func NewHandler(level Level) *DefaultHandler {
	return &DefaultHandler{Level: level}
}

func (l *DefaultHandler) SetWriter(w Writer) {
	l.Writers = []Writer{w}
}

func (l *DefaultHandler) AddWriter(w Writer) {

	if l.Writers == nil {
		l.Writers = []Writer{}
	}

	l.Writers = append(l.Writers, w)
}

func (l *DefaultHandler) Write(msg *MSG) {

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

func (l *DefaultHandler) NewMSG() *MSG {

	return NewMSG(l)
}

func (l *DefaultHandler) WithExitCode(exitCode int) *MSG {

	return l.NewMSG().WithExitCode(exitCode)
}

func (l *DefaultHandler) WithError(err error) *MSG {

	return l.NewMSG().WithError(err)
}

func (l *DefaultHandler) WithField(key string, value interface{}) *MSG {

	return l.NewMSG().WithField(key, value)
}

func (l *DefaultHandler) WithRequest(req *http.Request) *MSG {

	return l.NewMSG().WithRequest(req)
}

func (l *DefaultHandler) WithResponse(res *http.Response) *MSG {

	return l.NewMSG().WithResponse(res)
}

func (l *DefaultHandler) Info(format string, a ...interface{}) {

	l.NewMSG().Info(format, a...)
}

func (l *DefaultHandler) Debug(format string, a ...interface{}) {

	l.NewMSG().Debug(format, a...)
}

func (l *DefaultHandler) Fatal(format string, a ...interface{}) {

	l.NewMSG().Fatal(format, a...)
}

func (l *DefaultHandler) Warning(format string, a ...interface{}) {

	l.NewMSG().Warning(format, a...)
}

func (l *DefaultHandler) Error(format string, a ...interface{}) {

	l.NewMSG().Error(format, a...)
}
