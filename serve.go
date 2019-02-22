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

type Work struct {
	*http.Server
	twig *Twig
}

func NewWork() *Work {
	return &Work{
		Server: &http.Server{
			Addr: DefaultAddress,
		},
	}
}

func (w *Work) Attach(t *Twig) {
	w.twig = t
	w.Handler = t
}

func (w *Work) Shutdown(ctx context.Context) error {
	return w.Server.Shutdown(ctx)
}

func (w *Work) Start() (err error) {
	go func() {
		if err = w.Server.ListenAndServe(); err != nil {
			w.twig.Logger.Panic(err)
		}
	}()
	return
}

// -------------------------------------- New

func MustListener(l net.Listener, err error) net.Listener {
	if err != nil {
		panic(err)
	}
	return l
}

type KeepAliveListener struct {
	*net.TCPListener
	count uint64
}

func NewKeepAliveListener(address string) (l *KeepAliveListener, err error) {
	var addr *net.TCPAddr
	var ln *net.TCPListener

	if addr, err = net.ResolveTCPAddr("tcp", address); err != nil {
		return
	}

	if ln, err = net.ListenTCP("tcp", addr); err != nil {
		return
	}

	l = &KeepAliveListener{TCPListener: ln}

	return
}

func (l *KeepAliveListener) Accept() (net.Conn, error) {
	conn, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(3 * time.Minute)
	l.count++

	return &tcpConn{
		KeepAliveListener: l,
		Conn:              conn,
	}, nil
}

type tcpConn struct {
	net.Conn
	*KeepAliveListener
}

func (c tcpConn) Close() error {
	c.KeepAliveListener.count--
	return c.Conn.Close()
}
