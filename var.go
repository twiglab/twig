package twig

import (
	"fmt"
	"net/http"
)

/*
MaxParam URL中最大的参数，注意这个是全局生效的，
无论你有多少路由，请确保最大的参数个数必须小于MaxParam
*/
var MaxParam int = 3

var (
	/*
		NotFoundHandler 全局404处理方法， 如果需要修改
		twig.NotFoundHandler = func (c twig.C) {
				...
		}
	*/
	NotFoundHandler = func(c C) error {
		return ErrNotFound
	}

	/*
		MethodNotAllowedHandler 全局405处理方法
	*/
	MethodNotAllowedHandler = func(c C) error {
		return ErrMethodNotAllowed
	}
)

var DefaultHttpErrorHandler = func(err error, c C) {

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
			err = c.Json(code, msg)
		}
		if err != nil {
			c.Logger().Println(err)
		}
	}
}
