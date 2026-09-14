package downloader

import (
	"context"
	"io"

	"golang.org/x/time/rate"
)

// ThrottledReader wraps an io.Reader and rate-limits byte consumption
type ThrottledReader struct {
	r       io.Reader
	limiter *rate.Limiter
	ctx     context.Context
}

// NewThrottledReader returns an io.Reader that respects the given rate limiter
func NewThrottledReader(ctx context.Context, r io.Reader, limiter *rate.Limiter) io.Reader {
	if limiter == nil {
		return r 
	}
	return &ThrottledReader{
		r:       r,
		limiter: limiter,
		ctx:     ctx,
	}
}

func (tr *ThrottledReader) Read(p []byte) (int, error) {
	n, err := tr.r.Read(p)
	if n > 0 && tr.limiter != nil {
		// WaitN reserves 'n' tokens from the bucket or blocks until tokens are available.
		// If ctx is cancelled while waiting, it unblocks and returns ctx.Err() immediately.
		if waitErr := tr.limiter.WaitN(tr.ctx, n); waitErr != nil {
			return n, waitErr
		}
	}
	return n, err
}
