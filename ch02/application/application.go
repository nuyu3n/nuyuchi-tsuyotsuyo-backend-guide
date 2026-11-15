package application

import (
	"context"
	"errors"
	"time"
)

// 締切切れを表すエラー。呼び出した側は errors.Is で見分けられる。
var ErrDeadlinePassed = errors.New("締切を過ぎています")

type Application struct {
	ID       int64
	Status   string
	Deadline time.Time
}

// Store は、データの出し入れだけを受け持つ。
// interface なので、この2つのメソッドを持つ型なら何でも渡せる（本番はデータベース、テストでは偽物）。
type Store interface {
	Find(ctx context.Context, id int64) (Application, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

// Service は、業務のルールだけを受け持つ。HTTPもSQLも知らない。
type Service struct {
	Store Store
	// 今の時刻を返す関数。外から渡すので、テストでは好きな時刻にできる。
	Now func() time.Time
}

// Submit は、申し込みを提出済みにする。締切を過ぎていたらエラーを返す。
// ctx（context.Context）は、途中で処理をやめさせる合図を運ぶもの。チャプター9で扱う。
func (s *Service) Submit(ctx context.Context, id int64) error {
	app, err := s.Store.Find(ctx, id)
	if err != nil {
		return err
	}
	// 今の時刻が締切より後なら、提出させない
	if s.Now().After(app.Deadline) {
		return ErrDeadlinePassed
	}
	return s.Store.UpdateStatus(ctx, id, "submitted")
}
