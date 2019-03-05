// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

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
