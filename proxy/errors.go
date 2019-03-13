// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import "fmt"

// Error unable to cast go error to ProxyError
const ERRCastErrorFailed = ErrorStr("cast to type proxy error failed")

// ErrorStr is a string representation of ProxyError, to be used as a constant
type ErrorStr string

// Err returns the ProxyError value of the string
func (e ErrorStr) Err() *ProxyError {
	return NewError(string(e))
}

// ProxyError is the custom error type for the proxy package. A ProxyError contains a type string, and an error value
// describes the details of the specific error.
//
// A ProxyError returned as a normal error, can be easily cast back to a ProxyError with UnmarshalError.
//
// Types of ProxyErrors can be compared against each other using Match.
type ProxyError struct {
	errType string
	error   error
}

// ProxyError fulfills the golang error type
func (e *ProxyError) Error() string {

	if e.error == nil {
		return e.errType
	}

	return fmt.Sprintf("%s: %s", e.errType, e.error.Error())
}

// WithReason creates a new golang error and adds it to ProxyError using the format string provided
func (e *ProxyError) WithReason(format string, a ...interface{}) *ProxyError {

	e.error = fmt.Errorf(format, a...)
	return e
}

// WithError adds a golang error to ProxyError. This is used to nest errors, or chains of errors in the ProxyError type
func (e *ProxyError) WithError(err error) *ProxyError {

	if err != nil {
		e.error = err
	}

	return e
}

// GetError returns the golang error value stored in ProxyError
func (e *ProxyError) GetError() error {

	return e.error
}

// IsProxyError determines if the supplied error is a ProxyError
func (e *ProxyError) IsProxyError(err error) bool {

	if _, ok := err.(*ProxyError); ok {
		return true
	}

	return false
}

// MustUnmarshalError casts an error value to a ProxyError, or panics on failure
func (e *ProxyError) MustUnmarshalError(err error) {

	if castErr := e.UnmarshalError(err); castErr != nil {
		panic(ERRCastErrorFailed.Err().WithError(castErr).Error())
	}
}

// UnmarshalError casts an error value to a ProxyError
func (e *ProxyError) UnmarshalError(err error) error {

	if newError, ok := err.(*ProxyError); ok {
		e.errType = newError.errType
		e.error = newError.error
		return nil
	}

	return ERRCastErrorFailed.Err()
}

// Match determines if the ProxyError have the same error type
func (e *ProxyError) Match(err *ProxyError) bool {

	return e.errType == err.errType
}

// NewError creates a ProxyError of the specified type
func NewError(errType string) *ProxyError {

	return &ProxyError{errType: errType}
}
