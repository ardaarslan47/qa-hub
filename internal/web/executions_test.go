package web_test

import (
	"net/http"
	"net/url"
	"slices"
	"testing"
)

func TestEnteredStatusesAppearOnRunPage(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]

	a.mustEnterStatus(run, "Login works", "Passed")
	a.mustEnterStatus(run, "Logout works", "Failed")

	want := []string{"Login works Passed", "Logout works Failed"}
	if got := a.executions(run); !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestNewStatusReplacesThePreviousOne(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]

	a.mustEnterStatus(run, "Login works", "Failed")
	a.mustEnterStatus(run, "Login works", "Passed")
	a.mustEnterStatus(run, "Logout works", "Passed")
	a.mustEnterStatus(run, "Logout works", "Failed")
	a.mustEnterStatus(run, "Logout works", "Failed")

	want := []string{"Login works Passed", "Logout works Failed"}
	if got := a.executions(run); !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestStatusesBelongToTheirOwnRun(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustStartRun("Sprint 1")
	a.mustStartRun("Sprint 2")
	runs := a.runs()

	a.mustEnterStatus(runs[0], "Login works", "Failed")

	if got, want := a.executions(runs[0]), []string{"Login works Failed"}; !slices.Equal(got, want) {
		t.Errorf("%s executions = %q, want %q", runs[0].Text, got, want)
	}
	if got, want := a.executions(runs[1]), []string{"Login works Untested"}; !slices.Equal(got, want) {
		t.Errorf("%s executions = %q, want %q", runs[1].Text, got, want)
	}
}

func TestStatusesSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	first.mustCreateCase("Login works")
	first.mustCreateCase("Logout works")
	first.mustCreateCase("Reset password works")
	first.mustStartRun("Sprint 1")
	run := first.runs()[0]
	first.mustEnterStatus(run, "Login works", "Failed")
	first.mustEnterStatus(run, "Login works", "Passed")
	first.mustEnterStatus(run, "Logout works", "Failed")
	first.stop()

	second := startApp(t, dbPath)

	want := []string{"Login works Passed", "Logout works Failed", "Reset password works Untested"}
	if got := second.executions(second.runs()[0]); !slices.Equal(got, want) {
		t.Errorf("executions after restart = %q, want %q", got, want)
	}
}

func TestOnlyPassedAndFailedAreAccepted(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]
	a.mustEnterStatus(run, "Login works", "Passed")

	for _, status := range []string{"Untested", "", "passed", "FAILED", " Passed", "Skipped"} {
		for _, title := range []string{"Login works", "Logout works"} {
			if code, _ := a.enterStatus(run, title, status); code != http.StatusBadRequest {
				t.Errorf("enter %q for %q: status %d, want %d", status, title, code, http.StatusBadRequest)
			}
		}
	}

	want := []string{"Login works Passed", "Logout works Untested"}
	if got := a.executions(run); !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestStatusForMissingStatusFieldIsRejected(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]

	code, _ := a.do(http.MethodPost, a.statusAction(run, "Login works"), url.Values{})

	if code != http.StatusBadRequest {
		t.Errorf("status %d, want %d", code, http.StatusBadRequest)
	}
	if got, want := a.executions(run), []string{"Login works Untested"}; !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestStatusCannotBeEnteredOnAnExecutionOutsideTheRun(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	a.mustStartRun("Sprint 1")
	a.mustStartRun("Sprint 2")
	runs := a.runs()
	form := url.Values{"status": {"Passed"}}

	paths := []string{
		"/runs/1/executions/3/status",
		"/runs/2/executions/1/status",
		"/runs/1/executions/99/status",
		"/runs/99/executions/1/status",
		"/runs/1/executions/0/status",
		"/runs/1/executions/-1/status",
		"/runs/abc/executions/1/status",
		"/runs/1/executions/abc/status",
	}
	for _, path := range paths {
		if code, _ := a.do(http.MethodPost, path, form); code != http.StatusNotFound {
			t.Errorf("POST %s: status %d, want %d", path, code, http.StatusNotFound)
		}
	}

	for _, run := range runs {
		want := []string{"Login works Untested", "Logout works Untested"}
		if got := a.executions(run); !slices.Equal(got, want) {
			t.Errorf("%s executions = %q, want %q", run.Text, got, want)
		}
	}
}

func TestRunPageOffersPassedAndFailedForEveryExecution(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")
	a.mustCreateCase("Logout works")
	a.mustStartRun("Sprint 1")
	run := a.runs()[0]
	a.mustEnterStatus(run, "Login works", "Passed")

	for _, title := range []string{"Login works", "Logout works"} {
		if got, want := a.statusChoices(run, title), []string{"Passed", "Failed"}; !slices.Equal(got, want) {
			t.Errorf("%s choices = %q, want %q", title, got, want)
		}
	}
}
