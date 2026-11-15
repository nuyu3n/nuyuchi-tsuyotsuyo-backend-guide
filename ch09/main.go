package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example/dbgen"
)

var ErrAlreadySubmitted = errors.New("もう提出されています")

func listTitles(ctx context.Context, db *pgxpool.Pool, eventID int64) ([]string, error) {
	rows, err := db.Query(ctx, `SELECT title FROM applications WHERE event_id = $1 ORDER BY id`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}
	return titles, rows.Err()
}

func approve(ctx context.Context, db *pgxpool.Pool, id, userID int64) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE applications SET status = 'approved' WHERE id = $1 AND status = 'submitted'`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("申し込み %d は承認できる状態ではありません", id)
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO notifications (user_id, body) VALUES ($1, $2)`, userID, "申し込みが承認されました")
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func submit(ctx context.Context, q *dbgen.Queries, id int64) error {
	n, err := q.SubmitApplication(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAlreadySubmitted
	}
	return nil
}

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 準備
	_, err = db.Exec(ctx, `
DROP TABLE IF EXISTS notifications, applications, organizations;
CREATE TABLE organizations (id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, name text NOT NULL);
CREATE TABLE applications (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id bigint NOT NULL,
  organization_id bigint NOT NULL REFERENCES organizations (id),
  title text NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  note text,
  created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE notifications (id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, user_id bigint NOT NULL, body text NOT NULL CHECK (body <> ''));
INSERT INTO organizations (name) VALUES ('2年1組'), ('写真部');
INSERT INTO applications (event_id, organization_id, title, status, note)
VALUES (1, 1, 'お化け屋敷', 'submitted', '暗幕を使う'), (1, 2, '写真展', 'draft', NULL);`)
	if err != nil {
		log.Fatal(err)
	}

	titles, err := listTitles(ctx, db, 1)
	fmt.Println("listTitles:", titles, err)

	// プレースホルダ
	var n int
	err = db.QueryRow(ctx, `SELECT count(*) FROM applications WHERE title = $1`, "写真展' OR '1'='1").Scan(&n)
	fmt.Println("placeholder:", n, err)

	// ANY
	rows, _ := db.Query(ctx, `SELECT name FROM organizations WHERE id = ANY($1) ORDER BY id`, []int64{1, 2})
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	fmt.Println("ANY:", names, err)

	// NULL
	var note1, note2 *string
	db.QueryRow(ctx, `SELECT note FROM applications WHERE id = 1`).Scan(&note1)
	db.QueryRow(ctx, `SELECT note FROM applications WHERE id = 2`).Scan(&note2)
	fmt.Println("NULL:", *note1, note2 == nil)
	var s string
	err = db.QueryRow(ctx, `SELECT note FROM applications WHERE id = 2`).Scan(&s)
	fmt.Println("NULL into string:", err)

	// トランザクション
	fmt.Println("approve 1:", approve(ctx, db, 1, 7))
	fmt.Println("approve 1 again:", approve(ctx, db, 1, 7))

	// sqlc
	q := dbgen.New(db)
	app, err := q.GetApplication(ctx, 2)
	fmt.Printf("sqlc get: %s %s note.Valid=%v %v\n", app.Title, app.Status, app.Note.Valid, err)
	list, err := q.ListApplicationsByEvent(ctx, 1)
	for _, a := range list {
		fmt.Println("sqlc list:", a.ID, a.Title, a.Status, a.OrganizationName)
	}
	fmt.Println("submit 2:", submit(ctx, q, 2))
	fmt.Println("submit 2 again:", submit(ctx, q, 2))
	_, err = q.GetApplication(ctx, 999)
	fmt.Println("not found:", err, errors.Is(err, pgx.ErrNoRows))

	// ctx のタイムアウト
	tctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	_, err = db.Exec(tctx, `SELECT pg_sleep(2)`)
	fmt.Println("timeout:", err)

	var cnt int
	db.QueryRow(ctx, `SELECT count(*) FROM notifications`).Scan(&cnt)
	fmt.Println("notifications:", cnt)
	fmt.Println("pool max conns:", db.Config().MaxConns)
}
