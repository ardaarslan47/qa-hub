package qahub

import (
	"context"
	"database/sql"
	"errors"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrProjectNameRequired = errors.New("project name is required")
	ErrProjectNameTaken    = errors.New("project name is taken")
	ErrProjectNotFound     = errors.New("project not found")
)

type ProjectName struct {
	text string
}

func ParseProjectName(raw string) (ProjectName, error) {
	text := trimInvisible(raw)
	if text == "" {
		return ProjectName{}, ErrProjectNameRequired
	}
	return ProjectName{text: text}, nil
}

func (n ProjectName) String() string {
	return n.text
}

type ProjectID int64

type Project struct {
	ID   ProjectID
	Name ProjectName
}

func (s *Store) CreateProject(ctx context.Context, name ProjectName) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO projects (name) VALUES (?)", name.text)
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return ErrProjectNameTaken
	}
	return err
}

func (s *Store) Projects(ctx context.Context) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name FROM projects ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name.text); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *Store) FindProject(ctx context.Context, id ProjectID) (Project, error) {
	p := Project{ID: id}
	err := s.db.QueryRowContext(ctx, "SELECT name FROM projects WHERE id = ?", id).Scan(&p.Name.text)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	return p, err
}
