package application

import (
	"context"
	"errors"
	"time"
)

var ErrDeadlinePassed = errors.New("締切を過ぎています")

type Application struct {
	ID       int64
	Status   string
	Deadline time.Time
}

// Store は、データの出し入れだけを受け持つ。
type Store interface {
	Find(ctx context.Context, id int64) (Application, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

// Service は、業務のルールだけを受け持つ。HTTPもSQLも知らない。
type Service struct {
	Store Store
	Now   func() time.Time
}

func (s *Service) Submit(ctx context.Context, id int64) error {
	app, err := s.Store.Find(ctx, id)
	if err != nil {
		return err
	}
	if s.Now().After(app.Deadline) {
		return ErrDeadlinePassed
	}
	return s.Store.UpdateStatus(ctx, id, "submitted")
}
