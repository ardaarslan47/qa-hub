package qahub

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	names, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	var applied int
	if err := db.QueryRow("PRAGMA user_version").Scan(&applied); err != nil {
		return err
	}
	if applied > len(names) {
		return fmt.Errorf("database is at version %d, this build knows only %d migrations", applied, len(names))
	}
	for i := applied; i < len(names); i++ {
		if err := applyMigration(db, i+1, names[i]); err != nil {
			return err
		}
	}
	return nil
}

func applyMigration(db *sql.DB, version int, name string) error {
	if want := fmt.Sprintf("migrations/%04d_", version); !strings.HasPrefix(name, want) {
		return fmt.Errorf("migration %s must be named %s<name>.sql", name, want)
	}
	script, err := migrations.ReadFile(name)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(string(script)); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return err
	}
	return tx.Commit()
}
