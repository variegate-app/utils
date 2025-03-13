package roundtripper

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/variegate-app/utils/logger"

	"go.uber.org/zap"
)

type Logging struct {
	next http.RoundTripper
	log  *logger.Instance
}

func NewLogging(next http.RoundTripper, l *logger.Instance) *Logging {
	return &Logging{
		next: next,
		log:  l,
	}
}

func (rt *Logging) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	b, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		rt.log.ErrorCtx(req.Context(), "error reading body", zap.Error(err))
		return resp, err
	}

	defer func(begin time.Time) {
		rt.log.InfoCtx(req.Context(), "Request",
			zap.String("method", req.Method),
			zap.String("host", req.URL.Scheme+"://"+req.URL.Host+req.URL.Path),
			zap.Any("error", err),
			zap.Duration("took", time.Since(begin)),
			zap.ByteString("body", b),
		)
	}(time.Now())

	req.Body = io.NopCloser(strings.NewReader(string(b)))

	resp, err = rt.next.RoundTrip(req)

	if err != nil {
		rt.log.ErrorCtx(req.Context(), "error response", zap.Error(err))

		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		br, errResp := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if errResp != nil {
			rt.log.ErrorCtx(req.Context(), "error reading body", zap.Error(err))
			return resp, err
		}

		resp.Body = io.NopCloser(strings.NewReader(string(br)))

		rt.log.InfoCtx(req.Context(), "Response",
			zap.String("status", resp.Status),
			zap.String("host", req.URL.Scheme+"://"+req.URL.Host+req.URL.Path),
			zap.ByteString("body", br),
		)
	}

	return resp, err
}
