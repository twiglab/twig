package middleware

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/twiglab/twig"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type RequestIDConfig struct {
	Skipper   Skipper
	Generator func() string
}

var DefaultRequestIDConfig = RequestIDConfig{
	Skipper:   DefaultSkipper,
	Generator: generator,
}

func generator() string {
	return fmt.Sprintf("%d_%d", gen(), rnd())
}

func rnd() int {
	return rand.Intn(77777)
}

func gen() int64 {
	return time.Now().UnixNano()
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
				resp.Header().Set(twig.HeaderXRequestID, config.Generator())
			}

			return next(c)
		}
	}
}

func RequestID() twig.MiddlewareFunc {
	return RequestIDWithConfig(DefaultRequestIDConfig)
}
