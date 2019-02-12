/*
expamle:

package main

import (
	"time"

	"github.com/twiglab/twig"
	"github.com/twiglab/twig/middleware"
	"golang.org/x/time/rate"
)

func main() {
	web := twig.TODO()

	web.EnableDebug()
	web.Pre(middleware.NewRateLimiter(rate.NewLimiter(rate.Every(100), 1000)))

	web.Config().
		Get("/hello", twig.HelloTwig)

	web.Start()

	twig.Signal(twig.Graceful(web, 15*time.Second))

}
*/
package middleware

import "github.com/twiglab/twig"

type Allower interface {
	Allow() bool
}

type RateLimiterConifg struct {
	Skipper Skipper
	Allower Allower
}

var DefaultRateLimiterConfig = RateLimiterConifg{
	Skipper: DefaultSkipper,
}

func NewRateLimiter(allower Allower) twig.MiddlewareFunc {
	config := DefaultRateLimiterConfig
	config.Allower = allower
	return NewRateLimiterWithConifg(config)
}

func NewRateLimiterWithConifg(config RateLimiterConifg) twig.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultRateLimiterConfig.Skipper
	}
	if config.Allower == nil {
		panic("Limiter is nil")
	}

	return func(next twig.HandlerFunc) twig.HandlerFunc {
		return func(c twig.Ctx) error {
			if config.Skipper(c) || config.Allower.Allow() {
				return next(c)
			}
			return twig.ErrTooManyRequests
		}
	}
}
