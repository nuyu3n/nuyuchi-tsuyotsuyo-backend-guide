package main

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"log"
	"log/slog"
	"os"
	"strings"
	texttemplate "text/template"

	"github.com/jackc/pgx/v5/pgxpool"
)

func countByTitleUnsafe(ctx context.Context, db *pgxpool.Pool, title string) (int, error) {
	var n int
	// title をそのままSQLの文字列に埋めこんでしまう（危ない）
	query := fmt.Sprintf(`SELECT count(*) FROM applications WHERE title = '%s'`, title)
	err := db.QueryRow(ctx, query).Scan(&n)
	return n, err
}

func countByTitle(ctx context.Context, db *pgxpool.Pool, title string) (int, error) {
	var n int
	// title は $1 として、SQLとは別に渡す（安全）
	err := db.QueryRow(ctx, `SELECT count(*) FROM applications WHERE title = $1`, title).Scan(&n)
	return n, err
}

// Email は、ログに出すときに伏せ字にするメールアドレス。
// LogValue というメソッドを持つ型は、slog がログに書くとき、その結果を代わりに使う。
type Email string

func (e Email) LogValue() slog.Value {
	// @ の前と後ろに分ける。@ がなければ ok が false になる
	name, domain, ok := strings.Cut(string(e), "@")
	if !ok || name == "" {
		return slog.StringValue("***")
	}
	// 先頭の1文字だけ残して隠す（tanaka@example.com → t***@example.com）
	return slog.StringValue(name[:1] + "***@" + domain)
}

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(ctx, `
DROP TABLE IF EXISTS applications;
CREATE TABLE applications (id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, title text NOT NULL);
INSERT INTO applications (title) VALUES ('お化け屋敷'), ('写真展'), ('焼きそば');`)
	if err != nil {
		log.Fatal(err)
	}

	for _, input := range []string{"写真展", "x' OR '1'='1"} {
		n1, err1 := countByTitleUnsafe(ctx, db, input)
		n2, err2 := countByTitle(ctx, db, input)
		fmt.Printf("%-14q unsafe=%d %v / safe=%d %v\n", input, n1, err1, n2, err2)
	}

	// XSS
	title := `<script>alert("盗んだ")</script>`
	var a, b bytes.Buffer
	// text/template はそのまま埋めこみ、html/template は < などを無害な文字に置きかえる
	texttemplate.Must(texttemplate.New("").Parse(`<h1>{{.}}</h1>`)).Execute(&a, title)
	htmltemplate.Must(htmltemplate.New("").Parse(`<h1>{{.}}</h1>`)).Execute(&b, title)
	fmt.Println("text/template:", a.String())
	fmt.Println("html/template:", b.String())

	// ログの伏せ字
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	logger.Info("ログインコードを送った", "email", Email("tanaka@example.com"))
}
