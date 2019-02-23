package twig

import (
	"net/http"
)

type Matcher interface {
	Match(req *http.Request) bool
}

type MatcherFunc func(req *http.Request) bool

func (f MatcherFunc) Match(req *http.Request) bool {
	return f(req)
}

type MatchedMux struct {
	Muxer
	Matcher
}

func NewMatchedMux(m Muxer, f MatcherFunc) *MatchedMux {
	return &MatchedMux{
		Muxer:   m,
		Matcher: f,
	}
}

type Muxes struct {
	ms  []*MatchedMux
	def Muxer
}

func (m *Muxes) Lookup(method string, path string, req *http.Request) MuxerCtx {
	for _, mux := range m.ms {
		if mux.Match(req) {
			return mux.Lookup(method, path, req)
		}
	}

	return m.def.Lookup(method, path, req)
}

func (m *Muxes) AddHandler(method string, path string, h HandlerFunc, ms ...MiddlewareFunc) Router {
	return m.def.AddHandler(method, path, h, ms...)
}

func (m *Muxes) Use(ms ...MiddlewareFunc) {
	m.def.Use(ms...)
}

func (m *Muxes) AddMuxer(mux Muxer, f MatcherFunc) {
	mf := NewMatchedMux(mux, f)
	m.ms = append(m.ms, mf)
}
