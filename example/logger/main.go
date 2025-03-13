package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/variegate-app/utils/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const exampleHost = "0.0.0.0:8180"

func main() {
	secretFields := []string{"password", "secret", "token"}
	l, err := logger.New(zapcore.DebugLevel, secretFields...)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cncl := context.WithTimeout(context.Background(), 2*time.Second)
	ctx = l.WithContextFields(ctx,
		zap.Int("pid", os.Getpid()),
		zap.String("app", "logger_example"))

	defer l.Sync()
	l.InfoCtx(ctx, "1", zap.String("password", "SECRET"))
	l.InfoCtx(ctx, "2", zap.String("secret", "SECRET"))
	l.InfoCtx(ctx, "3", zap.String("api_key", "SECRET"))
	l.InfoCtx(ctx, "4", zap.String("secret_key", "SECRET"))

	r := &http.ServeMux{}
	r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.InfoCtx(r.Context(), "5", zap.String("token", "SECRET"))
		w.WriteHeader(http.StatusInternalServerError)
	}))

	s := &http.Server{
		ErrorLog:          l.Std(),
		Handler:           r,
		Addr:              exampleHost,
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			cncl()
		}
	}()

	go func() {
		cli := &http.Client{Timeout: time.Second}
		req := &http.Request{
			Method: http.MethodGet,
			URL:    &url.URL{Scheme: "http", Host: exampleHost},
		}
		resp, _ := cli.Do(req)
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp != nil && resp.StatusCode == http.StatusInternalServerError {
			_ = s.Shutdown(ctx)
		}
	}()

	<-ctx.Done()
}
