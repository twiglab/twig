package middleware

import (
	"fmt"

	"github.com/twiglab/twig"
)

type (
	// SecureConfig 安全中间件配置
	SecureConfig struct {
		Skipper Skipper

		// XSSProtection 防御跨域攻击(XSS)
		// 设置 `X-XSS-Protection`
		// 可选，默认为 "1; mode=block".
		XSSProtection string

		// ContentTypeNosniff 设置`X-Content-Type-Options`
		ContentTypeNosniff string

		// XFrameOptions
		// 可选，默认为"SAMEORIGIN".
		// 可选的值为：
		// - "SAMEORIGIN"
		// - "DENY"
		// - "ALLOW-FROM uri"
		XFrameOptions string

		// HSTSMaxAge 设置`Strict-Transport-Security`
		// 可选，SSL链接下默认为31536000
		HSTSMaxAge int

		// HSTSExcludeSubdomains 排除子域名
		// 可选，默认false
		HSTSExcludeSubdomains bool

		ContentSecurityPolicy string
	}
)

var (
	DefaultSecureConfig = SecureConfig{
		Skipper:            DefaultSkipper,
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
		HSTSMaxAge:         31536000,
	}
)

// Secure 返回Secure中间件
func Secure() twig.MiddlewareFunc {
	return SecureWithConfig(DefaultSecureConfig)
}

// SecureWithConfig
func SecureWithConfig(config SecureConfig) twig.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultSecureConfig.Skipper
	}

	return func(next twig.HandlerFunc) twig.HandlerFunc {
		return func(c twig.Ctx) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Req()
			res := c.Resp()

			if config.XSSProtection != "" {
				res.Header().Set(twig.HeaderXXSSProtection, config.XSSProtection)
			}
			if config.ContentTypeNosniff != "" {
				res.Header().Set(twig.HeaderXContentTypeOptions, config.ContentTypeNosniff)
			}
			if config.XFrameOptions != "" {
				res.Header().Set(twig.HeaderXFrameOptions, config.XFrameOptions)
			}
			if (c.IsTls() || (req.Header.Get(twig.HeaderXForwardedProto) == "https")) && config.HSTSMaxAge != 0 {
				subdomains := ""
				if !config.HSTSExcludeSubdomains {
					subdomains = "; includeSubdomains"
				}
				res.Header().Set(twig.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", config.HSTSMaxAge, subdomains))
			}
			if config.ContentSecurityPolicy != "" {
				res.Header().Set(twig.HeaderContentSecurityPolicy, config.ContentSecurityPolicy)
			}
			return next(c)
		}
	}
}
