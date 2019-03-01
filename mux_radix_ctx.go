package twig

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

/*
var radixTreeCtxPool sync.Pool

func init() {
	radixTreeCtxPool.New = newRadixTreeCtx()
}
*/

type radixTreeCtx struct {
	req   *http.Request
	resp  *ResponseWrap
	query url.Values
	store M
	twig  *Twig

	// -------------

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

func (c *radixTreeCtx) writeContentType(value string) {
	WriteContentType(c.resp, value)
}

func (c *radixTreeCtx) Resp() *ResponseWrap {
	return c.resp
}

func (c *radixTreeCtx) Req() *http.Request {
	return c.req
}

func (c *radixTreeCtx) SetReq(r *http.Request) {
	c.req = r
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
	if ip := c.req.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := c.req.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.req.RemoteAddr)
	return ra
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

const indexPage string = "index.html"

func (c *radixTreeCtx) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c.Resp(), c.Req(), fi.Name(), fi.ModTime(), f)
	return
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
	http.SetCookie(c.Resp(), cookie)
}

func (c *radixTreeCtx) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

func (c *radixTreeCtx) JSON(code int, val interface{}) error {
	enc := json.NewEncoder(c.resp)
	if c.twig.Debug {
		enc.SetIndent("", "\t")
	}

	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.resp.WriteHeader(code)
	return enc.Encode(val)
}

func (c *radixTreeCtx) JSONBlob(code int, bs []byte) error {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, bs)
}

func (c *radixTreeCtx) JSONP(code int, callback string, val interface{}) (err error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if c.twig.Debug {
		enc.SetIndent("", "\t")
	}
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.resp.WriteHeader(code)

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

func (c *radixTreeCtx) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = c.resp.Write(b); err != nil {
		return
	}
	_, err = c.resp.Write([]byte(");"))
	return
}

func (c *radixTreeCtx) Blob(code int, contentType string, bs []byte) (err error) {
	return Byte(c.resp, code, contentType, bs)
}

func (c *radixTreeCtx) HTMLBlob(code int, bs []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, bs)
}

func (c *radixTreeCtx) HTML(code int, html string) error {
	return c.HTMLBlob(code, []byte(html))
}

func (c *radixTreeCtx) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if err != nil {
		return
	}
	return c.XMLBlob(code, b)
}

func (c *radixTreeCtx) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.resp.Write(b)
	return
}

func (c *radixTreeCtx) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = c.resp.ReadFrom(r)
	return
}

func (c *radixTreeCtx) String(code int, str string) error {
	return UnsafeString(c.resp, code, str)
}

func (c *radixTreeCtx) Stringf(code int, format string, v ...interface{}) error {
	return UnsafeString(c.resp, code, fmt.Sprintf(format, v...))
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

func (c *radixTreeCtx) NoContent(code int) error {
	c.resp.WriteHeader(code)
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
