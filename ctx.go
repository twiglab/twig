package twig

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// Ctx 接口，用于向Handler传递Twig上下文数据，并提供简化操作完成请求处理
type Ctx interface {
	Req() *http.Request
	Resp() *ResponseWrap

	// IsTls 当前请求是否为Tls
	IsTls() bool
	//IsWebSocket 当前请求是否是WebSocket
	IsWebSocket() bool
	// IsXMLHttpRequest 当前请求是否为Ajax
	IsXMLHttpRequest() bool

	// Scheme 当前请求的的Scheme
	Scheme() string

	// RealIP 对方的IP
	RealIP() string
	// Path 当前请求的注册路径
	Path() string

	// Param 获取当前请求的URL参数
	Param(string) string

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

	// JSON JSON方式输出
	JSON(int, interface{}) error
	JSONP(int, string, interface{}) error

	XML(int, interface{}) error

	Blob(int, string, []byte) error
	Stream(int, string, io.Reader) error

	String(int, string) error
	Stringf(int, string, ...interface{}) error

	Cookie(string) (*http.Cookie, error)
	SetCookie(*http.Cookie)
	Cookies() []*http.Cookie

	NoContent() error
	Error(error)
	Redirect(int, string) error

	Twig() *Twig

	Logger() Logger
}

type MuxerCtx interface {
	Handler() HandlerFunc
	Release()
	Reset(http.ResponseWriter, *http.Request, *Twig)

	Ctx
}
