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
	// 読み終わったら閉じる
	defer rows.Close()

	var titles []string
	// 1行ずつ読み進める。行がなくなると false になり、ループが終わる
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}
	// 読んでいる途中でエラーが起きていないかも返す
	return titles, rows.Err()
}

func approve(ctx context.Context, db *pgxpool.Pool, id, userID int64) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	// 途中で return したら、それまでの変更はすべて取り消される
	defer tx.Rollback(ctx)

	// 提出済みのものだけを承認済みにする
	tag, err := tx.Exec(ctx,
		`UPDATE applications SET status = 'approved' WHERE id = $1 AND status = 'submitted'`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("申し込み %d は承認できる状態ではありません", id)
	}
	// 同じトランザクションの中で、通知も作る
	_, err = tx.Exec(ctx,
		`INSERT INTO notifications (user_id, body) VALUES ($1, $2)`, userID, "申し込みが承認されました")
	if err != nil {
		return err
	}
	// 両方うまくいったときだけ確定する
	return tx.Commit(ctx)
}

func submit(ctx context.Context, q *dbgen.Queries, id int64) error {
	// n は、書きかえた行の数
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
	// 環境変数 DATABASE_URL から接続先を読み、接続のプールを作る
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		// 作れなければ、ここでプログラムを終わらせる
		log.Fatal(err)
	}
	// プログラムが終わるときに接続を閉じる
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

	// ANY: スライス []int64{1, 2} をそのまま $1 に渡せる。ANY($1) は「この中のどれか」
	rows, _ := db.Query(ctx, `SELECT name FROM organizations WHERE id = ANY($1) ORDER BY id`, []int64{1, 2})
	// すべての行を読んで、string のスライスにまとめる
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	fmt.Println("ANY:", names, err)

	// NULL: *string は「string を指すポインタ」。何も指していない状態（nil）で NULL を表せる
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

	// sqlc が作ったコードを使う。SQLを書かずに、Goの関数として呼べる
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

	// ctx のタイムアウト: 100ミリ秒たったら時間切れになる ctx を作る
	tctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	// 使い終わったら後片付けをする
	defer cancel()
	_, err = db.Exec(tctx, `SELECT pg_sleep(2)`)
	fmt.Println("timeout:", err)

	var cnt int
	db.QueryRow(ctx, `SELECT count(*) FROM notifications`).Scan(&cnt)
	fmt.Println("notifications:", cnt)
	fmt.Println("pool max conns:", db.Config().MaxConns)
}
