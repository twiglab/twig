package twig

import (
	"net"
	"net/http"
	"reflect"
	"strings"
	"unsafe"
)

func WriteContentType(w http.ResponseWriter, val string) {
	header := w.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, val)
	}
}

func WriteHeaderCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func IsTLS(r *http.Request) bool {
	return r.TLS != nil
}

func Byte(w http.ResponseWriter, code int, contentType string, bs []byte) (err error) {
	WriteContentType(w, contentType)
	w.WriteHeader(code)
	_, err = w.Write(bs)
	return
}

/*
func String(w http.ResponseWriter, code int, str string) (err error) {
	WriteContentType(w, MIMETextPlainCharsetUTF8)
	WriteHeaderCode(w, code)
	_, err = io.WriteString(w, str)
	return
}
*/

func UnsafeToBytes(s string) []byte {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}))
}

func String(w http.ResponseWriter, code int, str string) (err error) {
	WriteContentType(w, MIMETextPlainCharsetUTF8)
	WriteHeaderCode(w, code)
	_, err = w.Write(UnsafeToBytes(str))
	return
}

func IsWebSocket(r *http.Request) bool {
	upgrade := r.Header.Get(HeaderUpgrade)
	return upgrade == "websocket" || upgrade == "Websocket"
}

func IsXMLHTTPRequest(r *http.Request) bool {
	return strings.Contains(
		r.Header.Get(HeaderXRequestedWith),
		XMLHttpRequest,
	)
}

func Scheme(r *http.Request) string {
	if IsTLS(r) {
		return "https"
	}
	if scheme := r.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := r.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := r.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := r.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func RealIP(r *http.Request) string {
	if ip := r.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := r.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}
