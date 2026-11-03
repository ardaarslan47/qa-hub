package web_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func untested(titles ...string) []executionRow {
	var rows []executionRow
	for _, title := range titles {
		rows = append(rows, executionRow{Title: title, Status: "Untested"})
	}
	return rows
}

func TestStartedRunIsListedAndItsPageHasAnUntestedExecutionPerCase(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.createCase("Login works")
	a.createCase("Logout works")

	resp := a.startRun("Sprint 1")

	if resp.Status != http.StatusOK {
		t.Fatalf("start run = %d, want 200", resp.Status)
	}
	run := a.onlyRun()
	if run.Name != "Sprint 1" {
		t.Fatalf("run name = %q, want Sprint 1", run.Name)
	}
	page := a.get(run.Path)
	if got := elementTexts(page.Body, "h1"); !reflect.DeepEqual(got, []string{"Sprint 1"}) {
		t.Fatalf("run page heading = %q", got)
	}
	if got, want := executionRows(page.Body), untested("Login works", "Logout works"); !reflect.DeepEqual(got, want) {
		t.Fatalf("executions = %v, want %v", got, want)
	}
}

func TestBlankRunNameIsRejectedAndNothingIsStarted(t *testing.T) {
	for name, runName := range map[string]string{"empty": "", "spaces": "   ", "whitespace": " \t\n "} {
		t.Run(name, func(t *testing.T) {
			a := startApp(t, tempDB(t))
			a.createCase("Login works")

			resp := a.startRun(runName)

			if resp.Status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", resp.Status)
			}
			if !strings.Contains(resp.Body, "Name is required.") {
				t.Fatalf("body lacks the error message: %s", resp.Body)
			}
			if got := a.runLinks(); len(got) != 0 {
				t.Fatalf("runs = %v, want none", got)
			}
		})
	}
}

func TestRunCannotStartWithoutCases(t *testing.T) {
	a := startApp(t, tempDB(t))

	resp := a.startRun("Sprint 1")

	if resp.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.Status)
	}
	if !strings.Contains(resp.Body, "Create a case before starting a run.") {
		t.Fatalf("body lacks the error message: %s", resp.Body)
	}
	if got := a.runLinks(); len(got) != 0 {
		t.Fatalf("a rejected run was left behind: %v", got)
	}

	a.createCase("Login works")
	if resp := a.startRun("Sprint 1"); resp.Status != http.StatusOK {
		t.Fatalf("start run after creating a case = %d, want 200", resp.Status)
	}
	if got := a.runNames(); !reflect.DeepEqual(got, []string{"Sprint 1"}) {
		t.Fatalf("runs = %q, want one Sprint 1", got)
	}
}

func TestRunNameIsTrimmed(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.createCase("Login works")

	a.startRun("  Sprint 1 \n")

	if got := a.runNames(); !reflect.DeepEqual(got, []string{"Sprint 1"}) {
		t.Fatalf("runs = %q", got)
	}
}

func TestTwoRunsMayShareAName(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.createCase("Login works")

	a.startRun("Sprint 1")
	resp := a.startRun("Sprint 1")

	if resp.Status != http.StatusOK {
		t.Fatalf("second start = %d, want 200", resp.Status)
	}
	links := a.runLinks()
	if len(links) != 2 || links[0].Name != "Sprint 1" || links[1].Name != "Sprint 1" || links[0].Path == links[1].Path {
		t.Fatalf("runs = %v, want two distinct runs named Sprint 1", links)
	}
}

func TestRunIsASnapshotOfTheCasesThatExistedWhenItStarted(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.createCase("Login works")
	a.startRun("Sprint 1")
	a.createCase("Logout works")
	a.startRun("Sprint 2")

	links := a.runLinks()
	if len(links) != 2 {
		t.Fatalf("runs = %v, want two", links)
	}
	if got, want := a.executionRows(links[0].Path), untested("Login works"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Sprint 1 executions = %v, want %v", got, want)
	}
	if got, want := a.executionRows(links[1].Path), untested("Login works", "Logout works"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Sprint 2 executions = %v, want %v", got, want)
	}
}

func TestEveryCaseGetsItsOwnExecutionEvenWithTheSameTitle(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.createCase("Checkout")
	a.createCase("Checkout")

	a.startRun("Sprint 1")

	if got, want := a.executionRows(a.onlyRun().Path), untested("Checkout", "Checkout"); !reflect.DeepEqual(got, want) {
		t.Fatalf("executions = %v, want %v", got, want)
	}
}

func TestUnknownRunIsNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))

	for _, path := range []string{"/runs/1", "/runs/999", "/runs/abc"} {
		if resp := a.get(path); resp.Status != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, resp.Status)
		}
	}
}

func TestRunsAndExecutionsSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	first.createCase("Login works")
	first.createCase("Logout works")
	first.startRun("Sprint 1")
	first.startRun("Sprint 1")
	first.createCase("Password reset")
	first.startRun("Sprint 2")
	first.stop()

	second := startApp(t, dbPath)

	links := second.runLinks()
	if len(links) != 3 {
		t.Fatalf("runs after restart = %v, want three", links)
	}
	if got := second.runNames(); !reflect.DeepEqual(got, []string{"Sprint 1", "Sprint 1", "Sprint 2"}) {
		t.Fatalf("run names after restart = %q", got)
	}
	for _, link := range links[:2] {
		if got, want := second.executionRows(link.Path), untested("Login works", "Logout works"); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s executions after restart = %v, want %v", link.Path, got, want)
		}
	}
	if got, want := second.executionRows(links[2].Path), untested("Login works", "Logout works", "Password reset"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Sprint 2 executions after restart = %v, want %v", got, want)
	}
}
