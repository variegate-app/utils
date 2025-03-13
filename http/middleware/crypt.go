package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Crypter struct {
	key       []byte
	headerKey string
}

func NewCrypter(key []byte, headerKey string) *Crypter {
	return &Crypter{
		key:       key,
		headerKey: headerKey,
	}
}

func (c *Crypter) WithCrypt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerHash := r.Header.Get(c.headerKey)
		if headerHash != "" && c.key != nil {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				resp := make(map[string]string)
				w.WriteHeader(http.StatusInternalServerError)
				resp["error"] = err.Error()
				jsonResp, _ := json.Marshal(resp)
				_, _ = w.Write(jsonResp)
				return
			}

			r.Body = io.NopCloser(strings.NewReader(string(b)))
			h := hmac.New(sha256.New, c.key)
			h.Write(b)
			newHash := fmt.Sprintf("%x", h.Sum(nil))
			if headerHash != newHash {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
