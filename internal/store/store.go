package store

import (
	"context"
	"database/sql"
	"embed"
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
