package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/fickleDude/metrics.git/internal/logger"
)

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		//gather info
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   &responseData{status: 0, size: 0},
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		status := lw.responseData.status
		size := lw.responseData.size
		//request
		logger.Log.Info("HTTP request",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.String("duration", duration.String()),
		)
		//response
		logger.Log.Info("HTTP response",
			zap.Int("status", status),
			zap.Int("size", size),
		)
	}

	return http.HandlerFunc(logFn)
}
