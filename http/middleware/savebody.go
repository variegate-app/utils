package middleware

import (
	"bytes"
	"io"
	"net/http"
)

type repeatableReader struct {
	io.Reader
	readBuf   *bytes.Buffer
	backupBuf *bytes.Buffer
}

func repeatableRead(r io.Reader) io.Reader {
	readBuf := bytes.Buffer{}
	readBuf.ReadFrom(r)
	backBuf := bytes.Buffer{}

	return repeatableReader{
		io.TeeReader(&readBuf, &backBuf),
		&readBuf,
		&backBuf,
	}
}

func (r repeatableReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		r.reset()
	}
	return n, err
}

func (r repeatableReader) Close() error { return nil }

func (r repeatableReader) reset() {
	io.Copy(r.readBuf, r.backupBuf)
}

// SaveBody избавляет от необходимости перезаписывать body в request
func SaveBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = io.NopCloser(repeatableRead(r.Body))
		next.ServeHTTP(w, r)
	})
}
