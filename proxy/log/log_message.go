// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"encoding/json"
	"fmt"
	"github.com/benburkert/dns"
	"net/http"
	"sort"
	"strings"
	"time"
)

type MSG struct {
	logger   Handler
	exitCode int

	Timestamp    time.Time              `json:"timestamp"`
	Message      string                 `json:"message"`
	Fields       map[string]interface{} `json:"fields,omitempty"`
	Request      *RequestRecord         `json:"request,omitempty"`
	Response     *ResponseRecord        `json:"response,omitempty"`
	DNS          *DNSRecord             `json:"dns,omitempty"`
	ErrorMessage string                 `json:"error,omitempty"`
	Level        Level                  `json:"level"`
}

func NewMSG(l Handler) *MSG {

	return &MSG{
		logger: l,
		Fields: map[string]interface{}{},
	}
}

func (l *MSG) WithExitCode(exitCode int) *MSG {

	l.exitCode = exitCode

	return l
}

func (l *MSG) WithError(err error) *MSG {

	l.ErrorMessage = fmt.Sprintf("%v", err)

	return l
}

func (l *MSG) WithField(key string, value interface{}) *MSG {

	l.Fields[key] = value

	return l
}

func (l *MSG) WithRequest(req *http.Request) *MSG {

	l.Request = &RequestRecord{}
	err := l.Request.Load(req)
	if err != nil {
		l.logger.WithError(err).Error("failed to log request")
	}

	return l
}

func (l *MSG) WithResponse(res *http.Response) *MSG {

	l.Response = &ResponseRecord{}
	err := l.Response.Load(res)
	if err != nil {
		l.logger.WithError(err).Error("failed to log response")
	}

	return l
}

func (l *MSG) WithDNSQuestions(questions []dns.Question) *MSG {

	if l.DNS == nil {
		l.DNS = &DNSRecord{}
	}

	l.DNS.AddQuestions(questions)

	return l
}

func (l *MSG) WithDNSAnswer(name string, ttl time.Duration, record dns.Record) *MSG {

	if l.DNS == nil {
		l.DNS = &DNSRecord{}
	}

	l.DNS.AddAnswer(name, ttl, record)

	return l
}

func (l *MSG) WithDNSNXDomain() *MSG {

	if l.DNS == nil {
		l.DNS = &DNSRecord{}
	}

	l.DNS.AddNXDomain()

	return l
}

func (l *MSG) Info(format string, a ...interface{}) {

	l.log(INFO, format, a...)
}

func (l *MSG) Debug(format string, a ...interface{}) {

	l.log(DEBUG, format, a...)
}

func (l *MSG) Fatal(format string, a ...interface{}) {

	l.log(FATAL, format, a...)
}

func (l *MSG) Warning(format string, a ...interface{}) {

	l.log(WARNING, format, a...)
}

func (l *MSG) Error(format string, a ...interface{}) {

	l.log(ERROR, format, a...)
}

func (l *MSG) JSON() []byte {

	msg, err := json.Marshal(l)
	if err != nil {
		l.WithError(err).Error("error marshaling log to json")
	}

	return msg
}

func (l *MSG) String() (msg string) {

	msg = fmt.Sprintf("%s %s:", l.Timestamp.Format(time.RFC3339), l.Level)

	if l.Request != nil {
		msg = fmt.Sprintf("%s [%s] %s", msg, l.Request.Method, l.Request.URL.String())
	} else if l.DNS != nil {
		msg = fmt.Sprintf("%s [DNS]", msg)
	}

	if l.Message != "" {
		msg = fmt.Sprintf("%s %s", msg, strings.Replace(l.Message, "\"", "\\\"", -1))
	}

	if l.ErrorMessage != "" {
		msg = fmt.Sprintf("%s err=\"%s\"", msg, strings.Replace(l.ErrorMessage, "\"", "\\\"", -1))
	}

	var keys []string
	for key := range l.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		msg = fmt.Sprintf("%s %s=\"%s\"",
			msg,
			strings.Replace(key, " ", "_", -1),
			strings.Replace(fmt.Sprintf("%v", l.Fields[key]), "\"", "\\\"", -1))
	}

	if l.DNS != nil && len(l.DNS.Questions) > 0 {

		var qSlice []string
		for _, q := range l.DNS.Questions {
			questionStr := fmt.Sprintf("{Name:%s,Type:%s,Class:%v}", q.Name, q.Type, q.Class)
			qSlice = append(qSlice, questionStr)
		}

		msg = fmt.Sprintf("%s questions=[%s]", msg, strings.Join(qSlice, ","))
	}

	if l.DNS != nil && len(l.DNS.Answers) > 0 {

		var rSlice []string
		for _, a := range l.DNS.Answers {
			var answerStr string

			if a.NXDomain == "" {
				data, _ := json.Marshal(a.Record)
				record := strings.Replace(string(data), `"`, "", -1)
				answerStr = fmt.Sprintf("{Name:%s,TTL:%s,Record:%v}", a.Name, a.TTL, record)
			} else {
				answerStr = "{NXDomain}"
			}

			rSlice = append(rSlice, answerStr)
		}

		msg = fmt.Sprintf("%s answers=[%s]", msg, strings.Join(rSlice, ","))
	}

	return msg
}

func (l *MSG) log(level Level, format string, a ...interface{}) {
	l.Timestamp = time.Now()
	l.Level = level
	l.Message = fmt.Sprintf(format, a...)
	l.logger.Write(l)
}
