package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
DROP TABLE IF EXISTS entries, slots;
CREATE TABLE slots (
  id       bigint PRIMARY KEY,
  capacity int    NOT NULL
);
CREATE TABLE entries (
  id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slot_id bigint NOT NULL REFERENCES slots (id),
  user_id bigint NOT NULL
);
INSERT INTO slots (id, capacity) VALUES (1, 1);
`

// 守っていない申し込み
func enterUnsafe(ctx context.Context, db *pgxpool.Pool, userID int64) error {
	var capacity, count int
	err := db.QueryRow(ctx, `SELECT capacity FROM slots WHERE id = 1`).Scan(&capacity)
	if err != nil {
		return err
	}
	err = db.QueryRow(ctx, `SELECT count(*) FROM entries WHERE slot_id = 1`).Scan(&count)
	if err != nil {
		return err
	}
	if count >= capacity {
		return fmt.Errorf("満員です")
	}
	time.Sleep(100 * time.Millisecond) // ほかの処理に時間がかかった、という想定
	_, err = db.Exec(ctx, `INSERT INTO entries (slot_id, user_id) VALUES (1, $1)`, userID)
	return err
}

// 行をロックして守る申し込み
func enterLocked(ctx context.Context, db *pgxpool.Pool, userID int64) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var capacity, count int
	err = tx.QueryRow(ctx, `SELECT capacity FROM slots WHERE id = 1 FOR UPDATE`).Scan(&capacity)
	if err != nil {
		return err
	}
	err = tx.QueryRow(ctx, `SELECT count(*) FROM entries WHERE slot_id = 1`).Scan(&count)
	if err != nil {
		return err
	}
	if count >= capacity {
		return fmt.Errorf("満員です")
	}
	time.Sleep(100 * time.Millisecond) // ほかの処理に時間がかかった、という想定
	_, err = tx.Exec(ctx, `INSERT INTO entries (slot_id, user_id) VALUES (1, $1)`, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func run(ctx context.Context, db *pgxpool.Pool, name string, enter func(context.Context, *pgxpool.Pool, int64) error) {
	if _, err := db.Exec(ctx, schema); err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			enter(ctx, db, int64(i))
		})
	}
	wg.Wait()

	var count int
	db.QueryRow(ctx, `SELECT count(*) FROM entries`).Scan(&count)
	fmt.Printf("%s: 定員1人に %d人 入った\n", name, count)
}

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	run(ctx, db, "守っていない", enterUnsafe)
	run(ctx, db, "ロックした", enterLocked)
}
