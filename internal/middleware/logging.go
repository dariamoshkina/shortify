package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		size   int
		status int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.responseData.status = status
}

func WithLogging(l *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			respData := &responseData{size: 0, status: 0}
			responseWriter := &loggingResponseWriter{ResponseWriter: w, responseData: respData}

			uri := r.RequestURI
			method := r.Method

			h.ServeHTTP(responseWriter, r)

			duration := time.Since(start)

			l.Infoln("uri", uri, "method", method, "duration", duration, "status", respData.status, "size", respData.size)
		}

		return http.HandlerFunc(logFn)
	}
}
