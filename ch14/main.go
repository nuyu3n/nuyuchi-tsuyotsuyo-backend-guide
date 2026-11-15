package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		level := slog.LevelInfo
		if rec.status >= 500 {
			level = slog.LevelError
		}
		logger.Log(r.Context(), level, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /applications/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":42}`))
	})
	mux.HandleFunc("POST /applications/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	handler := logRequests(logger, mux)

	for _, req := range []*http.Request{
		httptest.NewRequest("GET", "/applications/42", nil),
		httptest.NewRequest("POST", "/applications/42/approve", nil),
	} {
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
}
