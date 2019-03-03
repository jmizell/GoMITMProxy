package proxy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var Log = &LogHandler{
	Level:  LogINFO,
	Format: LOGTEXT,
}

type LogLevel uint8

const (
	LogFATAL LogLevel = 1 << iota
	LogERROR
	LogWARNING
	LogINFO
	LogDEBUG
)

func (l LogLevel) String() string {

	switch l {
	case LogFATAL:
		return "FATAL"
	case LogDEBUG:
		return "DEBUG"
	case LogINFO:
		return "INFO"
	case LogWARNING:
		return "WARNING"
	default:
		return "ERROR"
	}
}

func (l *LogLevel) Parse(level string) {

	level = strings.ToUpper(level)

	switch level {
	case "FATAL":
		*l = LogFATAL
	case "DEBUG":
		*l = LogDEBUG
	case "INFO":
		*l = LogINFO
	case "WARNING":
		*l = LogWARNING
	default:
		*l = LogERROR
	}
}

type LogFormat uint8

const (
	LogJSON LogFormat = 1 << iota
	LOGTEXT
)

func (l LogFormat) String() string {

	switch l {
	case LogJSON:
		return "json"
	default:
		return "text"
	}
}

func (l *LogFormat) Parse(level string) {

	level = strings.ToLower(level)

	switch level {
	case "json":
		*l = LogJSON
	default:
		*l = LOGTEXT
	}
}

type LogHandler struct {
	lock sync.Mutex
	file *os.File

	Level          LogLevel  `json:"log_level"`
	Format         LogFormat `json:"log_format"`
	RequestLogFile string    `json:"request_log_file"`
}

func (l *LogHandler) Write(msg *LogMSG) {

	if msg.Level <= l.Level {
		if l.Format == LogJSON {
			l.writeJSON(msg)
		} else {
			l.writeText(msg)
		}
	}

	l.writeRequestLog(msg)

	if msg.Level == LogFATAL {

		if msg.exitCode > 0 {
			os.Exit(msg.exitCode)
		}

		os.Exit(1)
	}
}

func (l *LogHandler) writeJSON(msg *LogMSG) {

	log, err := json.Marshal(msg)
	if err != nil {
		l.WithError(err).Error("error marshaling log to json")
	}
	fmt.Println(string(log))
}

func (l *LogHandler) writeText(msg *LogMSG) {

	fmt.Println(msg.Text())
}

func (l *LogHandler) writeRequestLog(msg *LogMSG) {
	var err error

	if msg.Request != nil && l.RequestLogFile != "" {
		l.lock.Lock()
		defer l.lock.Unlock()

		if l.file == nil {
			l.file, err = os.OpenFile(l.RequestLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				l.WithError(err).Fatal("error opening log file for writing")
			}
		}

		_, err := l.file.Write(msg.JSON())
		if err != nil {
			l.WithError(err).Fatal("error writing log message to file")
		}
	}
}

func (l *LogHandler) Close() {

	if l.file != nil {
		err := l.file.Close()
		if err != nil {
			l.WithError(err).Fatal("error closing log file")
		}
	}
}

func (l *LogHandler) NewMSG() *LogMSG {

	return &LogMSG{
		logger: l,
		Fields: map[string]interface{}{},
	}
}

func (l *LogHandler) WithExitCode(exitCode int) *LogMSG {

	return l.NewMSG().WithExitCode(exitCode)
}

func (l *LogHandler) WithError(err error) *LogMSG {

	return l.NewMSG().WithError(err)
}

func (l *LogHandler) WithField(key string, value interface{}) *LogMSG {

	return l.NewMSG().WithField(key, value)
}

func (l *LogHandler) WithRequest(req *http.Request) *LogMSG {

	return l.NewMSG().WithRequest(req)
}

func (l *LogHandler) Info(format string, a ...interface{}) {

	l.NewMSG().Info(format, a...)
}

func (l *LogHandler) Debug(format string, a ...interface{}) {

	l.NewMSG().Debug(format, a...)
}

func (l *LogHandler) Fatal(format string, a ...interface{}) {

	l.NewMSG().Fatal(format, a...)
}

func (l *LogHandler) Warning(format string, a ...interface{}) {

	l.NewMSG().Warning(format, a...)
}

func (l *LogHandler) Error(format string, a ...interface{}) {

	l.NewMSG().Error(format, a...)
}

type LogMSG struct {
	logger   *LogHandler
	exitCode int

	Timestamp    time.Time              `json:"timestamp"`
	Message      string                 `json:"message"`
	Fields       map[string]interface{} `json:"fields,omitempty"`
	Request      *RequestRecord         `json:"request,omitempty"`
	ErrorMessage string                 `json:"error,omitempty"`
	Level        LogLevel               `json:"level"`
}

