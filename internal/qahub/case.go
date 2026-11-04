package qahub

import (
	"context"
	"errors"
)

var ErrTitleRequired = errors.New("title is required")

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
	ID    CaseID
	Title Title
}

func (s *Store) CreateCase(ctx context.Context, project ProjectID, title Title) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO cases (project_id, title) VALUES (?, ?)", project, title.text)
	return err
}

func (s *Store) Cases(ctx context.Context, project ProjectID) ([]Case, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title FROM cases WHERE project_id = ? ORDER BY id", project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cases []Case
	for rows.Next() {
		var c Case
		if err := rows.Scan(&c.ID, &c.Title.text); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, rows.Err()
}
