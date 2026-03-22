package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipWriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || !slices.Contains([]string{"application/json", "text/html"}, r.Header.Get("Content-Type")) {
			next.ServeHTTP(w, r)
			return
		}
		gzip, err := gzip.NewWriterLevel(w, flate.BestCompression)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gzip.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gzip}, r)
	})
}

type gzipReader struct {
	http.Request
	Reader io.ReadCloser
}

func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}

func (r *gzipReader) Close() error {
	return r.Reader.Close()
}

func GzipReader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gzip, err := gzip.NewReader(r.Body)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gzip.Close()
		body, err := io.ReadAll(gzip)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		w.Write(body)
	})
}
