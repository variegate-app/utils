package roundtripper

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

type crypt struct {
	rt        http.RoundTripper
	key       []byte
	headerKey string
}

func NewCrypt(rt http.RoundTripper, key []byte, headerKey string) http.RoundTripper {
	return &crypt{
		rt:  rt,
		key: key,
	}
}

func (c *crypt) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	b, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return resp, err
	}

	h := hmac.New(sha256.New, c.key)
	h.Write(b)
	req.Header.Add(c.headerKey, fmt.Sprintf("%x", h.Sum(nil)))
	req.Body = io.NopCloser(bytes.NewReader(b))
	return c.rt.RoundTrip(req)
}
