# ぬゆちのつよつよバックエンド解説 サンプルコード

Zennの本「[ぬゆちのつよつよバックエンド解説](https://zenn.dev/nuyuchi/books/nuyuchi-tsuyotsuyo-backend-guide)」に出てくるコードとSQLを、チャプターごとにまとめたものです。

この本は何かを1つ完成させる本ではないので、ここにあるのは「本文の例を、実際に動かして確かめるためのコード」です。本文の実行結果は、ここにあるコードで出したものです。

## 動作を確かめた環境

- Go 1.27.4
- PostgreSQL 18.7（Dockerの `postgres:18-alpine`）

## PostgreSQLを用意する

```sh
docker run --name pg -e POSTGRES_PASSWORD=pw -p 5432:5432 -d postgres:18-alpine
docker exec pg createdb -U postgres sample
```

> [!WARNING]
> いくつかのプログラムは、始めに `DROP TABLE` してテーブルを作り直します。必ず、このサンプルのためだけのデータベースを指定してください。

## チャプターごとの中身

| フォルダ | チャプター | 中身 | 動かし方 |
|---|---|---|---|
| `ch02` | 2 バックエンドの全体像 | 業務のルールを、HTTPやSQLなしでテストする | `go test ./...` |
| `ch03` | 3 APIを設計する | エラーの形（RFC 9457）、OpenAPIの定義 | `go run .`、`cd openapi-check && go run . openapi.yaml broken.yaml` |
| `ch04` | 4 SQLの基本 | SELECT、JOIN、集計、NULL、JSONB | `psql` で `ch04.sql` を実行 |
| `ch05` | 5 テーブルを設計する | 制約、多対多、フォームの版、保管 | `psql` で `ch05.sql` を実行 |
| `ch06` | 6 トランザクションと同時実行 | 定員1人に20人が同時に申し込む実験 | `DATABASE_URL=... go run .` |
| `ch07` | 7 インデックスと性能 | 50万行での EXPLAIN、N+1問題 | `psql` で `ch07.sql`、その後 `DATABASE_URL=... go run .` |
| `ch08` | 8 マイグレーション | goose のマイグレーション | `GOOSE_DRIVER=postgres GOOSE_DBSTRING=... go tool goose -dir migrations up` |
| `ch09` | 9 GoからDBを使う | pgx、トランザクション、sqlc | `go tool sqlc generate`、`DATABASE_URL=... go run .` |
| `ch10` | 10 認証 | ワンタイムコード、セッション、Cookie、CSRF | `go test ./... -v` |
| `ch11` | 11 認可 | 権限の判定と、そのテスト。一覧のSQL | `go test ./... -v`、`psql` で `list.sql` |
| `ch12` | 12 セキュリティと個人情報 | SQLインジェクション、XSS、ログの伏せ字、回数の制限、匿名化 | `DATABASE_URL=... go run .`、`go test ./limit`、`psql` で `ch12.sql` |
| `ch13` | 13 テスト | 本物のPostgreSQLを使うテスト | `TEST_DATABASE_URL=... go test ./... -v` |
| `ch14` | 14 長く運用する | リクエストのログ、書き換えられない監査ログ | `go run .`、`psql` で `ch14.sql` |

`DATABASE_URL` などには、たとえば `postgres://postgres:pw@localhost:5432/sample` を入れます。

`psql` でSQLのファイルを実行するには、次のようにします。

```sh
docker exec -i pg psql -U postgres -d sample < ch04/ch04.sql
```

SQLのファイルの中には、エラーになることを確かめるための文（制約に違反する `INSERT` など）も入っています。そこでエラーが出るのは正しい動きです。

## ライセンス

[MIT License](LICENSE)
