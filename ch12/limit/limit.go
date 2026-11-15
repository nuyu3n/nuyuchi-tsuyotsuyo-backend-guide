package limit

import (
	"sync"
	"time"
)

// Limiter は、キー（メールアドレスやIPアドレス）ごとに、
// 一定の時間の中で何回まで許すかを数える。
type Limiter struct {
	mu     sync.Mutex    // 同時に使われても壊れないようにするための鍵
	max    int           // 許す回数
	window time.Duration // 数える時間の幅
	// キーごとに、許した時刻を覚えておく
	hits map[string][]time.Time
}

// New は、window の間に max 回まで許す Limiter を作る。
func New(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, hits: map[string][]time.Time{}}
}

func (l *Limiter) Allow(key string, now time.Time) bool {
	// 鍵をかけて、ほかのリクエストを待たせる。関数を抜けるときに外す
	l.mu.Lock()
	defer l.mu.Unlock()

	// window より古い記録は捨てて、最近のものだけ残す
	var recent []time.Time
	for _, t := range l.hits[key] {
		if now.Sub(t) < l.window {
			recent = append(recent, t)
		}
	}
	// もう上限に達していたら断る
	if len(recent) >= l.max {
		l.hits[key] = recent
		return false
	}
	// 今回の分を記録して、許す
	l.hits[key] = append(recent, now)
	return true
}
