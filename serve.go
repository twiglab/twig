package twig

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Server Http处理器接口
type Server interface {
	Attacher
	Cycler
}

type twigServer struct {
	*http.Server
	twig *Twig
	ln   net.Listener
}

func (s *twigServer) Attach(t *Twig) {
	s.twig = t
	s.Handler = t
}

func (s *twigServer) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

type httpServer struct {
	*twigServer
}

func (h *httpServer) Start() (err error) {
	go func() {
		if err = h.Server.Serve(h.ln); err != nil {
			h.twig.Logger.Panic(err)
		}
	}()
	return
}

func WrapListener(ln net.Listener) Server {
	return &httpServer{
		twigServer: &twigServer{
			Server: &http.Server{},
			ln:     ln,
		},
	}
}

func NewServer(addr string) Server {
	return WrapListener(NewKeepAliveListener(addr))
}

type tlsServer struct {
	*twigServer
	cert string
	key  string
}

func (h *tlsServer) Start() (err error) {
	go func() {
		if err = h.Server.ServeTLS(h.ln, h.cert, h.key); err != nil {
			h.twig.Logger.Panic(err)
		}
	}()
	return
}

func WrapListenerTLS(ln net.Listener) Server {
	return WrapListenerTLSCert(ln, "", "")
}

func WrapListenerTLSCert(ln net.Listener, cert, key string) Server {
	return &tlsServer{
		twigServer: &twigServer{
			Server: &http.Server{},
			ln:     ln,
		},
		cert: cert,
		key:  key,
	}
}

func WrapListenerTLSConfig(ln net.Listener, config *tls.Config) Server {
	return WrapListenerTLS(tls.NewListener(ln, config))
}

type KeepAliveListener struct {
	*net.TCPListener
}

func NewKeepAliveListener(address string) (l *KeepAliveListener) {

	var addr *net.TCPAddr
	var ln *net.TCPListener
	var err error

	if addr, err = net.ResolveTCPAddr("tcp", address); err != nil {
		panic(err)
	}

	if ln, err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}

	l = &KeepAliveListener{TCPListener: ln}

	return
}

func (l *KeepAliveListener) Accept() (net.Conn, error) {
	var conn *net.TCPConn
	var err error
	if conn, err = l.AcceptTCP(); err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(3 * time.Minute)

	return conn, err
}
