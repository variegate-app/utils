package roundtripper

import (
	"context"
	"net/http"
	"time"
)

type Retry struct {
	next http.RoundTripper
	wait []time.Duration
}

func NewRetry(next http.RoundTripper, wait ...time.Duration) *Retry {
	return &Retry{
		next: next,
		wait: wait,
	}
}

func (rt *Retry) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = rt.next.RoundTrip(req)

	if err != nil || resp.StatusCode == http.StatusInternalServerError {
		return rt.RetryRequest(req.Context(), req)
	}

	return resp, err
}

func (rt *Retry) RetryRequest(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	intervals := make(chan struct{})

	go func(ctx context.Context, c chan<- struct{}, w []time.Duration) {
		for _, in := range w {
			select {
			case <-time.After(in):
				c <- struct{}{}
			case <-ctx.Done():
				close(c)
				return
			}
		}
		close(c)
	}(req.Context(), intervals, rt.wait)

	for {
		select {
		case _, ok := <-intervals:
			if !ok {
				return resp, err
			}

			resp, err = rt.next.RoundTrip(req)
			if err != nil || resp.StatusCode == http.StatusInternalServerError {
				continue
			}

			return resp, err
		case <-req.Context().Done():
			return resp, err
		}
	}
}
