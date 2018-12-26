package twig

import (
	"bufio"
	"net"
	"net/http"
)

type ResponseWarp struct {
	before    []func()
	after     []func()
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

func (r *ResponseWarp) Before(fn func()) {
	r.before = append(r.before, fn)
}

func (r *ResponseWarp) After(fn func()) {
	r.after = append(r.after, fn)
}

func (r *ResponseWarp) WriteHeader(code int) {
	if r.Committed {
		return
	}
	for _, fn := range r.before {
		fn()
	}
	r.Status = code
	r.Writer.WriteHeader(code)
	r.Committed = true
}

func (r *ResponseWarp) Write(b []byte) (n int, err error) {
	if !r.Committed {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.Writer.Write(b)
	r.Len += int64(n)
	for _, fn := range r.after {
		fn()
	}
	return
}

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
