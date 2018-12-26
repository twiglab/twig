package twig

import (
	"fmt"
	"net/http"
)

type HttpError struct {
	Code     int
	Msg      interface{}
	Internal error
}

func NewHttpError(code int, msg ...interface{}) *HttpError {
	e := &HttpError{
		Code: code,
		Msg:  http.StatusText(code),
	}

	if len(msg) > 0 {
		e.Msg = msg[0]
	}

	return e
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("code = %d, msg = %v", e.Code, e.Msg)
}

func (e *HttpError) SetInternal(err error) *HttpError {
	e.Internal = err
	return e
}
