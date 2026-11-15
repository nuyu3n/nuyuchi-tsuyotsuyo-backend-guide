package main

import (
	"encoding/json"
	"net/http"
)

// Problem は、エラーのときに返すJSONの形。RFC 9457 にそろえている。
type Problem struct {
	// `json:"..."` は、JSONにしたときの名前。omitempty は「空なら出さない」という意味。
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	Code   string `json:"code"`
}

// writeProblem は、エラーを Problem の形のJSONにして返す。
func writeProblem(w http.ResponseWriter, status int, code, detail string) {
	// 「これはエラーを表すJSONです」と相手に伝える
	w.Header().Set("Content-Type", "application/problem+json")
	// ステータスコード（404 など）を書く。本文より先に書く決まり
	w.WriteHeader(status)
	// 構造体をJSONに変換して書き出す
	json.NewEncoder(w).Encode(Problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Code:   code,
	})
}
