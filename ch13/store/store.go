package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Querier は、プール（*pgxpool.Pool）とトランザクション（pgx.Tx）のどちらでも満たせる。
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// ListVisibleApplicationIDs は、userID の人が見てよい申し込みの ID を返す。
func ListVisibleApplicationIDs(ctx context.Context, q Querier, eventID, userID int64) ([]int64, error) {
	rows, err := q.Query(ctx, `
		SELECT a.id
		FROM applications AS a
		WHERE a.event_id = $1
		  AND EXISTS (
		    SELECT 1
		    FROM role_assignments AS r
		    WHERE r.user_id = $2
		      AND r.event_id = a.event_id
		      AND (
		        r.role IN ('staff', 'viewer')
		        OR (r.role = 'leader' AND r.organization_id = a.organization_id)
		      )
		  )
		ORDER BY a.id`, eventID, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[int64])
}
