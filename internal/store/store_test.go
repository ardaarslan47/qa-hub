package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ardaarslan47/qa-hub/internal/domain"
	"github.com/ardaarslan47/qa-hub/internal/store"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "qa-hub.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestDatabaseRefusesABlankTitleEvenWhenParsingWasSkipped(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)

	if err := st.CreateCase(ctx, domain.CaseTitle{}); err == nil {
		t.Fatal("create case with a zero CaseTitle succeeded, want an error")
	}

	cases, err := st.ListCases(ctx)
	if err != nil {
		t.Fatalf("list cases: %v", err)
	}
	if len(cases) != 0 {
		t.Fatalf("cases = %v, want none", cases)
	}
}

func TestDatabaseRefusesABlankRunNameEvenWhenParsingWasSkipped(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	title, err := domain.ParseCaseTitle("Login works")
	if err != nil {
		t.Fatalf("parse title: %v", err)
	}
	if err := st.CreateCase(ctx, title); err != nil {
		t.Fatalf("create case: %v", err)
	}

	if err := st.StartRun(ctx, domain.RunName{}); err == nil {
		t.Fatal("start run with a zero RunName succeeded, want an error")
	}

	runs, err := st.ListRuns(ctx)
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("runs = %v, want none", runs)
	}
}
