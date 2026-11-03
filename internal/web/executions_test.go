package web_test

import (
	"net/http"
	"net/url"
	"slices"
	"testing"
)

func startRunWithCases(t *testing.T, app *runningApp, runName string, titles ...string) runLink {
	t.Helper()
	for _, title := range titles {
		app.createCase(title)
	}
	app.startRun(runName)
	return app.onlyRun()
}

func TestEnteredStatusesAppearOnTheRunPageAndOtherExecutionsStayUntested(t *testing.T) {
	app := startApp(t, newDBPath(t))
	run := startRunWithCases(t, app, "Sprint 1", "Login works", "Logout works", "Profile works")

	app.pressStatusButton(run, "Login works", "Passed")
	app.pressStatusButton(run, "Logout works", "Failed")

	want := []executionRow{{"Login works", "Passed"}, {"Logout works", "Failed"}, {"Profile works", "Untested"}}
	if got := app.executionsOf(run); !slices.Equal(got, want) {
		t.Fatalf("executions = %v, want %v", got, want)
	}
}

func TestEnteringAStatusOverwritesTheEarlierOne(t *testing.T) {
	app := startApp(t, newDBPath(t))
	run := startRunWithCases(t, app, "Sprint 1", "Login works")

	app.pressStatusButton(run, "Login works", "Failed")
	app.pressStatusButton(run, "Login works", "Passed")

	if got, want := app.executionsOf(run), []executionRow{{"Login works", "Passed"}}; !slices.Equal(got, want) {
		t.Fatalf("executions after Failed then Passed = %v, want %v", got, want)
	}

	app.pressStatusButton(run, "Login works", "Passed")

	if got, want := app.executionsOf(run), []executionRow{{"Login works", "Passed"}}; !slices.Equal(got, want) {
		t.Fatalf("executions after Passed again = %v, want %v", got, want)
	}

	app.pressStatusButton(run, "Login works", "Failed")

	if got, want := app.executionsOf(run), []executionRow{{"Login works", "Failed"}}; !slices.Equal(got, want) {
		t.Fatalf("executions after Passed then Failed = %v, want %v", got, want)
	}
}

func TestStatusEnteredInOneRunDoesNotTouchTheSameCaseInAnotherRun(t *testing.T) {
	app := startApp(t, newDBPath(t))
	app.createCase("Login works")
	app.startRun("Sprint 1")
	app.startRun("Sprint 2")
	links := app.runLinks()

	app.pressStatusButton(links[0], "Login works", "Failed")

	if got, want := app.executionsOf(links[0]), []executionRow{{"Login works", "Failed"}}; !slices.Equal(got, want) {
		t.Fatalf("executions of %q = %v, want %v", links[0].Name, got, want)
	}
	if got, want := app.executionsOf(links[1]), []executionRow{{"Login works", "Untested"}}; !slices.Equal(got, want) {
		t.Fatalf("executions of %q = %v, want %v", links[1].Name, got, want)
	}
}

func TestEnteringAStatusRedirectsToTheRunPageAtThatExecution(t *testing.T) {
	app := startApp(t, newDBPath(t))
	run := startRunWithCases(t, app, "Sprint 1", "Login works", "Logout works")
	path, form := app.statusButtonRequest(run, "Logout works", "Passed")
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	resp, err := client.PostForm(app.server.URL+path, form)
	if err != nil {
		t.Fatalf("post %s: %v", path, err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}
	location, err := resp.Location()
	if err != nil {
		t.Fatalf("location: %v", err)
	}
	if location.Path != run.Path {
		t.Fatalf("redirect path = %q, want the run page %q", location.Path, run.Path)
	}
	item := executionItemIn(t, app.page(run.Path), "Logout works")
	if location.Fragment == "" || location.Fragment != attr(item, "id") {
		t.Fatalf("redirect fragment = %q, want the id %q of the execution", location.Fragment, attr(item, "id"))
	}
}

func TestStatusOutsideTheClosedSetIsRejectedAndNothingChanges(t *testing.T) {
	invalid := map[string]string{
		"empty":     "",
		"untested":  "Untested",
		"lowercase": "passed",
		"unknown":   "Blocked",
		"padded":    " Passed",
	}
	for name, value := range invalid {
		t.Run(name, func(t *testing.T) {
			app := startApp(t, newDBPath(t))
			run := startRunWithCases(t, app, "Sprint 1", "Login works", "Logout works")
			app.pressStatusButton(run, "Login works", "Passed")
			passedPath, _ := app.statusButtonRequest(run, "Login works", "Passed")
			untestedPath, _ := app.statusButtonRequest(run, "Logout works", "Passed")

			for _, target := range []string{passedPath, untestedPath} {
				status, _ := app.postForm(target, url.Values{"status": {value}})

				if status != http.StatusUnprocessableEntity {
					t.Fatalf("status = %d, want %d", status, http.StatusUnprocessableEntity)
				}
			}
			want := []executionRow{{"Login works", "Passed"}, {"Logout works", "Untested"}}
			if got := app.executionsOf(run); !slices.Equal(got, want) {
				t.Fatalf("executions = %v, want %v", got, want)
			}
		})
	}
}

func TestEnteringAStatusOnAnUnknownExecutionIsNotFound(t *testing.T) {
	app := startApp(t, newDBPath(t))
	run := startRunWithCases(t, app, "Sprint 1", "Login works")

	for _, path := range []string{"/executions/999/status", "/executions/0/status", "/executions/-1/status", "/executions/login/status"} {
		status, _ := app.postForm(path, url.Values{"status": {"Passed"}})

		if status != http.StatusNotFound {
			t.Errorf("POST %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}
	if got, want := app.executionsOf(run), []executionRow{{"Login works", "Untested"}}; !slices.Equal(got, want) {
		t.Fatalf("executions = %v, want %v", got, want)
	}
}

func TestStatusesSurviveRestartOnTheSameDatabaseFile(t *testing.T) {
	dbPath := newDBPath(t)
	first := startApp(t, dbPath)
	run := startRunWithCases(t, first, "Sprint 1", "Login works", "Logout works", "Profile works")
	first.pressStatusButton(run, "Login works", "Passed")
	first.pressStatusButton(run, "Logout works", "Failed")
	first.pressStatusButton(run, "Logout works", "Passed")
	first.pressStatusButton(run, "Logout works", "Failed")
	before := first.executionsOf(run)
	first.stop()

	second := startApp(t, dbPath)

	want := []executionRow{{"Login works", "Passed"}, {"Logout works", "Failed"}, {"Profile works", "Untested"}}
	got := second.executionsOf(run)
	if !slices.Equal(got, before) || !slices.Equal(got, want) {
		t.Fatalf("executions after restart = %v, want %v (before the restart: %v)", got, want, before)
	}
}
