// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type ResponseRecord struct {
	Status           string
	StatusCode       int
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	Body             string
	ContentLength    int64
	TransferEncoding []string
	Uncompressed     bool
	Trailer          http.Header
	TLS              bool
	TimeStamp        time.Time

	bodyBuffer *bytes.Buffer
}

func (r *ResponseRecord) Load(res *http.Response) (err error) {

	r.TimeStamp = time.Now()
	r.Status = res.Status
	r.StatusCode = res.StatusCode
	r.Proto = res.Proto
	r.ProtoMajor = res.ProtoMajor
	r.ProtoMinor = res.ProtoMinor
	r.Header = res.Header
	r.ContentLength = res.ContentLength
	r.TransferEncoding = res.TransferEncoding
	r.Uncompressed = res.Uncompressed
	r.Trailer = res.Trailer
	r.TLS = res.TLS != nil

	r.bodyBuffer = bytes.NewBuffer(make([]byte, 0))
	res.Body = ioutil.NopCloser(io.TeeReader(res.Body, r.bodyBuffer))

	return nil
}

func (r *ResponseRecord) ReadBody() error {

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

func (r *ResponseRecord) MarshalJSON() ([]byte, error) {
	type RequestRecordAlias ResponseRecord

	if err := r.ReadBody(); err != nil {
		return nil, err
	}

	return json.Marshal((*RequestRecordAlias)(r))
}
