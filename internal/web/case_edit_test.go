package web_test

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestBlankTitleOrInvalidPriorityRejectsAnEditAndLeavesTheCaseUnchanged(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCaseWith("Login works", "Sign in.", "High")
	login := a.caseLink("Login works")
	want := caseFields{Title: "Login works", Description: "Sign in.", Priority: "High"}

	for _, title := range []string{"", "   ", "\t\r\n", "\u00a0\u3000", "\u200b\ufeff\u2060"} {
		status, body := a.editCase(login, title, "Changed.", "Low")
		if status != http.StatusUnprocessableEntity {
			t.Errorf("edit with title %q: status %d, want %d", title, status, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(body, "Title is required.") {
			t.Errorf("edit with title %q: page does not show the error\n%s", title, body)
		}
	}

	forms := map[string]url.Values{
		"missing":      {"title": {"Changed"}, "description": {"Changed."}},
		"empty":        {"title": {"Changed"}, "description": {"Changed."}, "priority": {""}},
		"unknown":      {"title": {"Changed"}, "description": {"Changed."}, "priority": {"Urgent"}},
		"lower case":   {"title": {"Changed"}, "description": {"Changed."}, "priority": {"high"}},
		"padded":       {"title": {"Changed"}, "description": {"Changed."}, "priority": {" Medium"}},
		"repeated bad": {"title": {"Changed"}, "description": {"Changed."}, "priority": {"Critical", "Low"}},
	}
	for name, form := range forms {
		status, body := a.do(http.MethodPost, login.Href, form)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("%s priority: status %d, want %d", name, status, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(body, "Priority must be Low, Medium or High.") {
			t.Errorf("%s priority: page does not show the error\n%s", name, body)
		}
	}

	if got := a.caseFields(login); got != want {
		t.Errorf("case after rejected edits = %+v, want %+v", got, want)
	}
	if got, wantList := a.libraryCases(), []libraryCase{{"Login works", "High"}}; !slices.Equal(got, wantList) {
		t.Errorf("library cases = %v, want %v", got, wantList)
	}
}

func TestRejectedEditKeepsWhatTheUserTyped(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCaseWith("Login works", "Sign in.", "High")
	login := a.caseLink("Login works")

	status, body := a.editCase(login, "", "Sign in with a valid\nemail and password.", "Low")
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if got, want := caseFieldsOf(t, body), (caseFields{Description: "Sign in with a valid\nemail and password.", Priority: "Low"}); got != want {
		t.Errorf("rejected form = %+v, want %+v", got, want)
	}

	status, body = a.editCase(login, "Logout works", "Sign out.", "Urgent")
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if got, want := caseFieldsOf(t, body), (caseFields{Title: "Logout works", Description: "Sign out.", Priority: "High"}); got != want {
		t.Errorf("rejected form = %+v, want %+v", got, want)
	}
}

func TestEditingAMissingCaseIs404(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	login := a.caseLink("Login works")
	form := url.Values{"title": {"Changed"}, "description": {"Changed."}, "priority": {"Low"}}

	for _, path := range []string{"/library/0", "/library/-1", "/library/99999", "/library/abc", "/library/9223372036854775808"} {
		if status, _ := a.do(http.MethodGet, path, nil); status != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
		}
		if status, _ := a.do(http.MethodPost, path, form); status != http.StatusNotFound {
			t.Errorf("POST %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}

	if got, want := a.caseTitles(), []string{"Login works", "Logout works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
	if got, want := a.caseFields(login), (caseFields{Title: "Login works", Priority: "Medium"}); got != want {
		t.Errorf("case = %+v, want %+v", got, want)
	}
}

func TestEditSavesTitleDescriptionAndPriorityWithTheSameTrimmingAsCreate(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCaseWith("Login works", "Sign in.", "Low")
	a.mustCreateCase("Logout works")
	login := a.caseLink("Login works")

	a.mustEditCase(login, " \t Logout works\u200b\n", "  Sign out.  ", "High")

	want := caseFields{Title: "Logout works", Description: "Sign out.", Priority: "High"}
	if got := a.caseFields(login); got != want {
		t.Errorf("case = %+v, want %+v", got, want)
	}
	wantList := []libraryCase{{"Logout works", "High"}, {"Logout works", "Medium"}}
	if got := a.libraryCases(); !slices.Equal(got, wantList) {
		t.Errorf("library cases = %v, want %v", got, wantList)
	}

	a.mustEditCase(login, "Logout works", "\u200b ", "Low")

	if got, want := a.caseFields(login).Description, ""; got != want {
		t.Errorf("description = %q, want %q", got, want)
	}
}
