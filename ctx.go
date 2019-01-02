package twig

import (
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/twiglab/twig/internal/json"
)

// Ctx 接口，用于向Handler传递Twig上下文数据，并提供简化操作完成请求处理
type Ctx interface {
	Req() *http.Request
	Resp() *ResponseWarp

	SetReq(*http.Request)

	IsTls() bool
	IsWebSocket() bool

	Scheme() string

	RealIP() string
	Path() string

	Param(string) string

	QueryParam(string) string
	QueryParams() url.Values
	QueryString() string

	FormValue(string) string
	FormParams() (url.Values, error)

	Get(string) interface{}
	Set(string, interface{})

	JSON(int, interface{}) error
	JSONBlob(int, []byte) error
	JSONP(int, string, interface{}) error
	JSONPBlob(int, string, []byte) error

	HTMLBlob(int, []byte) error
	HTML(int, string) error

	XML(int, interface{}) error
	XMLBlob(int, []byte) error

	Blob(int, string, []byte) error
	Stream(int, string, io.Reader) error

	String(int, string) error
	Stringf(int, string, ...interface{}) error

	URL(string, ...interface{}) string

	Cookie(string) (*http.Cookie, error)
	SetCookie(*http.Cookie)
	Cookies() []*http.Cookie

	NoContent(int) error
	Error(error)
	Redirect(int, string) error

	Twig() *Twig

	Logger() Logger
}

// MCtx 接口用于Twig内部管理，Twig受到请求后，通过MCtx和Muxer交互
type MCtx interface {
	Twig() *Twig
	Logger() Logger

	SetHandler(HandlerFunc)
	Handler() HandlerFunc

	SetPath(string)
	Reset(http.ResponseWriter, *http.Request)

	SetParamNames([]string)
	SetParamValues([]string)
	ParamNames() []string
	ParamValues() []string

	SetRoutes(map[string]Route)
}

type ctx struct {
	req  *http.Request
	resp *ResponseWarp

	path string

	pnames  []string
	pvalues []string

	query   url.Values
	handler HandlerFunc

	t *Twig

	store M

	routes map[string]Route
}

func (c *ctx) Twig() *Twig {
	return c.t
}

func (c *ctx) writeContentType(value string) {
	header := c.Resp().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *ctx) Resp() *ResponseWarp {
	return c.resp
}

func (c *ctx) Req() *http.Request {
	return c.req
}

func (c *ctx) SetReq(r *http.Request) {
	c.req = r
}

func (c *ctx) IsTls() bool {
	return c.req.TLS != nil
}

func (c *ctx) IsWebSocket() bool {
	upgrade := c.req.Header.Get(HeaderUpgrade)
	return upgrade == "websocket" || upgrade == "Websocket"
}

func (c *ctx) Scheme() string {
	if c.IsTls() {
		return "https"
	}
	if scheme := c.req.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := c.req.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := c.req.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := c.req.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func (c *ctx) RealIP() string {
	if ip := c.req.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := c.req.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.req.RemoteAddr)
	return ra
}

func (c *ctx) Path() string {
	return c.path
}

func (c *ctx) SetPath(p string) {
	c.path = p
}

func (c *ctx) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}
		}
	}
	return ""
}

func (c *ctx) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query.Get(name)
}

func (c *ctx) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}

	return c.query
}

func (c *ctx) QueryString() string {
	return c.req.URL.RawQuery
}

func (c *ctx) FormValue(name string) string {
	return c.req.FormValue(name)
}

func (c *ctx) FormParams() (url.Values, error) {
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

/*
func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

func (c *context) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(defaultMemory)
	return c.request.MultipartForm, err
}
*/

func (c *ctx) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

func (c *ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp(), cookie)
}

func (c *ctx) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

func (c *ctx) JSON(code int, val interface{}) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return c.JSONBlob(code, bs)
}

func (c *ctx) JSONBlob(code int, bs []byte) error {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, bs)
}

func (c *ctx) JSONP(code int, callback string, val interface{}) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.JSONPBlob(code, callback, bs)
}

func (c *ctx) JSONPBlob(code int, callback string, b []byte) (err error) {
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

func (c *ctx) Blob(code int, contentType string, bs []byte) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = c.resp.Write(bs)
	return
}

func (c *ctx) HTMLBlob(code int, bs []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, bs)
}

func (c *ctx) HTML(code int, html string) error {
	return c.HTMLBlob(code, []byte(html))
}

func (c *ctx) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if err != nil {
		return
	}
	return c.XMLBlob(code, b)
}

func (c *ctx) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.resp.Write(b)
	return
}

func (c *ctx) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = io.Copy(c.resp, r)
	return
}

func (c *ctx) String(code int, str string) error {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(str))
}

func (c *ctx) Stringf(code int, format string, v ...interface{}) error {
	return c.String(code, fmt.Sprintf(format, v...))
}

func (c *ctx) Get(key string) interface{} {
	return c.store[key]
}

func (c *ctx) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(M)
	}
	c.store[key] = val
}

func (c *ctx) NoContent(code int) error {
	c.resp.WriteHeader(code)
	return nil
}

func (c *ctx) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.resp.Header().Set(HeaderLocation, url)
	c.resp.WriteHeader(code)
	return nil
}

func (c *ctx) Error(err error) {
	c.t.HttpErrorHandler(err, c)
}

func (c *ctx) Handler() HandlerFunc {
	return c.handler
}

func (c *ctx) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *ctx) Logger() Logger {
	return c.t.Logger
}

func (c *ctx) SetParamNames(n []string) {
	c.pnames = n
}

func (c *ctx) SetParamValues(v []string) {
	c.pvalues = v
}

func (c *ctx) ParamNames() []string {
	return c.pnames
}

func (c *ctx) ParamValues() []string {
	return c.pvalues
}

func (c *ctx) Reset(w http.ResponseWriter, r *http.Request) {
	c.req = r
	c.resp.reset(w)
	c.query = nil
	c.handler = NotFoundHandler
	c.store = nil
	c.path = ""
	c.pnames = nil

	c.routes = nil
}

func (c *ctx) SetRoutes(rs map[string]Route) {
	c.routes = rs
}

func (c *ctx) URL(name string, params ...interface{}) string {
	for _, route := range c.routes {
		if route.Name() == name {
			return Reverse(route.Path(), params...)
		}
	}
	return ""
}
