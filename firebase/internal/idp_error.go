package internal

import (
	"net/http"
	"net/url"
)

type idpError struct {
	error
	code int
	Raw  string
}

func newIdpError(err error) error {
	return &url.Error{
		Err: idpError{
			error: err,
			code:  http.StatusBadRequest,
		},
	}
}

func newIdpErrorFromResponse(err error, code int, data string) error {
	return &url.Error{
		Err: idpError{
			error: err,
			Raw:   data,
			code:  code,
		},
	}
}

func (err idpError) Temporary() bool {
	switch err.code {
	case http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotImplemented, http.StatusBadRequest:
		return false
	default:
		return true
	}
}

func (err idpError) Timeout() bool {
	switch err.code {
	case http.StatusRequestTimeout:
		return true
	default:
		return false
	}
}
