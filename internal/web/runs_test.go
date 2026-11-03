package web_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestStartedRunAppearsInRunList(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")

	a.mustStartRun("Sprint 1")

	if got, want := a.runNames(), []string{"Sprint 1"}; !slices.Equal(got, want) {
		t.Errorf("run names = %q, want %q", got, want)
	}
}

func TestRunPageHasAnUntestedExecutionPerCase(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	a.mustStartRun("Sprint 1")

	run := a.runs()[0]

	want := []string{"Login works Untested", "Logout works Untested"}
	if got := a.executions(run); !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestBlankRunNameIsRejectedWithAnError(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")

	for _, name := range []string{"", "   ", "\t\r\n", "\u00a0\u3000", "\u200b\ufeff\u2060"} {
		status, body := a.startRun(name)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("start run %q: status %d, want %d", name, status, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(body, "Name is required.") {
			t.Errorf("start run %q: page does not show the error\n%s", name, body)
		}
	}

	if got := a.runNames(); len(got) != 0 {
		t.Errorf("run names = %q, want none", got)
	}
}

func TestCaseCreatedAfterRunStartedStaysOutOfThatRun(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	a.mustStartRun("Sprint 1")
	a.mustCreateCase("Reset password works")

	a.mustStartRun("Sprint 2")

	runs := a.runs()
	if len(runs) != 2 {
		t.Fatalf("runs = %v, want 2", runs)
	}
	first := []string{"Login works Untested", "Logout works Untested"}
	if got := a.executions(runs[0]); !slices.Equal(got, first) {
		t.Errorf("%s executions = %q, want %q", runs[0].Text, got, first)
	}
	second := []string{"Login works Untested", "Logout works Untested", "Reset password works Untested"}
	if got := a.executions(runs[1]); !slices.Equal(got, second) {
		t.Errorf("%s executions = %q, want %q", runs[1].Text, got, second)
	}
}

func TestTwoRunsMayShareAName(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")

	a.mustStartRun("Sprint 1")
	a.mustStartRun("Sprint 1")

	if got, want := a.runNames(), []string{"Sprint 1", "Sprint 1"}; !slices.Equal(got, want) {
		t.Errorf("run names = %q, want %q", got, want)
	}
	runs := a.runs()
	if runs[0].Href == runs[1].Href {
		t.Errorf("both runs link to %s", runs[0].Href)
	}
}

func TestRunNameIsTrimmed(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")

	a.mustStartRun(" \t Sprint 1​\n")

	if got, want := a.runNames(), []string{"Sprint 1"}; !slices.Equal(got, want) {
		t.Errorf("run names = %q, want %q", got, want)
	}
}

func TestRunsAndExecutionsSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	first.mustCreateCase("Login works")
	first.mustCreateCase("Logout works")
	first.mustStartRun("Sprint 1")
	first.mustCreateCase("Reset password works")
	first.mustStartRun("Sprint 2")
	first.stop()

	second := startApp(t, dbPath)

	if got, want := second.runNames(), []string{"Sprint 1", "Sprint 2"}; !slices.Equal(got, want) {
		t.Fatalf("run names after restart = %q, want %q", got, want)
	}
	runs := second.runs()
	wantExecutions := [][]string{
		{"Login works Untested", "Logout works Untested"},
		{"Login works Untested", "Logout works Untested", "Reset password works Untested"},
	}
	for i, run := range runs {
		if got := second.executions(run); !slices.Equal(got, wantExecutions[i]) {
			t.Errorf("%s executions after restart = %q, want %q", run.Text, got, wantExecutions[i])
		}
	}
}

func TestUnknownRunIsNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustStartRun("Sprint 1")

	for _, path := range []string{"/runs/2", "/runs/0", "/runs/-1", "/runs/abc", "/runs/1x"} {
		if status, _ := a.do(http.MethodGet, path, nil); status != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}
}

func TestRunCannotStartWithoutCases(t *testing.T) {
	a := startApp(t, tempDB(t))

	status, body := a.startRun("Sprint 1")

	if status != http.StatusUnprocessableEntity {
		t.Errorf("start run: status %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(body, "Create a case before starting a run.") {
		t.Errorf("page does not show the error\n%s", body)
	}
	if got := a.runNames(); len(got) != 0 {
		t.Errorf("run names = %q, want none", got)
	}
}
