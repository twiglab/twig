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

type matchedMux struct {
	Muxer
	Matcher
}

type muxes struct {
	ms  []*matchedMux
	def Muxer
}

func (m *muxes) Lookup(method string, path string, req *http.Request) MuxerCtx {
	for _, mux := range m.ms {
		if mux.Match(req) {
			return mux.Lookup(method, path, req)
		}
	}

	return m.def.Lookup(method, path, req)
}

func (m *muxes) AddHandler(method string, path string, h HandlerFunc, ms ...MiddlewareFunc) Router {
	return m.def.AddHandler(method, path, h, ms...)
}

func (m *muxes) Use(ms ...MiddlewareFunc) {
	m.def.Use(ms...)
}

func (m *muxes) AddMuxer(mux Muxer, match Matcher) {
	mf := &matchedMux{
		Muxer:   m,
		Matcher: match,
	}
	m.ms = append(m.ms, mf)
}
