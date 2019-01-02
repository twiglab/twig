package twig

import (
	"bufio"
	"net"
	"net/http"
)

// RespCallBack ResponseWriter回调函数
type RespCallBack func(http.ResponseWriter)

// 包装http.ResponseWrite
// 提供以下增强：
// 1, Hijack功能
// 2, 通过Committed防止输出先于header
type ResponseWarp struct {
	before    []RespCallBack
	after     []RespCallBack
	Writer    http.ResponseWriter
	Status    int
	Len       int64
	Committed bool
}

func NewResponseWarp(w http.ResponseWriter) *ResponseWarp {
	return &ResponseWarp{Writer: w}
}

func (r *ResponseWarp) Header() http.Header {
	return r.Writer.Header()
}

func (r *ResponseWarp) Flush() {
	r.Writer.(http.Flusher).Flush()
}

func (r *ResponseWarp) CloseNotify() <-chan bool {
	return r.Writer.(http.CloseNotifier).CloseNotify()
}

func (r *ResponseWarp) Before(fn RespCallBack) {
	r.before = append(r.before, fn)
}

func (r *ResponseWarp) After(fn RespCallBack) {
	r.after = append(r.after, fn)
}

// 设置header时查是否已经输出内容
func (r *ResponseWarp) WriteHeader(code int) {
	if r.Committed {
		return
	}
	for _, fn := range r.before {
		fn(r)
	}
	r.Status = code
	r.Writer.WriteHeader(code)
	r.Committed = true
}

// 输出时候检查是否设置Header
func (r *ResponseWarp) Write(b []byte) (n int, err error) {
	if !r.Committed {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.Writer.Write(b)
	r.Len += int64(n)
	for _, fn := range r.after {
		fn(r)
	}
	return
}

// Hijack Hijack 支持
func (r *ResponseWarp) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

func (r *ResponseWarp) reset(w http.ResponseWriter) {
	r.before = nil
	r.after = nil
	r.Writer = w
	r.Len = 0
	r.Status = http.StatusOK
	r.Committed = false
}
