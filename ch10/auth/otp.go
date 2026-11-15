package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"time"
)

const (
	codeTTL     = 10 * time.Minute
	maxAttempts = 5
)

var (
	ErrExpired         = errors.New("コードの有効期限が切れています")
	ErrTooManyAttempts = errors.New("試行回数の上限に達しました")
	ErrWrongCode       = errors.New("コードが違います")
)

// LoginCode は、データベースに保存する1件のログインコード。
// コードそのものではなく、ハッシュだけを持つ。
type LoginCode struct {
	Email     string
	CodeHash  [32]byte
	ExpiresAt time.Time
	Attempts  int
	Used      bool
}

// NewCode は、6桁のコードと、保存するための LoginCode を作る。
func NewCode(email string, now time.Time) (string, LoginCode, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", LoginCode{}, err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	return code, LoginCode{
		Email:     email,
		CodeHash:  sha256.Sum256([]byte(code)),
		ExpiresAt: now.Add(codeTTL),
	}, nil
}

// Verify は、入力されたコードを確かめる。lc は書き換えられるので、呼んだ後に保存し直す。
func Verify(lc *LoginCode, input string, now time.Time) error {
	if lc.Used || now.After(lc.ExpiresAt) {
		return ErrExpired
	}
	if lc.Attempts >= maxAttempts {
		return ErrTooManyAttempts
	}
	lc.Attempts++
	hash := sha256.Sum256([]byte(input))
	if subtle.ConstantTimeCompare(hash[:], lc.CodeHash[:]) != 1 {
		return ErrWrongCode
	}
	lc.Used = true
	return nil
}
