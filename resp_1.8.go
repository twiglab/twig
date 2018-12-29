// +build go1.8

package twig

import "net/http"

func (r *ResponseWarp) Push(target string, opts *http.PushOptions) error {
	return r.Writer.(http.Pusher).Push(target, opts)
}
