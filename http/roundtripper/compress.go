package roundtripper

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

type Compress struct {
	next http.RoundTripper
}

func NewCompress(next http.RoundTripper) *Compress {
	return &Compress{
		next: next,
	}
}

func (rt *Compress) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	b, err := io.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		return
	}
	if _, err = g.Write(b); err != nil {
		return
	}
	if err = g.Close(); err != nil {
		return
	}

	url := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	r, err := http.NewRequestWithContext(req.Context(), req.Method, url, &buf)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")
	r.Header.Set("Content-Type", req.Header.Get("Content-Type"))

	return rt.next.RoundTrip(r)
}
