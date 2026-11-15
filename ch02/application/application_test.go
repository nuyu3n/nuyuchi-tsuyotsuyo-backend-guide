package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	app    Application
	status string
}

func (f *fakeStore) Find(ctx context.Context, id int64) (Application, error) {
	return f.app, nil
}

func (f *fakeStore) UpdateStatus(ctx context.Context, id int64, status string) error {
	f.status = status
	return nil
}

func TestSubmitAfterDeadline(t *testing.T) {
	deadline := time.Date(2027, 10, 1, 0, 0, 0, 0, time.UTC)
	store := &fakeStore{app: Application{ID: 1, Deadline: deadline}}
	s := &Service{
		Store: store,
		Now:   func() time.Time { return deadline.Add(time.Minute) },
	}

	err := s.Submit(context.Background(), 1)
	if !errors.Is(err, ErrDeadlinePassed) {
		t.Fatalf("err = %v, want ErrDeadlinePassed", err)
	}
	if store.status != "" {
		t.Errorf("status = %q, want 変わらない", store.status)
	}
}
