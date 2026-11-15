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

// コードの有効期限は10分、試せるのは5回まで
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
	Email string
	// SHA-256 のハッシュは32バイトなので、32バイトの配列 [32]byte で持つ
	CodeHash  [32]byte
	ExpiresAt time.Time
	Attempts  int
	Used      bool
}

// NewCode は、6桁のコードと、保存するための LoginCode を作る。
func NewCode(email string, now time.Time) (string, LoginCode, error) {
	// 0〜999999 の数を、予測できない方法で選ぶ（crypto/rand を使う）
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", LoginCode{}, err
	}
	// 6桁に足りないときは頭を0で埋める（42 → "000042"）
	code := fmt.Sprintf("%06d", n.Int64())
	// 利用者に送るのは code。保存するのはハッシュだけ
	return code, LoginCode{
		Email:     email,
		CodeHash:  sha256.Sum256([]byte(code)),
		ExpiresAt: now.Add(codeTTL),
	}, nil
}

// Verify は、入力されたコードを確かめる。lc は書き換えられるので、呼んだ後に保存し直す。
func Verify(lc *LoginCode, input string, now time.Time) error {
	// 使用済みか期限切れなら、もう使えない
	if lc.Used || now.After(lc.ExpiresAt) {
		return ErrExpired
	}
	// 間違いが多すぎたら、正しいコードでも受け付けない
	if lc.Attempts >= maxAttempts {
		return ErrTooManyAttempts
	}
	// 合っていても間違っていても、試した回数を1つ増やす
	lc.Attempts++
	// 入力をハッシュにして、保存しておいたハッシュと比べる。
	// ConstantTimeCompare は、何文字目まで合っていたかで時間が変わらない比べ方。
	// hash[:] は、配列をスライスとして渡すための書き方。
	hash := sha256.Sum256([]byte(input))
	if subtle.ConstantTimeCompare(hash[:], lc.CodeHash[:]) != 1 {
		return ErrWrongCode
	}
	// 一度使ったコードは、二度と使えないようにする
	lc.Used = true
	return nil
}
