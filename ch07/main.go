package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type row struct {
	ID      int64
	OrgID   int64
	OrgName string
}

// N+1: 一覧を取ってから、1件ずつ団体の名前を取りに行く
func listNPlusOne(ctx context.Context, db *pgxpool.Pool) ([]row, int, error) {
	queries := 0
	rows, err := db.Query(ctx, `SELECT id, organization_id FROM applications ORDER BY id LIMIT 100`)
	queries++
	if err != nil {
		return nil, queries, err
	}
	var list []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.OrgID); err != nil {
			return nil, queries, err
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		return nil, queries, err
	}
	for i := range list {
		// list[i] の団体の名前を、list[i].OrgName に読み込む
		err := db.QueryRow(ctx, `SELECT name FROM organizations WHERE id = $1`, list[i].OrgID).Scan(&list[i].OrgName)
		queries++
		if err != nil {
			return nil, queries, err
		}
	}
	return list, queries, nil
}

// JOIN: 1回で取る
func listJoin(ctx context.Context, db *pgxpool.Pool) ([]row, int, error) {
	// JOIN で、申し込みと団体の名前を1回のSQLでまとめて取る
	rows, err := db.Query(ctx, `
		SELECT a.id, a.organization_id, o.name
		FROM applications AS a
		JOIN organizations AS o ON o.id = a.organization_id
		ORDER BY a.id
		LIMIT 100`)
	if err != nil {
		return nil, 1, err
	}
	defer rows.Close()
	var list []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.OrgID, &r.OrgName); err != nil {
			return nil, 1, err
		}
		list = append(list, r)
	}
	return list, 1, rows.Err()
}

func measure(ctx context.Context, db *pgxpool.Pool, name string, f func(context.Context, *pgxpool.Pool) ([]row, int, error)) {
	start := time.Now()
	list, queries, err := f(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %d件, SQL %d回, %v\n", name, len(list), queries, time.Since(start).Round(time.Millisecond))
}

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	measure(ctx, db, "N+1 ", listNPlusOne)
	measure(ctx, db, "JOIN", listJoin)
}
