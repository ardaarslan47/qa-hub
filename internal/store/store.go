package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	_ "modernc.org/sqlite"

	"github.com/ardaarslan47/qa-hub/internal/domain"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	migrations, err := fs.Sub(embeddedMigrations, "migrations")
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := migrate(context.Background(), db, migrations); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) CreateCase(ctx context.Context, title domain.CaseTitle) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO cases (title) VALUES (?)`, title.String())
	return err
}

func (s *Store) ListCases(ctx context.Context) ([]domain.Case, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title FROM cases ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []domain.Case
	for rows.Next() {
		var c domain.Case
		if err := rows.Scan(&c.ID, &c.Title); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, rows.Err()
}

func (s *Store) StartRun(ctx context.Context, name domain.RunName) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	runResult, err := tx.ExecContext(ctx, `INSERT INTO runs (name) VALUES (?)`, name.String())
	if err != nil {
		return err
	}
	runID, err := runResult.LastInsertId()
	if err != nil {
		return err
	}
	executionsResult, err := tx.ExecContext(ctx,
		`INSERT INTO executions (run_id, case_id, title) SELECT ?, id, title FROM cases ORDER BY id`, runID)
	if err != nil {
		return err
	}
	executionCount, err := executionsResult.RowsAffected()
	if err != nil {
		return err
	}
	if executionCount == 0 {
		return domain.ErrNoCases
	}
	return tx.Commit()
}

func (s *Store) ListRuns(ctx context.Context) ([]domain.Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM runs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []domain.Run
	for rows.Next() {
		var r domain.Run
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

func (s *Store) GetRun(ctx context.Context, id int64) (domain.Run, error) {
	var r domain.Run
	err := s.db.QueryRowContext(ctx, `SELECT id, name FROM runs WHERE id = ?`, id).Scan(&r.ID, &r.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Run{}, domain.ErrRunNotFound
	}
	return r, err
}

func (s *Store) ListExecutions(ctx context.Context, runID int64) ([]domain.Execution, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, run_id, case_id, title, status FROM executions WHERE run_id = ? ORDER BY id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []domain.Execution
	for rows.Next() {
		var e domain.Execution
		var storedStatus sql.NullString
		if err := rows.Scan(&e.ID, &e.RunID, &e.CaseID, &e.Title, &storedStatus); err != nil {
			return nil, err
		}
		if storedStatus.Valid {
			status, err := domain.ParseStatus(storedStatus.String)
			if err != nil {
				return nil, fmt.Errorf("execution %d: stored status: %v", e.ID, err)
			}
			e.Status = &status
		}
		executions = append(executions, e)
	}
	return executions, rows.Err()
}

func (s *Store) SetExecutionStatus(ctx context.Context, id int64, status domain.Status) (int64, error) {
	var runID int64
	err := s.db.QueryRowContext(ctx,
		`UPDATE executions SET status = ? WHERE id = ? RETURNING run_id`, status.String(), id).Scan(&runID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrExecutionNotFound
	}
	return runID, err
}
