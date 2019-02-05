package middleware

import (
	"github.com/twiglab/twig"
)

type RequestIDConfig struct {
	Skipper   Skipper
	Generator func(twig.Ctx) string
}

var DefaultRequestIDConfig = RequestIDConfig{
	Skipper:   DefaultSkipper,
	Generator: generator,
}

func generator(c twig.Ctx) string {
	return twig.GenID(c)
}

func RequestIDWithConfig(config RequestIDConfig) twig.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultRecoverConfig.Skipper
	}

	if config.Generator == nil {
		config.Generator = generator
	}

	return func(next twig.HandlerFunc) twig.HandlerFunc {
		return func(c twig.Ctx) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Req()
			resp := c.Resp()

			if id := req.Header.Get(twig.HeaderXRequestID); id == "" {
				resp.Header().Set(twig.HeaderXRequestID, config.Generator(c))
			}

			return next(c)
		}
	}
}

func RequestID() twig.MiddlewareFunc {
	return RequestIDWithConfig(DefaultRequestIDConfig)
}
