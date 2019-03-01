package twig

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// ResponseWrap 包装http.ResponseWrite
type ResponseWrap struct {
	Writer    http.ResponseWriter
	Status    int
	Committed bool
}

func newResponseWrap(w http.ResponseWriter) *ResponseWrap {
	return &ResponseWrap{Writer: w}
}

func (r *ResponseWrap) Header() http.Header {
	return r.Writer.Header()
}

func (r *ResponseWrap) Flush() {
	r.Writer.(http.Flusher).Flush()
}

// 设置header时查是否已经输出内容
func (r *ResponseWrap) WriteHeader(code int) {
	if r.Committed {
		return
	}
	r.Status = code
	r.Writer.WriteHeader(code)
	r.Committed = true
}

// 输出时候检查是否设置Header
func (r *ResponseWrap) Write(b []byte) (n int, err error) {
	if !r.Committed {
		r.WriteHeader(OK)
	}
	_, err = r.Writer.Write(b)
	return
}

func (r *ResponseWrap) ReadFrom(src io.Reader) (n int64, e error) {
	if !r.Committed {
		r.WriteHeader(OK)
	}

	_, e = io.Copy(r.Writer, src)
	return
}

// Hijack Hijack 支持
func (r *ResponseWrap) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

func (r *ResponseWrap) reset(w http.ResponseWriter) {
	r.Writer = w
	r.Status = OK
	r.Committed = false
}
