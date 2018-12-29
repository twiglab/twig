package middleware

import "github.com/twiglab/twig"

type Skipper func(twig.Ctx) bool

func DefaultSkipper(_ twig.Ctx) bool {
	return false
}

func Suggest() []twig.MiddlewareFunc {
	return []twig.MiddlewareFunc{Recover()}
}
