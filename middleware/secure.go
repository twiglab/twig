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

		// ContentTypeNosniff provides protection against overriding Content-Type
		// header by setting the `X-Content-Type-Options` header.
		// Optional. Default value "nosniff".
		ContentTypeNosniff string

		// XFrameOptions can be used to indicate whether or not a browser should
		// be allowed to render a page in a <frame>, <iframe> or <object> .
		// Sites can use this to avoid clickjacking attacks, by ensuring that their
		// content is not embedded into other sites.provides protection against
		// clickjacking.
		// Optional. Default value "SAMEORIGIN".
		// Possible values:
		// - "SAMEORIGIN" - The page can only be displayed in a frame on the same origin as the page itself.
		// - "DENY" - The page cannot be displayed in a frame, regardless of the site attempting to do so.
		// - "ALLOW-FROM uri" - The page can only be displayed in a frame on the specified origin.
		XFrameOptions string `yaml:"x_frame_options"`

		// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
		// long (in seconds) browsers should remember that this site is only to
		// be accessed using HTTPS. This reduces your exposure to some SSL-stripping
		// man-in-the-middle (MITM) attacks.
		// Optional. Default value 0.
		HSTSMaxAge int `yaml:"hsts_max_age"`

		// HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`
		// header, excluding all subdomains from security policy. It has no effect
		// unless HSTSMaxAge is set to a non-zero value.
		// Optional. Default value false.
		HSTSExcludeSubdomains bool `yaml:"hsts_exclude_subdomains"`

		// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
		// security against cross-site scripting (XSS), clickjacking and other code
		// injection attacks resulting from execution of malicious content in the
		// trusted web page context.
		// Optional. Default value "".
		ContentSecurityPolicy string `yaml:"content_security_policy"`
	}
)

var (
	DefaultSecureConfig = SecureConfig{
		Skipper:            DefaultSkipper,
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
	}
)

// Secure returns a Secure middleware.
// Secure middleware provides protection against cross-site scripting (XSS) attack,
// content type sniffing, clickjacking, insecure connection and other code injection
// attacks.
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
