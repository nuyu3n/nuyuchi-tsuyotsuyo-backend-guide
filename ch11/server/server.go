package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"example/authz"
)

type store interface {
	FindApplication(ctx context.Context, id int64) (authz.Application, error)
}

// currentActor は、ログインの確認（チャプター10）で分かった人と役割を返す。ここでは形だけ。
func currentActor(r *http.Request) authz.Actor { return authz.Actor{} }

func writeJSON(w http.ResponseWriter, v any) { json.NewEncoder(w).Encode(v) }

// 穴のあるハンドラ
func getApplicationBad(s store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		app, err := s.FindApplication(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, app)
	}
}

// 直したハンドラ
func getApplication(s store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		app, err := s.FindApplication(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if !authz.Can(currentActor(r), authz.ActionView, app) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, app)
	}
}
