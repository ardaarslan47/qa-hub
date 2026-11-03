package qahub

import (
	"context"
	"errors"
	"strings"
)

var ErrTitleRequired = errors.New("title is required")

type Case struct {
	ID    int64
	Title string
}

func (s *Store) CreateCase(ctx context.Context, title string) (Case, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Case{}, ErrTitleRequired
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO cases (title) VALUES (?)`, title)
	if err != nil {
		return Case{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Case{}, err
	}
	return Case{ID: id, Title: title}, nil
}

func (s *Store) Cases(ctx context.Context) ([]Case, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title FROM cases ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []Case
	for rows.Next() {
		var c Case
		if err := rows.Scan(&c.ID, &c.Title); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, rows.Err()
}