func (l *LogMSG) WithExitCode(exitCode int) *LogMSG {

	l.exitCode = exitCode

	return l
}

func (l *LogMSG) WithError(err error) *LogMSG {

	l.ErrorMessage = fmt.Sprintf("%v", err)

	return l
}

func (l *LogMSG) WithField(key string, value interface{}) *LogMSG {

	l.Fields[key] = value

	return l
}

func (l *LogMSG) WithRequest(req *http.Request) *LogMSG {

	l.Request = &RequestRecord{}
	err := l.Request.Load(req)
	if err != nil {
		l.logger.WithError(err).Error("failed to log request")
	}

	return l
}

func (l *LogMSG) Info(format string, a ...interface{}) {

	l.log(LogINFO, format, a...)
}

func (l *LogMSG) Debug(format string, a ...interface{}) {

	l.log(LogDEBUG, format, a...)
}

func (l *LogMSG) Fatal(format string, a ...interface{}) {

	l.log(LogFATAL, format, a...)
}

func (l *LogMSG) Warning(format string, a ...interface{}) {

	l.log(LogWARNING, format, a...)
}

func (l *LogMSG) Error(format string, a ...interface{}) {

	l.log(LogERROR, format, a...)
}

func (l *LogMSG) JSON() []byte {

	msg, err := json.Marshal(l)
	if err != nil {
		l.WithError(err).Error("error marshaling log to json")
	}

	return msg
}

func (l *LogMSG) Text() (msg string) {

	msg = fmt.Sprintf("%s %s:", l.Timestamp.Format(time.RFC3339), l.Level)

	if l.Request != nil {
		msg = fmt.Sprintf("%s [%s] %s", msg, l.Request.Method, l.Request.URL.String())
	}

	if l.Message != "" {
		msg = fmt.Sprintf("%s %s", msg, strings.Replace(l.Message, "\"", "\\\"", -1))
	}

	if l.ErrorMessage != "" {
		msg = fmt.Sprintf("%s err=\"%s\"", msg, strings.Replace(l.ErrorMessage, "\"", "\\\"", -1))
	}

	for key, value := range l.Fields {
		msg = fmt.Sprintf("%s %s=\"%s\"",
			msg,
			strings.Replace(key, " ", "_", -1),
			strings.Replace(fmt.Sprintf("%v", value), "\"", "\\\"", -1))
	}

	return msg
}

func (l *LogMSG) log(level LogLevel, format string, a ...interface{}) {
	l.Timestamp = time.Now()
	l.Level = level
	l.Message = fmt.Sprintf(format, a...)
	l.logger.Write(l)
}

type RequestRecord struct {
	Method           string              `json:"method"`
	URL              *url.URL            `json:"url"`
	Proto            string              `json:"proto"`
	ProtoMajor       int                 `json:"proto_major"`
	ProtoMinor       int                 `json:"proto_minor"`
	Header           map[string][]string `json:"header"`
	Body             string              `json:"body,omitempty"`
	ContentLength    int64               `json:"content_length,omitempty"`
	TransferEncoding []string            `json:"transfer_encoding,omitempty"`
	Host             string              `json:"host"`
	Form             url.Values          `json:"form,omitempty"`
	PostForm         url.Values          `json:"post_form,omitempty"`
	MultipartForm    *multipart.Form     `json:"multipart_form,omitempty"`
	Trailer          map[string][]string `json:"trailer,omitempty"`
	RemoteAddr       string              `json:"remote_addr"`
	RequestURI       string              `json:"request_uri"`
	TLS              bool                `json:"tls"`
	TimeStamp        time.Time           `json:"time_stamp"`
}

func (r *RequestRecord) Load(req *http.Request) (err error) {

	r.TimeStamp = time.Now()
	r.Method = req.Method
	r.URL = req.URL
	r.Proto = req.Proto
	r.ProtoMajor = req.ProtoMajor
	r.ProtoMinor = req.ProtoMinor
	r.Header = req.Header
	r.ContentLength = req.ContentLength
	r.TransferEncoding = req.TransferEncoding
	r.Host = req.Host
	r.Method = req.Method
	r.Form = req.Form
	r.PostForm = req.PostForm
	r.MultipartForm = req.MultipartForm
	r.Trailer = req.Trailer
	r.RemoteAddr = req.RemoteAddr
	r.RequestURI = req.RequestURI
	r.TLS = req.TLS != nil

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("failed to read body, %s", err.Error())
	}
	r.Body = base64.StdEncoding.EncodeToString(bodyBytes)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return nil
}

func (r *RequestRecord) MarshalIndent() []byte {
	d, _ := json.MarshalIndent(r, "", "  ")
	return d
}
