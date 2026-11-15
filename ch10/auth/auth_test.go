package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	now := time.Date(2027, 4, 1, 9, 0, 0, 0, time.UTC)

	t.Run("正しいコードは1回だけ使える", func(t *testing.T) {
		code, lc, err := NewCode("a@example.com", now)
		if err != nil {
			t.Fatal(err)
		}
		if err := Verify(&lc, code, now); err != nil {
			t.Fatalf("1回目: %v", err)
		}
		if err := Verify(&lc, code, now); !errors.Is(err, ErrExpired) {
			t.Fatalf("2回目: %v, want ErrExpired", err)
		}
	})

	t.Run("期限を過ぎたら使えない", func(t *testing.T) {
		code, lc, _ := NewCode("a@example.com", now)
		err := Verify(&lc, code, now.Add(11*time.Minute))
		if !errors.Is(err, ErrExpired) {
			t.Fatalf("err = %v, want ErrExpired", err)
		}
	})

	t.Run("5回間違えたら、正しいコードでも使えない", func(t *testing.T) {
		code, lc, _ := NewCode("a@example.com", now)
		wrong := "000000"
		if code == wrong {
			wrong = "111111"
		}
		for range 5 {
			if err := Verify(&lc, wrong, now); !errors.Is(err, ErrWrongCode) {
				t.Fatalf("err = %v, want ErrWrongCode", err)
			}
		}
		if err := Verify(&lc, code, now); !errors.Is(err, ErrTooManyAttempts) {
			t.Fatalf("err = %v, want ErrTooManyAttempts", err)
		}
	})
}

func TestNewCodeIsSixDigits(t *testing.T) {
	for range 1000 {
		code, _, err := NewCode("a@example.com", time.Now())
		if err != nil {
			t.Fatal(err)
		}
		if len(code) != 6 {
			t.Fatalf("code = %q", code)
		}
	}
}

func TestSessionCookie(t *testing.T) {
	token, _, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	SetSessionCookie(rec, token)
	t.Log(len(token))
	t.Log(rec.Header().Get("Set-Cookie"))
}

func TestCrossOriginProtection(t *testing.T) {
	handler := http.NewCrossOriginProtection().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("approved"))
	}))

	for _, site := range []string{"same-origin", "cross-site"} {
		req := httptest.NewRequest("POST", "https://admin.example.com/applications/42/approve", nil)
		req.Header.Set("Sec-Fetch-Site", site)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		t.Logf("%s: %d %s", site, rec.Code, rec.Body.String())
	}
}
