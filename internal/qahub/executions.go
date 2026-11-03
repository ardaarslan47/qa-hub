package qahub

import (
	"context"
	"database/sql"
	"errors"
)

var ErrInvalidStatus = errors.New("status must be Passed or Failed")

type Status string

const (
	Passed Status = "Passed"
	Failed Status = "Failed"
)

func ParseStatus(value string) (Status, error) {
	switch status := Status(value); status {
	case Passed, Failed:
		return status, nil
	}
	return "", ErrInvalidStatus
}

type Execution struct {
	ID     int64
	CaseID int64
	Title  string
	Status *Status
}

func (s *Store) RecordStatus(ctx context.Context, runID, executionID int64, status Status) error {
	result, err := s.db.ExecContext(ctx, `UPDATE executions SET status = ? WHERE id = ? AND run_id = ?`, string(status), executionID, runID)
	if err != nil {
		return err
	}
	matched, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if matched == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) executions(ctx context.Context, runID int64) ([]Execution, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, case_id, title, status FROM executions WHERE run_id = ? ORDER BY id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []Execution
	for rows.Next() {
		var e Execution
		var status sql.NullString
		if err := rows.Scan(&e.ID, &e.CaseID, &e.Title, &status); err != nil {
			return nil, err
		}
		if status.Valid {
			recorded := Status(status.String)
			e.Status = &recorded
		}
		executions = append(executions, e)
	}
	return executions, rows.Err()
}
