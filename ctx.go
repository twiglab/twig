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

type UrlParams map[string]string

// Ctx 接口，用于向Handler传递Twig上下文数据，并提供简化操作完成请求处理
type Ctx interface {
	Req() *http.Request
	Resp() *ResponseWrap

	SetReq(*http.Request)

	IsTls() bool
	IsWebSocket() bool
	IsXMLHttpRequest() bool

	Scheme() string
	Proto() (string, int, int)

	RealIP() string
	Path() string

	Param(string) string
	Params() UrlParams

	QueryParam(string) string
	QueryParams() url.Values
	QueryString() string

	FormValue(string) string
	FormParams() (url.Values, error)

	FormFile(name string) (*multipart.FileHeader, error)
	MultipartForm() (*multipart.Form, error)

	File(file string) error
	Attachment(file, name string) error
	Inline(file, name string) error

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

type PureCtx struct {
	req   *http.Request
	resp  *ResponseWrap
	query url.Values
	store M
	twig  *Twig

	fact Ctx
}

func NewPureCtx() *PureCtx {
	return &PureCtx{
		resp: newResponseWrap(nil),
	}
}

func (c *PureCtx) SetFact(fact Ctx) {
	c.fact = fact
}

func (c *PureCtx) reset(w http.ResponseWriter, r *http.Request, t *Twig) {
	c.req = r
	c.resp.reset(w)
	c.query = nil
	c.store = nil
	c.twig = t
}

func (c *PureCtx) writeContentType(value string) {
	header := c.Resp().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *PureCtx) Resp() *ResponseWrap {
	return c.resp
}

func (c *PureCtx) Req() *http.Request {
	return c.req
}

func (c *PureCtx) SetReq(r *http.Request) {
	c.req = r
}

func (c *PureCtx) IsTls() bool {
	return c.req.TLS != nil
}

func (c *PureCtx) IsWebSocket() bool {
	upgrade := c.req.Header.Get(HeaderUpgrade)
	return upgrade == "websocket" || upgrade == "Websocket"
}

func (c *PureCtx) IsXMLHttpRequest() bool {
	return strings.Contains(
		c.req.Header.Get(HeaderXRequestedWith),
		XMLHttpRequest,
	)
}

func (c *PureCtx) Scheme() string {
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

func (c *PureCtx) Proto() (string, int, int) {
	return c.req.Proto, c.req.ProtoMajor, c.req.ProtoMinor
}

func (c *PureCtx) RealIP() string {
	if ip := c.req.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := c.req.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.req.RemoteAddr)
	return ra
}

func (c *PureCtx) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query.Get(name)
}

func (c *PureCtx) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}

	return c.query
}

func (c *PureCtx) QueryString() string {
	return c.req.URL.RawQuery
}

func (c *PureCtx) FormValue(name string) string {
	return c.req.FormValue(name)
}

func (c *PureCtx) FormParams() (url.Values, error) {
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

func (c *PureCtx) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.req.FormFile(name)
	return fh, err
}

func (c *PureCtx) MultipartForm() (*multipart.Form, error) {
	err := c.req.ParseMultipartForm(defaultMemory)
	return c.req.MultipartForm, err
}

func (c *PureCtx) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(c.fact)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, IndexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(c.fact)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c.Resp(), c.Req(), fi.Name(), fi.ModTime(), f)
	return
}

func (c *PureCtx) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *PureCtx) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *PureCtx) contentDisposition(file, name, dispositionType string) error {
	c.Resp().Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *PureCtx) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

func (c *PureCtx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp(), cookie)
}

func (c *PureCtx) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

func (c *PureCtx) JSON(code int, val interface{}) error {
	enc := json.NewEncoder(c.resp)
	if c.twig.Debug {
		enc.SetIndent("", "\t")
	}

	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.resp.WriteHeader(code)
	return enc.Encode(val)
}

func (c *PureCtx) JSONBlob(code int, bs []byte) error {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, bs)
}

func (c *PureCtx) JSONP(code int, callback string, val interface{}) (err error) {
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

func (c *PureCtx) JSONPBlob(code int, callback string, b []byte) (err error) {
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

func (c *PureCtx) Blob(code int, contentType string, bs []byte) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = c.resp.Write(bs)
	return
}

func (c *PureCtx) HTMLBlob(code int, bs []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, bs)
}

func (c *PureCtx) HTML(code int, html string) error {
	return c.HTMLBlob(code, []byte(html))
}

func (c *PureCtx) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)
	if err != nil {
		return
	}
	return c.XMLBlob(code, b)
}

func (c *PureCtx) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.resp.Write(b)
	return
}

func (c *PureCtx) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = c.resp.ReadFrom(r)
	return
}

func (c *PureCtx) String(code int, str string) error {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(str))
}

func (c *PureCtx) Stringf(code int, format string, v ...interface{}) error {
	return c.String(code, fmt.Sprintf(format, v...))
}

func (c *PureCtx) Get(key string) interface{} {
	return c.store[key]
}

func (c *PureCtx) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(M)
	}
	c.store[key] = val
}

func (c *PureCtx) NoContent(code int) error {
	c.resp.WriteHeader(code)
	return nil
}

// Redirect 重定向
func (c *PureCtx) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.resp.Header().Set(HeaderLocation, url)
	c.resp.WriteHeader(code)
	return nil
}

// Twig 获取当前Twig
func (c *PureCtx) Twig() *Twig {
	return c.twig
}

func (c *PureCtx) Logger() Logger {
	return c.twig.Logger
}

func (c *PureCtx) Error(e error) {
	c.twig.HttpErrorHandler(e, c.fact)
}

type muxerCtx interface {
	Release()
	reset(http.ResponseWriter, *http.Request, *Twig)
	Handler() HandlerFunc
}
