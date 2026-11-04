package qahub

import (
	"context"
	"database/sql"
	"errors"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrPreconditionNotFound = errors.New("precondition not found")
	ErrPreconditionCycle    = errors.New("preconditions form a cycle")
)

func (s *Store) Preconditions(ctx context.Context, id CaseID) ([]Case, error) {
	return s.queryCases(ctx, `
		SELECT c.id, c.title, c.description, c.priority
		FROM cases c
		JOIN case_preconditions link ON link.precondition_id = c.id
		WHERE link.case_id = ?
		ORDER BY c.id`, id)
}

func linkPreconditions(ctx context.Context, tx *sql.Tx, id CaseID, preconditions []CaseID) error {
	for _, precondition := range preconditions {
		_, err := tx.ExecContext(ctx,
			"INSERT OR IGNORE INTO case_preconditions (case_id, precondition_id) VALUES (?, ?)",
			id, precondition)
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY {
			return ErrPreconditionNotFound
		}
		if err != nil {
			return err
		}
	}
	var cycle bool
	err := tx.QueryRowContext(ctx, `
		WITH RECURSIVE reachable (id) AS (
			SELECT precondition_id FROM case_preconditions WHERE case_id = ?
			UNION
			SELECT link.precondition_id
			FROM case_preconditions link
			JOIN reachable ON link.case_id = reachable.id
		)
		SELECT EXISTS (SELECT 1 FROM reachable WHERE id = ?)`, id, id).Scan(&cycle)
	if err != nil {
		return err
	}
	if cycle {
		return ErrPreconditionCycle
	}
	return nil
}
