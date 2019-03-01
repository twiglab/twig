package twig

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type radixTreeCtx struct {
	req   *http.Request
	resp  *ResponseWrap
	query url.Values
	store M
	twig  *Twig

	handler HandlerFunc
	path    string

	pnames  []string
	pvalues []string

	tree *RadixTree
}

func newRadixTreeCtx(tree *RadixTree) *radixTreeCtx {
	c := &radixTreeCtx{
		pvalues: make([]string, tree.maxParam),
		tree:    tree,
		handler: NotFoundHandler,
		resp:    newResponseWrap(nil),
	}

	return c
}

func (c *radixTreeCtx) Resp() *ResponseWrap {
	return c.resp
}

func (c *radixTreeCtx) Req() *http.Request {
	return c.req
}

func (c *radixTreeCtx) IsTls() bool {
	return IsTLS(c.req)
}

func (c *radixTreeCtx) IsWebSocket() bool {
	return IsWebSocket(c.req)
}

func (c *radixTreeCtx) IsXMLHttpRequest() bool {
	return IsXMLHTTPRequest(c.req)
}

func (c *radixTreeCtx) Scheme() string {
	return Scheme(c.req)
}

func (c *radixTreeCtx) RealIP() string {
	return RealIP(c.req)
}

func (c *radixTreeCtx) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query.Get(name)
}

func (c *radixTreeCtx) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}

	return c.query
}

func (c *radixTreeCtx) QueryString() string {
	return c.req.URL.RawQuery
}

func (c *radixTreeCtx) FormValue(name string) string {
	return c.req.FormValue(name)
}

func (c *radixTreeCtx) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.req.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.req.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.req.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.req.Form, nil
}

func (c *radixTreeCtx) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.req.FormFile(name)
	return fh, err
}

func (c *radixTreeCtx) MultipartForm() (*multipart.Form, error) {
	err := c.req.ParseMultipartForm(defaultMemory)
	return c.req.MultipartForm, err
}

func (c *radixTreeCtx) File(file string) error {
	http.ServeFile(c.resp, c.req, file)
	return nil
}

func (c *radixTreeCtx) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *radixTreeCtx) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *radixTreeCtx) contentDisposition(file, name, dispositionType string) error {
	c.Resp().Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *radixTreeCtx) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

func (c *radixTreeCtx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.resp, cookie)
}

func (c *radixTreeCtx) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

func (c *radixTreeCtx) JSON(code int, val interface{}) error {
	enc := json.NewEncoder(c.resp)
	WriteContentType(c.resp, MIMEApplicationJSONCharsetUTF8)
	WriteHeaderCode(c.resp, code)
	return enc.Encode(val)
}

func (c *radixTreeCtx) JSONP(code int, callback string, val interface{}) (err error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if c.twig.Debug {
		enc.SetIndent("", "\t")
	}
	WriteContentType(c.resp, MIMEApplicationJavaScriptCharsetUTF8)
	WriteHeaderCode(c.resp, code)

	if _, err = buf.WriteString(callback + "("); err != nil {
		return
	}

	if err = enc.Encode(val); err != nil {
		return
	}

	if _, err = buf.WriteString(");"); err != nil {
		return
	}

	_, err = buf.WriteTo(c.resp)
	return
}

func (c *radixTreeCtx) Blob(code int, contentType string, bs []byte) (err error) {
	return Byte(c.resp, code, contentType, bs)
}

func (c *radixTreeCtx) XML(code int, v interface{}) (err error) {
	enc := xml.NewEncoder(c.resp)
	WriteContentType(c.resp, MIMEApplicationXMLCharsetUTF8)
	WriteHeaderCode(c.resp, code)
	return enc.Encode(v)
}

func (c *radixTreeCtx) Stream(code int, contentType string, r io.Reader) (err error) {
	WriteContentType(c.resp, contentType)
	WriteHeaderCode(c.resp, code)
	_, err = c.resp.ReadFrom(r)
	return
}

func (c *radixTreeCtx) String(code int, str string) error {
	return String(c.resp, code, str)
}

func (c *radixTreeCtx) Stringf(code int, format string, v ...interface{}) error {
	return String(c.resp, code, fmt.Sprintf(format, v...))
}

func (c *radixTreeCtx) Get(key string) interface{} {
	return c.store[key]
}

func (c *radixTreeCtx) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(M)
	}
	c.store[key] = val
}

func (c *radixTreeCtx) NoContent() error {
	WriteHeaderCode(c.resp, http.StatusNoContent)
	return nil
}

// Redirect 重定向
func (c *radixTreeCtx) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.resp.Header().Set(HeaderLocation, url)
	c.resp.WriteHeader(code)
	return nil
}

// Twig 获取当前Twig
func (c *radixTreeCtx) Twig() *Twig {
	return c.twig
}

func (c *radixTreeCtx) Logger() Logger {
	return c.twig.Logger
}

func (c *radixTreeCtx) Error(e error) {
	c.twig.HttpErrorHandler(e, c)
}

func (c *radixTreeCtx) Reset(w http.ResponseWriter, r *http.Request, t *Twig) {
	c.req = r
	c.resp.reset(w)
	c.query = nil
	c.store = nil
	c.twig = t

}

func (c *radixTreeCtx) Release() {
	c.tree.releaseCtx(c)
}

func (c *radixTreeCtx) Path() string {
	return c.path
}

func (c *radixTreeCtx) Handler() HandlerFunc {
	return c.handler
}

func (c *radixTreeCtx) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}
		}
	}
	return ""
}
