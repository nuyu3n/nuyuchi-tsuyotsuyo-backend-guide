package store

import (
	"context"
	"os"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		// データベースがない環境では、このパッケージのテストを飛ばす
		os.Exit(0)
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		panic(err)
	}
	// テーブルを作るSQLを読みこんで、実行する
	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		panic(err)
	}
	if _, err := db.Exec(ctx, string(schema)); err != nil {
		panic(err)
	}
	testDB = db
	// ここで、すべてのテストが動く
	code := m.Run()
	db.Close()
	// テストの結果（成功か失敗か）を終了コードとして返す
	os.Exit(code)
}

// beginTx は、テストが終わったら必ず取り消されるトランザクションを返す。
func beginTx(t *testing.T) pgx.Tx {
	// 失敗したとき、この関数ではなく、呼び出したテストの行を表示させる
	t.Helper()
	ctx := context.Background()
	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// テストが終わったらロールバックする
	t.Cleanup(func() { tx.Rollback(ctx) })
	return tx
}

func mustExec(t *testing.T, tx pgx.Tx, sql string, args ...any) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), sql, args...); err != nil {
		t.Fatal(err)
	}
}

func TestListVisibleApplicationIDs(t *testing.T) {
	const event, otherEvent = 1, 2
	// それぞれのユーザーID
	const leader, staff, viewer, nobody = 100, 200, 300, 400

	tests := []struct {
		name   string
		userID int64
		want   []int64
	}{
		{"代表は自分の団体の申し込みだけ", leader, []int64{1}},
		{"運営は行事のすべての申し込み", staff, []int64{1, 2}},
		{"閲覧のみの運営も行事のすべての申し込み", viewer, []int64{1, 2}},
		{"役割のない人には何も返らない", nobody, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := beginTx(t)
			// テスト用のデータを入れる。このトランザクションの中にしか入らない
			mustExec(t, tx, `INSERT INTO applications VALUES (1, $1, 10), (2, $1, 20), (3, $2, 10)`, event, otherEvent)
			mustExec(t, tx, `INSERT INTO role_assignments VALUES
				($1, $4, 'leader', 10), ($2, $4, 'staff', NULL), ($3, $4, 'viewer', NULL)`,
				leader, staff, viewer, event)

			got, err := ListVisibleApplicationIDs(context.Background(), tx, event, tt.userID)
			if err != nil {
				t.Fatal(err)
			}
			// スライスの中身が同じか比べる
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLeaderNeedsOrganization(t *testing.T) {
	tx := beginTx(t)
	// 団体のない代表を入れようとする。データベースの制約で断られるはず
	_, err := tx.Exec(context.Background(),
		`INSERT INTO role_assignments VALUES (100, 1, 'leader', NULL)`)
	if err == nil {
		t.Fatal("団体のない代表が登録できてしまった")
	}
}
