package web_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestCreatedCasesAppearInListInCreationOrder(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateCase("Login works")
	a.mustCreateCase(`<b>Cart</b> & "checkout"`)

	want := []string{"Login works", `<b>Cart</b> & "checkout"`}
	if got := a.caseTitles(); !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestBlankTitleIsRejectedWithAnError(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")

	for _, title := range []string{"", "   ", "\t\r\n", "\u00a0\u3000", "\u200b\ufeff\u2060"} {
		status, body := a.createCase(title)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("create case %q: status %d, want %d", title, status, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(body, "Title is required.") {
			t.Errorf("create case %q: page does not show the error\n%s", title, body)
		}
	}

	if got, want := a.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestTwoCasesMayShareATitle(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateCase("Login works")
	a.mustCreateCase("Login works")

	if got, want := a.caseTitles(), []string{"Login works", "Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestCasesSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	first.mustCreateCase("Login works")
	first.mustCreateCase("Logout works")
	first.stop()

	second := startApp(t, dbPath)

	if got, want := second.caseTitles(), []string{"Login works", "Logout works"}; !slices.Equal(got, want) {
		t.Errorf("case titles after restart = %q, want %q", got, want)
	}
}

func TestTitleIsTrimmed(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateCase(" \t Login works\u200b\n")

	if got, want := a.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}
