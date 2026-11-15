package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

// statusRecorder は、ResponseWriter を包んで、書かれたステータスコードを覚えておく。
// http.ResponseWriter をフィールド名なしで書くと（埋めこみ）、そのメソッドをそのまま使える。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader だけ横取りして、コードを覚えてから本物に渡す。
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logRequests は、next の前後にログを書く処理を足したハンドラを返す（ミドルウェア）。
func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// WriteHeader が呼ばれなかったときは 200 になるので、最初は 200 にしておく
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		// 本来の処理を呼ぶ。w の代わりに rec を渡すので、ステータスコードが記録される
		next.ServeHTTP(rec, r)

		level := slog.LevelInfo
		// 500番台はサーバー側の失敗なので、エラーとして記録する
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
	// JSON の形でログを標準出力に書く logger を作る
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /applications/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":42}`))
	})
	mux.HandleFunc("POST /applications/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	// mux へのすべてのリクエストを、ログに残す
	handler := logRequests(logger, mux)

	for _, req := range []*http.Request{
		httptest.NewRequest("GET", "/applications/42", nil),
		httptest.NewRequest("POST", "/applications/42/approve", nil),
	} {
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
}
