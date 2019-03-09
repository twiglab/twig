package twig

import (
	"context"
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

type lead struct {
	twig  *Twig
	works []Server
}

func (l *lead) Attach(t *Twig) {
	l.twig = t
}

func (l *lead) Start() error {
	for _, s := range l.works {
		if err := s.Start(); err != nil {
			l.twig.Logger.Println(err)
		}
	}
	return nil
}

func (l *lead) Shutdown(ctx context.Context) error {
	for _, s := range l.works {
		if err := s.Shutdown(ctx); err != nil {
			l.twig.Logger.Println(err)
		}
	}

	return nil
}

func (l *lead) AddServer(servers ...Server) {
	l.works = append(l.works, servers...)
	for _, s := range l.works {
		s.Attach(l.twig)
	}
}
