package limit

import (
	"testing"
	"time"
)

func TestAllow(t *testing.T) {
	// 1時間に3回まで
	l := New(3, time.Hour)
	now := time.Date(2027, 4, 1, 9, 0, 0, 0, time.UTC)

	// 1分おきに3回は許される
	for i := range 3 {
		if !l.Allow("a@example.com", now.Add(time.Duration(i)*time.Minute)) {
			t.Fatalf("%d回目が拒否された", i+1)
		}
	}
	if l.Allow("a@example.com", now.Add(10*time.Minute)) {
		t.Error("4回目が許された")
	}
	if !l.Allow("b@example.com", now.Add(10*time.Minute)) {
		t.Error("別の人まで拒否された")
	}
	if !l.Allow("a@example.com", now.Add(61*time.Minute)) {
		t.Error("1時間たっても拒否された")
	}
}
