package middleware

import (
	"fmt"

	"github.com/twiglab/twig"
)

type RecoverConfig struct {
	Skipper Skipper
}

var DefaultRecoverConfig = RecoverConfig{
	Skipper: DefaultSkipper,
}

func Recover() twig.MiddlewareFunc {
	return RecoverWithConfig(DefaultRecoverConfig)
}

func RecoverWithConfig(config RecoverConfig) twig.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultRecoverConfig.Skipper
	}

	return func(next twig.HandlerFunc) twig.HandlerFunc {
		return func(c twig.Ctx) error {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%V", r)
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}
