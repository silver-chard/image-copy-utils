package image

import (
	"context"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func CopyImage(ctx context.Context, srcRef, dstRef string, opts ...Option) error {
	opt := newDefaultOption()
	for _, o := range opts {
		o(opt)
	}

	src, err := name.ParseReference(srcRef)
	if err != nil {
		return err
	}
	dst, err := name.ParseReference(dstRef)
	if err != nil {
		return err
	}
	pullOpt := remote.WithTransport(opt.srcRounderTripper)
	pushOpt := remote.WithTransport(opt.dstRounderTripper)
	puller, err := remote.NewPuller(remote.WithAuth(opt.srcImageAuth), pullOpt, remote.WithJobs(opt.srcParallelism))
	if err != nil {
		return err
	}
	pusher, err := remote.NewPusher(remote.WithAuth(opt.dstImageAuth), pushOpt, remote.WithJobs(opt.dstParallelism))
	if err != nil {
		return err
	}
	desc, err := puller.Get(ctx, src)
	if err != nil {
		return err
	}
	if err = pusher.Push(ctx, dst, desc); err != nil {
		return err
	}
	return nil
}
