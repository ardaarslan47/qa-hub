package store_test

import (
	"context"
	"errors"
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

	err = st.StartRun(ctx, domain.RunName{})

	if err == nil {
		t.Fatal("start run with a zero RunName succeeded, want an error")
	}
	if errors.Is(err, domain.ErrNoCases) {
		t.Fatalf("start run with a zero RunName failed with %v, want the database CHECK to refuse it", err)
	}

	runs, err := st.ListRuns(ctx)
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("runs = %v, want none", runs)
	}
}

func TestDatabaseRefusesAStatusOutsideTheClosedSetEvenWhenParsingWasSkipped(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	title, err := domain.ParseCaseTitle("Login works")
	if err != nil {
		t.Fatalf("parse title: %v", err)
	}
	if err := st.CreateCase(ctx, title); err != nil {
		t.Fatalf("create case: %v", err)
	}
	name, err := domain.ParseRunName("Sprint 1")
	if err != nil {
		t.Fatalf("parse name: %v", err)
	}
	if err := st.StartRun(ctx, name); err != nil {
		t.Fatalf("start run: %v", err)
	}
	runs, err := st.ListRuns(ctx)
	if err != nil || len(runs) != 1 {
		t.Fatalf("list runs = %v, %v, want one run", runs, err)
	}
	executions, err := st.ListExecutions(ctx, runs[0].ID)
	if err != nil || len(executions) != 1 {
		t.Fatalf("list executions = %v, %v, want one execution", executions, err)
	}

	_, err = st.SetExecutionStatus(ctx, executions[0].ID, domain.Status{})

	if err == nil {
		t.Fatal("set a zero Status succeeded, want an error")
	}
	if errors.Is(err, domain.ErrExecutionNotFound) {
		t.Fatalf("set a zero Status failed with %v, want the database CHECK to refuse it", err)
	}
	after, err := st.ListExecutions(ctx, runs[0].ID)
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	if after[0].Status != nil {
		t.Fatalf("status = %v, want none", *after[0].Status)
	}
}
