package twig

import (
	"reflect"
	"unsafe"
)

func WriteContentType(c Ctx, value string) {
	header := c.Resp().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func IsTLS(c Ctx) bool {
	return c.Req().TLS != nil
}

func Byte(c Ctx, code int, contentType string, bs []byte) (err error) {
	WriteContentType(c, contentType)
	c.Resp().WriteHeader(code)
	_, err = c.Resp().Write(bs)
	return
}

func UnsafeToBytes(s string) []byte {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}))
}

func UnsafeString(c Ctx, code int, str string) error {
	return Byte(c, code, MIMETextPlainCharsetUTF8, UnsafeToBytes(str))
}
