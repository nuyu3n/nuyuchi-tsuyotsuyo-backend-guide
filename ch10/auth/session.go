package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"
)

const sessionTTL = 7 * 24 * time.Hour

// NewSessionToken は、ブラウザに渡すトークンと、データベースに保存するハッシュを作る。
func NewSessionToken() (token string, hash [32]byte, err error) {
	// 32バイトのランダムな値を作る
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", hash, err
	}
	// Cookie に入れられる文字だけの文字列にする
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, sha256.Sum256([]byte(token)), nil
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()), // 何秒で消えるか
		HttpOnly: true,                      // JavaScript から読めないようにする
		Secure:   true,                      // HTTPS のときだけ送る
		SameSite: http.SameSiteLaxMode,      // ほかのサイトからのリクエストには基本的に付けない
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // 負の値にすると、ブラウザはすぐに Cookie を消す
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
