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

type HttpErrorHandler func(error, Ctx)

// 默认的错误处理
func DefaultHttpErrorHandler(err error, c Ctx) {
	var code int = http.StatusInternalServerError
	var msg interface{}

	if e, ok := err.(*HttpError); ok {
		code = e.Code
		msg = e.Msg

		if e.Internal != nil {
			err = fmt.Errorf("%v, %v", err, e.Internal)
		}
	} else {
		msg = http.StatusText(code)
	}

	if m, ok := msg.(string); ok {
		msg = map[string]string{"msg": m}
	}

	if !c.Resp().Committed {
		if c.Req().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			c.Logger().Println(err)
		}
	}
}
