package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ardaarslan47/qa-hub/internal/domain"
	"github.com/ardaarslan47/qa-hub/internal/store"
)

func TestDatabaseRefusesABlankTitleEvenWhenParsingWasSkipped(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(filepath.Join(t.TempDir(), "qa-hub.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })

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
