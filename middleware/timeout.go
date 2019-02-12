package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/twiglab/twig"
)

type TimeOutConfig struct {
	Skipper  Skipper
	Duration time.Duration
}

var DefaultTimeOutConfig = TimeOutConfig{
	Skipper: DefaultSkipper,
}

func NewTimeOut(timeout time.Duration) twig.MiddlewareFunc {
	config := DefaultTimeOutConfig
	config.Duration = timeout
	return NewTimeOutWithConfig(config)
}

func NewTimeOutWithConfig(config TimeOutConfig) twig.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultTimeOutConfig.Skipper
	}

	if config.Duration <= 0 {
		config.Skipper = DefaultTimeOutConfig.Skipper
	}

	return func(next twig.HandlerFunc) twig.HandlerFunc {
		return func(c twig.Ctx) error {
			if config.Skipper(c) {
				return next(c)
			}

			r := c.Req()
			ctx, cancel := context.WithTimeout(r.Context(), config.Duration)
			r.WithContext(ctx)

			defer func(twigc twig.Ctx) {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					twigc.Resp().WriteHeader(http.StatusGatewayTimeout)
				}
			}(c)

			return next(c)
		}
	}
}
