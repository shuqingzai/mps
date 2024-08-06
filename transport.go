package mps

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// DefaultTransport Default http.Transport option
var DefaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
}

// HttpRoundTripFunc is a function that implements http.RoundTripper.
type HttpRoundTripFunc func(req *http.Request) (resp *http.Response, err error)

// RoundTrip implements http.RoundTripper.
func (fn HttpRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// HttpRoundTripWrapper is http.RoundTripper middleware function.
type HttpRoundTripWrapper func(rt http.RoundTripper) http.RoundTripper

// HttpRoundTripWrapperFunc is a function that implements HttpRoundTripWrapper.
type HttpRoundTripWrapperFunc func(rt http.RoundTripper) HttpRoundTripFunc

func (f HttpRoundTripWrapperFunc) wrapper() HttpRoundTripWrapper {
	return func(rt http.RoundTripper) http.RoundTripper { return f(rt) }
}

// WrapRoundTripFunc adds a transport middleware function that will give the caller
func (ctx *Context) WrapRoundTripFunc(fns ...HttpRoundTripWrapperFunc) *Context {
	wrappers := make([]HttpRoundTripWrapper, 0, len(fns))
	for i := range fns {
		wrappers = append(wrappers, fns[i].wrapper())
	}
	return ctx.WrapRoundTrip(wrappers...)
}

// WrapRoundTrip adds a transport middleware that will give the caller
// the ability to modify the request and response.
func (ctx *Context) WrapRoundTrip(wrappers ...HttpRoundTripWrapper) *Context {
	if len(wrappers) == 0 {
		return ctx
	}
	if ctx.wrappedRoundTrip == nil {
		ctx.wrappedRoundTrip = HttpRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			return ctx.roundTrip(req)
		})
		ctx.httpRoundTripWrappers = wrappers
	} else {
		ctx.httpRoundTripWrappers = append(ctx.httpRoundTripWrappers, wrappers...)
	}

	for _, w := range wrappers {
		ctx.wrappedRoundTrip = w(ctx.wrappedRoundTrip)
	}
	return ctx
}
