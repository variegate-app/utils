package middleware

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/variegate-app/utils/logger"

	"go.uber.org/zap"
)

func WithLog(h http.Handler, l *logger.Instance) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			l.ErrorCtx(r.Context(), "error reading body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(strings.NewReader(string(b)))

		h.ServeHTTP(w, r)

		duration := time.Since(start)

		l.InfoCtx(r.Context(), "new request",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.Duration("duration", duration),
			zap.ByteString("body", b),
		)
	}

	return http.HandlerFunc(logFn)
}
