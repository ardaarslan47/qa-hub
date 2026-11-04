package qahub

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRunNameRequired   = errors.New("run name is required")
	ErrRunNotFound       = errors.New("run not found")
	ErrNoCases           = errors.New("no cases to run")
	ErrExecutionNotFound = errors.New("execution not found")
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
	ID            ExecutionID
	Title         Title
	Description   Description
	Priority      Priority
	Preconditions []Title
	Status        *Status
}

func (s *Store) StartRun(ctx context.Context, project ProjectID, name RunName) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, "INSERT INTO runs (project_id, name) VALUES (?, ?)", project, name.text)
	if err != nil {
		return err
	}
	runID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	res, err = tx.ExecContext(ctx,
		`INSERT INTO executions (run_id, case_id, title, description, priority)
		SELECT ?, id, title, description, priority FROM cases ORDER BY id`,
		runID)
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
	_, err = tx.ExecContext(ctx,
		`INSERT INTO execution_preconditions (execution_id, title)
		SELECT e.id, c.title
		FROM executions e
		JOIN case_preconditions link ON link.case_id = e.case_id
		JOIN cases c ON c.id = link.precondition_id
		WHERE e.run_id = ?
		ORDER BY e.id, c.id`,
		runID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Runs(ctx context.Context, project ProjectID) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name FROM runs WHERE project_id = ? ORDER BY id", project)
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

func (s *Store) FindRun(ctx context.Context, project ProjectID, id RunID) (Run, error) {
	r := Run{ID: id}
	err := s.db.QueryRowContext(ctx, "SELECT name FROM runs WHERE id = ? AND project_id = ?", id, project).Scan(&r.Name.text)
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, ErrRunNotFound
	}
	return r, err
}

func (s *Store) Executions(ctx context.Context, project ProjectID, run RunID) ([]Execution, error) {
	executions, err := s.executionRows(ctx, project, run)
	if err != nil {
		return nil, err
	}
	copies, err := s.preconditionCopies(ctx, project, run)
	if err != nil {
		return nil, err
	}
	for i := range executions {
		executions[i].Preconditions = copies[executions[i].ID]
	}
	return executions, nil
}

func (s *Store) executionRows(ctx context.Context, project ProjectID, run RunID) ([]Execution, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.title, e.description, e.priority, e.status
		FROM executions e
		JOIN runs r ON r.id = e.run_id
		WHERE e.run_id = ? AND r.project_id = ?
		ORDER BY e.id`, run, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var executions []Execution
	for rows.Next() {
		var e Execution
		var status sql.NullString
		if err := rows.Scan(&e.ID, &e.Title.text, &e.Description.text, &e.Priority.name, &status); err != nil {
			return nil, err
		}
		if status.Valid {
			e.Status = &Status{name: status.String}
		}
		executions = append(executions, e)
	}
	return executions, rows.Err()
}

func (s *Store) preconditionCopies(ctx context.Context, project ProjectID, run RunID) (map[ExecutionID][]Title, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.execution_id, p.title
		FROM execution_preconditions p
		JOIN executions e ON e.id = p.execution_id
		JOIN runs r ON r.id = e.run_id
		WHERE e.run_id = ? AND r.project_id = ?
		ORDER BY p.id`, run, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	copies := map[ExecutionID][]Title{}
	for rows.Next() {
		var execution ExecutionID
		var title Title
		if err := rows.Scan(&execution, &title.text); err != nil {
			return nil, err
		}
		copies[execution] = append(copies[execution], title)
	}
	return copies, rows.Err()
}

func (s *Store) EnterStatus(ctx context.Context, project ProjectID, run RunID, execution ExecutionID, status Status) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE executions SET status = ?
		WHERE id = ? AND run_id = (SELECT id FROM runs WHERE id = ? AND project_id = ?)`,
		status.name, execution, run, project)
	if err != nil {
		return err
	}
	matched, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if matched == 0 {
		return ErrExecutionNotFound
	}
	return nil
}
