package qahub

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrCaseNotFound  = errors.New("case not found")
)

type Title struct {
	text string
}

func ParseTitle(raw string) (Title, error) {
	text := trimInvisible(raw)
	if text == "" {
		return Title{}, ErrTitleRequired
	}
	return Title{text: text}, nil
}

func (t Title) String() string {
	return t.text
}

type CaseID int64

type Case struct {
	ID          CaseID
	Title       Title
	Description Description
	Priority    Priority
}

func (s *Store) CreateCase(ctx context.Context, title Title, description Description, priority Priority, preconditions []CaseID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		"INSERT INTO cases (title, description, priority) VALUES (?, ?, ?)",
		title.text, description.text, priority.name)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	if err := linkPreconditions(ctx, tx, CaseID(id), preconditions); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Cases(ctx context.Context) ([]Case, error) {
	return s.queryCases(ctx, "SELECT id, title, description, priority FROM cases ORDER BY id")
}

func (s *Store) queryCases(ctx context.Context, query string, args ...any) ([]Case, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cases []Case
	for rows.Next() {
		var c Case
		if err := rows.Scan(&c.ID, &c.Title.text, &c.Description.text, &c.Priority.name); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, rows.Err()
}

func (s *Store) FindCase(ctx context.Context, id CaseID) (Case, error) {
	c := Case{ID: id}
	err := s.db.QueryRowContext(ctx, "SELECT title, description, priority FROM cases WHERE id = ?", id).
		Scan(&c.Title.text, &c.Description.text, &c.Priority.name)
	if errors.Is(err, sql.ErrNoRows) {
		return Case{}, ErrCaseNotFound
	}
	return c, err
}

func (s *Store) UpdateCase(ctx context.Context, id CaseID, title Title, description Description, priority Priority, preconditions []CaseID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		"UPDATE cases SET title = ?, description = ?, priority = ? WHERE id = ?",
		title.text, description.text, priority.name, id)
	if err != nil {
		return err
	}
	matched, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if matched == 0 {
		return ErrCaseNotFound
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM case_preconditions WHERE case_id = ?", id); err != nil {
		return err
	}
	if err := linkPreconditions(ctx, tx, id, preconditions); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) DeleteCase(ctx context.Context, id CaseID) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM cases WHERE id = ?", id)
	if err != nil {
		return err
	}
	matched, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if matched == 0 {
		return ErrCaseNotFound
	}
	return nil
}
