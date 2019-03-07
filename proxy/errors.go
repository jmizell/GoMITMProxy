// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import "fmt"

const ERRCastErrorFailed = ErrorStr("cast to type proxy error failed")

type ErrorStr string

func (e ErrorStr) Err() *Error {
	return NewError(string(e))
}

type Error struct {
	errType string
	error   error
}

func (e *Error) Error() string {

	if e.error == nil {
		return e.errType
	}

	return fmt.Sprintf("%s: %s", e.errType, e.error.Error())
}

func (e *Error) WithReason(format string, a ...interface{}) *Error {

	e.error = fmt.Errorf(format, a...)
	return e
}

func (e *Error) WithError(err error) *Error {

	if err != nil {
		e.error = err
	}

	return e
}

func (e *Error) GetError() error {

	return e.error
}

func (e *Error) IsProxyError(err error) bool {

	if _, ok := err.(*Error); ok {
		return true
	}

	return false
}

func (e *Error) MustUnmarshalError(err error) {

	if castErr := e.UnmarshalError(err); castErr != nil {
		panic("failed to cast error to proxy error")
	}
}

func (e *Error) UnmarshalError(err error) error {

	if newError, ok := err.(*Error); ok {
		e.errType = newError.errType
		e.error = newError.error
		return nil
	}

	return ERRCastErrorFailed.Err()
}

func (e *Error) Match(err *Error) bool {

	return e.errType == err.errType
}

func NewError(errType string) *Error {

	return &Error{errType: errType}
}
