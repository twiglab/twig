package middleware

import "github.com/twiglab/twig"

type Skipper func(twig.C) bool

func DefaultSkipper(_ twig.C) bool {
	return false
}

func Suggest() []twig.MiddlewareFunc {
	return []twig.MiddlewareFunc{}
}
