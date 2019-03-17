// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

type RequestRecord struct {
	Method           string
	URL              *url.URL
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           map[string][]string
	Body             string
	ContentLength    int64
	TransferEncoding []string
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          map[string][]string
	RemoteAddr       string
	RequestURI       string
	TLS              bool
	TimeStamp        time.Time

	bodyBuffer *bytes.Buffer
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

	r.bodyBuffer = bytes.NewBuffer(make([]byte, 0))
	req.Body = ioutil.NopCloser(io.TeeReader(req.Body, r.bodyBuffer))

	return nil
}

func (r *RequestRecord) ReadBody() error {

	if r.bodyBuffer == nil {
		return nil
	}

	bodyBytes, err := ioutil.ReadAll(r.bodyBuffer)
	if err != nil {
		return err
	}
	r.Body = base64.StdEncoding.EncodeToString(bodyBytes)

	return nil
}

func (r *RequestRecord) MarshalJSON() ([]byte, error) {
	type RequestRecordAlias RequestRecord

	if err := r.ReadBody(); err != nil {
		return nil, err
	}

	return json.Marshal((*RequestRecordAlias)(r))
}
