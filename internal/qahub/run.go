package qahub

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRunNameRequired = errors.New("run name is required")
	ErrRunNotFound     = errors.New("run not found")
	ErrNoCases         = errors.New("no cases to run")
)

type RunName struct {
	text string
}

func ParseRunName(raw string) (RunName, error) {
	text := trimInvisible(raw)
	if text == "" {
		return RunName{}, ErrRunNameRequired
	}
	return RunName{text: text}, nil
}

func (n RunName) String() string {
	return n.text
}

type RunID int64

type Run struct {
	ID   RunID
	Name RunName
}

type ExecutionID int64

type Execution struct {
	ID    ExecutionID
	Title Title
}

func (s *Store) StartRun(ctx context.Context, name RunName) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, "INSERT INTO runs (name) VALUES (?)", name.text)
	if err != nil {
		return err
	}
	runID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	res, err = tx.ExecContext(ctx,
		"INSERT INTO executions (run_id, case_id, title) SELECT ?, id, title FROM cases ORDER BY id", runID)
	if err != nil {
		return err
	}
	snapshotted, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if snapshotted == 0 {
		return ErrNoCases
	}
	return tx.Commit()
}

func (s *Store) Runs(ctx context.Context) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name FROM runs ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.Name.text); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

func (s *Store) FindRun(ctx context.Context, id RunID) (Run, error) {
	r := Run{ID: id}
	err := s.db.QueryRowContext(ctx, "SELECT name FROM runs WHERE id = ?", id).Scan(&r.Name.text)
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, ErrRunNotFound
	}
	return r, err
}

func (s *Store) Executions(ctx context.Context, run RunID) ([]Execution, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title FROM executions WHERE run_id = ? ORDER BY id", run)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var executions []Execution
	for rows.Next() {
		var e Execution
		if err := rows.Scan(&e.ID, &e.Title.text); err != nil {
			return nil, err
		}
		executions = append(executions, e)
	}
	return executions, rows.Err()
}
