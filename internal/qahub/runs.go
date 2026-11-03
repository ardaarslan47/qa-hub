package qahub

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrNameRequired = errors.New("name is required")
	ErrNoCases      = errors.New("no cases to run")
	ErrNotFound     = errors.New("not found")
)

type Run struct {
	ID   int64
	Name string
}

func (s *Store) StartRun(ctx context.Context, name string) (Run, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Run{}, ErrNameRequired
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `INSERT INTO runs (name) VALUES (?)`, name)
	if err != nil {
		return Run{}, err
	}
	runID, err := result.LastInsertId()
	if err != nil {
		return Run{}, err
	}
	result, err = tx.ExecContext(ctx, `INSERT INTO executions (run_id, case_id, title) SELECT ?, id, title FROM cases ORDER BY id`, runID)
	if err != nil {
		return Run{}, err
	}
	executions, err := result.RowsAffected()
	if err != nil {
		return Run{}, err
	}
	if executions == 0 {
		return Run{}, ErrNoCases
	}
	if err := tx.Commit(); err != nil {
		return Run{}, err
	}
	return Run{ID: runID, Name: name}, nil
}

func (s *Store) Runs(ctx context.Context) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM runs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

func (s *Store) FindRun(ctx context.Context, id int64) (Run, []Execution, error) {
	var run Run
	err := s.db.QueryRowContext(ctx, `SELECT id, name FROM runs WHERE id = ?`, id).Scan(&run.ID, &run.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, nil, ErrNotFound
	}
	if err != nil {
		return Run{}, nil, err
	}
	executions, err := s.executions(ctx, id)
	return run, executions, err
}
