package image

import (
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
)

type Option func(*options)
type options struct {
	srcImageAuth, dstImageAuth           authn.Authenticator
	srcRounderTripper, dstRounderTripper http.RoundTripper
	srcParallelism, dstParallelism       int
}

func newDefaultOption() *options {
	return &options{
		srcImageAuth:      authn.Anonymous,
		dstImageAuth:      authn.Anonymous,
		srcRounderTripper: http.DefaultTransport.(*http.Transport).Clone(),
		dstRounderTripper: http.DefaultTransport.(*http.Transport).Clone(),
		srcParallelism:    10,
		dstParallelism:    10,
	}
}

func SetSrcImageAuth(auth authn.Authenticator) Option {
	return func(o *options) { o.srcImageAuth = auth }
}
func SetDstImageAuth(auth authn.Authenticator) Option {
	return func(o *options) { o.dstImageAuth = auth }
}
func SetSrcRoundTripper(tripper http.RoundTripper) Option {
	return func(o *options) { o.srcRounderTripper = tripper }
}
func SetDstRoundTripper(tripper http.RoundTripper) Option {
	return func(o *options) { o.dstRounderTripper = tripper }
}

func SetSrcParallelism(parallelism int) Option {
	return func(o *options) {
		if parallelism > 0 {
			o.srcParallelism = parallelism
		}
	}
}
func SetDstParallelism(parallelism int) Option {
	return func(o *options) {
		if parallelism > 0 {
			o.dstParallelism = parallelism
		}
	}
}
