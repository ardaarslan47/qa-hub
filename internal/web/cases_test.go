package web_test

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestCreatedCasesAppearInLibraryInCreationOrder(t *testing.T) {
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

func TestPriorityAndDescriptionSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	first.mustCreateCaseWith("Login works", "Sign in.", "High")
	first.mustCreateCaseWith("Logout works", "", "Low")
	first.stop()

	second := startApp(t, dbPath).inNewProject("Mobile")
	second.mustStartRun("Sprint 1")

	wantCases := []libraryCase{{"Login works", "High"}, {"Logout works", "Low"}}
	if got := second.libraryCases(); !slices.Equal(got, wantCases) {
		t.Errorf("library cases after restart = %v, want %v", got, wantCases)
	}
	wantFields := []executionFields{{Priority: "High", Description: "Sign in."}, {Priority: "Low"}}
	if got := second.executionFields(second.runs()[0]); !slices.Equal(got, wantFields) {
		t.Errorf("execution fields = %q, want %q", got, wantFields)
	}
}

func TestTitleIsTrimmed(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateCase(" \t Login works\u200b\n")

	if got, want := a.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestLibraryIsEmptyUntilACaseIsCreated(t *testing.T) {
	a := startApp(t, tempDB(t))

	if got := a.caseTitles(); len(got) != 0 {
		t.Errorf("case titles = %q, want none", got)
	}
}

func TestPriorityIsChosenFromLowMediumAndHighWithMediumSelected(t *testing.T) {
	a := startApp(t, tempDB(t))

	want := []option{{"Low", false}, {"Medium", true}, {"High", false}}
	if got := a.priorityOptions(); !slices.Equal(got, want) {
		t.Errorf("priority options = %v, want %v", got, want)
	}
}

func TestLibraryListsEachCasesPriority(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateCaseWith("Login works", "", "High")
	a.mustCreateCaseWith("Logout works", "", "Low")
	a.mustCreateCase("Reset password works")

	want := []libraryCase{{"Login works", "High"}, {"Logout works", "Low"}, {"Reset password works", "Medium"}}
	if got := a.libraryCases(); !slices.Equal(got, want) {
		t.Errorf("library cases = %v, want %v", got, want)
	}
}

func TestDescriptionMayBeLeftBlank(t *testing.T) {
	a := startApp(t, tempDB(t)).inNewProject("Mobile")
	blanks := []string{"", "   ", "\t\r\n", "\u00a0\u3000", "\u200b\ufeff\u2060"}

	for i, description := range blanks {
		title := fmt.Sprintf("Case %d", i)
		if status, body := a.createCaseWith(title, description, "Medium"); status != http.StatusOK {
			t.Errorf("create case with description %q: status %d\n%s", description, status, body)
		}
	}

	if got := a.caseTitles(); len(got) != len(blanks) {
		t.Fatalf("case titles = %q, want %d cases", got, len(blanks))
	}
	a.mustStartRun("Sprint 1")
	want := slices.Repeat([]executionFields{{Priority: "Medium"}}, len(blanks))
	if got := a.executionFields(a.runs()[0]); !slices.Equal(got, want) {
		t.Errorf("execution fields = %q, want %q", got, want)
	}
}

func TestRunPageShowsEachExecutionsDescriptionAndPriority(t *testing.T) {
	a := startApp(t, tempDB(t)).inNewProject("Mobile")
	a.mustCreateCaseWith("Login works", "Sign in with a valid\nemail and password.", "High")
	a.mustCreateCaseWith("Logout works", "", "Low")
	a.mustCreateCaseWith("Cart <b>totals</b>", `  Shows "tax" & <i>shipping</i>  `, "Medium")
	a.mustStartRun("Sprint 1")

	want := []executionFields{
		{Priority: "High", Description: "Sign in with a valid\nemail and password."},
		{Priority: "Low"},
		{Priority: "Medium", Description: `Shows "tax" & <i>shipping</i>`},
	}
	if got := a.executionFields(a.runs()[0]); !slices.Equal(got, want) {
		t.Errorf("execution fields = %q, want %q", got, want)
	}
}

func TestRejectedFormKeepsWhatTheUserTyped(t *testing.T) {
	a := startApp(t, tempDB(t))

	status, body := a.createCaseWith("", "Sign in with a valid\nemail and password.", "High")
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if got, want := fieldValue(t, body, "description"), "Sign in with a valid\nemail and password."; got != want {
		t.Errorf("description = %q, want %q", got, want)
	}
	wantOptions := []option{{"Low", false}, {"Medium", false}, {"High", true}}
	if got := priorityOptionsOf(t, body); !slices.Equal(got, wantOptions) {
		t.Errorf("priority options = %v, want %v", got, wantOptions)
	}

	status, body = a.createCaseWith("Login works", "Sign in.", "Urgent")
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if got, want := fieldValue(t, body, "title"), "Login works"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	if got, want := fieldValue(t, body, "description"), "Sign in."; got != want {
		t.Errorf("description = %q, want %q", got, want)
	}
	wantOptions = []option{{"Low", false}, {"Medium", true}, {"High", false}}
	if got := priorityOptionsOf(t, body); !slices.Equal(got, wantOptions) {
		t.Errorf("priority options = %v, want %v", got, wantOptions)
	}
}

func TestInvalidOrMissingPriorityIsRejectedWithAnError(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")

	forms := map[string]url.Values{
		"missing":      {"title": {"Logout works"}},
		"empty":        {"title": {"Logout works"}, "priority": {""}},
		"unknown":      {"title": {"Logout works"}, "priority": {"Urgent"}},
		"lower case":   {"title": {"Logout works"}, "priority": {"high"}},
		"padded":       {"title": {"Logout works"}, "priority": {" Medium"}},
		"repeated bad": {"title": {"Logout works"}, "priority": {"Critical", "Low"}},
	}
	for name, form := range forms {
		status, body := a.do(http.MethodPost, "/library", form)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("%s priority: status %d, want %d", name, status, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(body, "Priority must be Low, Medium or High.") {
			t.Errorf("%s priority: page does not show the error\n%s", name, body)
		}
	}

	if got, want := a.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}
