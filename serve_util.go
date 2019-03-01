package twig

import (
	"crypto/tls"
	"net"
	"net/http"
)

func NewServerListener(ln net.Listener) Server {
	return &httpServer{
		twigServer: &twigServer{
			Server: &http.Server{},
			ln:     ln,
		},
	}
}

func NewServer(addr string) Server {
	return NewServerListener(NewKeepAliveListener(addr))
}

func NewServerTLS(addr, cert, key string) Server {
	ln := NewKeepAliveListener(addr)
	return NewServerListenerTLS(ln, cert, key)
}

func NewServerListenerTLS(ln net.Listener, cert, key string) Server {
	return &tlsServer{
		twigServer: &twigServer{
			Server: &http.Server{},
			ln:     ln,
		},
		cert: cert,
		key:  key,
	}
}

func NewServerConfigTLS(ln net.Listener, config *tls.Config) Server {
	return NewServerListenerTLS(tls.NewListener(ln, config), "", "")
}
