package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/variegate-app/utils/graceful"
)

const exampleHost = "0.0.0.0:8180"
const defaultLifetime = 10 * time.Second

type server struct {
	s *http.Server
}

type client struct {
	c *http.Client
	r chan<- struct{}
}

func NewServer() *server {
	r := &http.ServeMux{}
	r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	return &server{
		s: &http.Server{
			Handler:           r,
			Addr:              exampleHost,
			ReadHeaderTimeout: 1 * time.Second,
		},
	}
}

func NewClient(r chan<- struct{}) *client {
	return &client{
		c: &http.Client{Timeout: time.Second},
		r: r,
	}
}

func (s *server) Run(ctx context.Context) error {
	go func(c context.Context) {
		<-c.Done()
		_ = s.s.Shutdown(c)
		fmt.Println("server exited")
	}(ctx)

	return s.s.ListenAndServe()
}

func (c *client) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			close(c.r)
			fmt.Println("client exited")
			return nil
		default:
			req := &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Scheme: "http", Host: exampleHost},
			}
			resp, _ := c.c.Do(req)
			_ = resp.Body.Close()
			c.r <- struct{}{}
		}
	}
}

func main() {
	ctx, cncl := context.WithTimeout(context.Background(), defaultLifetime)
	defer cncl()

	result := make(chan struct{})
	srv := NewServer()
	cli := NewClient(result)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			select {
			case _, ok := <-result:
				if !ok {
					return
				}
				fmt.Println("request")
			case <-ctx.Done():
				fmt.Println("deadline")
				return
			}
		}
	}()

	manager := graceful.New(ctx, 1*time.Second)
	manager.AddTask(cli, srv)

	_ = manager.Wait(syscall.SIGTERM, syscall.SIGINT)
}
