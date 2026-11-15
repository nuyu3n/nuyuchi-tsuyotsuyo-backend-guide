package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeStore は、データベースの代わりに使う偽物。
// Find と UpdateStatus を持っているので、Store として渡せる。
type fakeStore struct {
	app    Application
	status string
}

func (f *fakeStore) Find(ctx context.Context, id int64) (Application, error) {
	// 何を聞かれても、用意しておいた申し込みを返す
	return f.app, nil
}

func (f *fakeStore) UpdateStatus(ctx context.Context, id int64, status string) error {
	// 書きこまれたステータスを覚えておき、あとでテストから確認する
	f.status = status
	return nil
}

func TestSubmitAfterDeadline(t *testing.T) {
	deadline := time.Date(2027, 10, 1, 0, 0, 0, 0, time.UTC)
	store := &fakeStore{app: Application{ID: 1, Deadline: deadline}}
	s := &Service{
		Store: store,
		// 「今」を、締切の1分後に固定する
		Now: func() time.Time { return deadline.Add(time.Minute) },
	}

	err := s.Submit(context.Background(), 1)
	// 締切切れのエラーが返ってくるはず
	if !errors.Is(err, ErrDeadlinePassed) {
		t.Fatalf("err = %v, want ErrDeadlinePassed", err)
	}
	// ステータスは書きかえられていないはず
	if store.status != "" {
		t.Errorf("status = %q, want 変わらない", store.status)
	}
}
