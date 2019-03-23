// Copyright 2019 The Jeremy Mizell. All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in the LICENSE file.

package proxy

import (
	"fmt"
	"testing"
)

func TestProxyError_Error(t *testing.T) {

	t.Parallel()

	t.Run("only_proxy_error", func(subTest *testing.T) {
		err := NewError("test_error")

		expectedError := "test_error"
		if err.Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})

	t.Run("with_reason", func(subTest *testing.T) {
		err := NewError("test_error").WithReason("test reason")

		expectedError := "test_error: test reason"
		if err.Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})

	t.Run("with_error", func(subTest *testing.T) {
		err := NewError("test_error").WithError(fmt.Errorf("test error"))

		expectedError := "test_error: test error"
		if err.Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})
}

func TestProxyError_GetError(t *testing.T) {

	t.Parallel()

	t.Run("only_proxy_error", func(subTest *testing.T) {
		err := NewError("test_error")

		if err.GetError() != nil {
			subTest.Fatalf("expected error=nil, but received error=\"%s\"", err.Error())
		}
	})

	t.Run("with_reason", func(subTest *testing.T) {
		err := NewError("test_error").WithReason("test reason")

		expectedError := "test reason"
		if err.GetError().Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})

	t.Run("with_error", func(subTest *testing.T) {
		err := NewError("test_error").WithError(fmt.Errorf("test error"))

		expectedError := "test error"
		if err.GetError().Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})
}

func TestProxyError_IsProxyError(t *testing.T) {

	t.Parallel()

	t.Run("only_proxy_error", func(subTest *testing.T) {
		err := error(NewError("test_error"))

		var emptyErr ProxyError
		if !emptyErr.IsProxyError(err) {
			t.Fatalf("expected error to be proxy error")
		}
	})

	t.Run("with_reason", func(subTest *testing.T) {
		err := error(NewError("test_error").WithReason("test reason"))

		var emptyErr ProxyError
		if !emptyErr.IsProxyError(err) {
			t.Fatalf("expected error to be proxy error")
		}
	})

	t.Run("with_error", func(subTest *testing.T) {
		err := error(NewError("test_error").WithError(fmt.Errorf("test error")))

		var emptyErr ProxyError
		if !emptyErr.IsProxyError(err) {
			t.Fatalf("expected error to be proxy error")
		}
	})


	t.Run("go_error", func(subTest *testing.T) {
		err := fmt.Errorf("not a proxy error")

		var emptyErr ProxyError
		if emptyErr.IsProxyError(err) {
			t.Fatalf("expected error to not be a proxy error")
		}
	})
}

func TestProxyError_UnmarshalError(t *testing.T) {

	t.Parallel()

	t.Run("only_proxy_error", func(subTest *testing.T) {
		err := error(NewError("test_error"))

		var emptyErr ProxyError
		if unmarshalErr := emptyErr.UnmarshalError(err); unmarshalErr != nil{
			t.Fatalf("expected error to be cast to ProxyError, %s", unmarshalErr)
		}

		expectedError := "test_error"
		if emptyErr.Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})

	t.Run("with_reason", func(subTest *testing.T) {
		err := error(NewError("test_error").WithReason("test reason"))

		var emptyErr ProxyError
		if unmarshalErr := emptyErr.UnmarshalError(err); unmarshalErr != nil{
			t.Fatalf("expected error to be cast to ProxyError, %s", unmarshalErr)
		}

		expectedError := "test_error: test reason"
		if emptyErr.Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})

	t.Run("with_error", func(subTest *testing.T) {
		err := error(NewError("test_error").WithError(fmt.Errorf("test error")))

		var emptyErr ProxyError
		if unmarshalErr := emptyErr.UnmarshalError(err); unmarshalErr != nil{
			t.Fatalf("expected error to be cast to ProxyError, %s", unmarshalErr)
		}

		expectedError := "test_error: test error"
		if emptyErr.Error() != expectedError {
			subTest.Fatalf("expected error=\"%s\", but received error=\"%s\"", expectedError, err.Error())
		}
	})

	t.Run("go_error", func(subTest *testing.T) {
		err := fmt.Errorf("not a proxy error")

		var emptyErr ProxyError
		if unmarshalErr := emptyErr.UnmarshalError(err); unmarshalErr == nil{
			t.Fatalf("go error should not unmarshal to ProxyError")
		}
	})
}

func TestProxyError_Match(t *testing.T) {

	t.Parallel()

	t.Run("match_only_proxy_error", func(subTest *testing.T) {
		err1 := NewError("test_error")
		err2 := NewError("test_error")

		if !err1.Match(err2) {
			t.Fatalf("err1 should match err2")
		}
	})

	t.Run("no_match_only_proxy_error", func(subTest *testing.T) {
		err1 := NewError("test_error1")
		err2 := NewError("test_error2")

		if err1.Match(err2) {
			t.Fatalf("err1 should not match err2")
		}
	})

	t.Run("match_with_reason", func(subTest *testing.T) {
		err1 := NewError("test_error").WithReason("test reason1")
		err2 := NewError("test_error").WithReason("test reason2")

		if !err1.Match(err2) {
			t.Fatalf("err1 should match err2")
		}
	})

	t.Run("match_with_reason", func(subTest *testing.T) {
		err1 := NewError("test_error1").WithReason("test reason1")
		err2 := NewError("test_error2").WithReason("test reason2")

		if err1.Match(err2) {
			t.Fatalf("err1 should not match err2")
		}
	})

	t.Run("match_with_error", func(subTest *testing.T) {
		err1 := NewError("test_error").WithError(fmt.Errorf("test error1"))
		err2 := NewError("test_error").WithError(fmt.Errorf("test error2"))

		if !err1.Match(err2) {
			t.Fatalf("err1 should match err2")
		}
	})

	t.Run("match_with_error", func(subTest *testing.T) {
		err1 := NewError("test_error1").WithError(fmt.Errorf("test error1"))
		err2 := NewError("test_error2").WithError(fmt.Errorf("test error2"))

		if err1.Match(err2) {
			t.Fatalf("err1 should not match err2")
		}
	})
}

func TestProxyError_MustUnmarshalError(t *testing.T) {

	t.Parallel()

	t.Run("panic", func(subtest *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected MustUnmarshalError to panic")
			}
		}()

		var emptyErr ProxyError
		emptyErr.MustUnmarshalError(fmt.Errorf("not a ProxyError"))
	})

	t.Run("no_panic", func(subtest *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("expected MustUnmarshalError not to panic")
			}
		}()

		var emptyErr ProxyError
		emptyErr.MustUnmarshalError(NewError("a ProxyError"))
	})
}