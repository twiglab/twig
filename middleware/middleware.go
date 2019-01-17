package middleware

import "github.com/twiglab/twig"

type Skipper func(twig.Ctx) bool
type BeforeFunc func(twig.Ctx)

func DefaultSkipper(_ twig.Ctx) bool {
	return false
}

func SuggestApi() []twig.MiddlewareFunc {
	return []twig.MiddlewareFunc{Recover()}
}
