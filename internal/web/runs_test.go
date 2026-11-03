package web_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestStartedRunAppearsInListAndItsPageHasAnUntestedExecutionPerCase(t *testing.T) {
	app := startApp(t, newDBPath(t))
	app.createCase("Login works")
	app.createCase("Logout works")

	status, _ := app.startRun("Sprint 1")

	if status != http.StatusOK {
		t.Fatalf("start run: status %d, want %d", status, http.StatusOK)
	}
	if got, want := app.runNames(), []string{"Sprint 1"}; !slices.Equal(got, want) {
		t.Fatalf("runs = %q, want %q", got, want)
	}
	want := []executionRow{{"Login works", "Untested"}, {"Logout works", "Untested"}}
	if got := app.executionsOf(app.onlyRun()); !slices.Equal(got, want) {
		t.Fatalf("executions = %v, want %v", got, want)
	}
}

func TestBlankRunNameIsRejectedWithErrorMessageAndNothingIsStored(t *testing.T) {
	blankNames := map[string]string{
		"empty":                           "",
		"spaces":                          "   ",
		"tab and newline":                 "\t \n",
		"zero width space":                "\u200b",
		"zero width joiner":               "\u200d",
		"word joiner":                     "\u2060",
		"byte order mark":                 "\ufeff",
		"soft hyphen":                     "\u00ad",
		"right-to-left mark":              "\u200f",
		"nul":                             "\x00",
		"zero width space between spaces": " \u200b ",
	}
	for name, blank := range blankNames {
		t.Run(name, func(t *testing.T) {
			app := startApp(t, newDBPath(t))
			app.createCase("Login works")

			status, body := app.startRun(blank)

			if status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want %d", status, http.StatusUnprocessableEntity)
			}
			if got := alertTextIn(t, body); !strings.Contains(got, "Name is required") {
				t.Fatalf("error message = %q, want it to contain %q", got, "Name is required")
			}
			if got := app.runNames(); len(got) != 0 {
				t.Fatalf("runs = %q, want none", got)
			}
		})
	}
}

func TestStartingARunWithoutCasesIsRejectedWithErrorMessageAndNothingIsStored(t *testing.T) {
	app := startApp(t, newDBPath(t))

	status, body := app.startRun("Sprint 1")

	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", status, http.StatusUnprocessableEntity)
	}
	if got := alertTextIn(t, body); !strings.Contains(got, "Create at least one case") {
		t.Fatalf("error message = %q, want it to contain %q", got, "Create at least one case")
	}
	if got := app.runNames(); len(got) != 0 {
		t.Fatalf("runs = %q, want none", got)
	}

	app.createCase("Login works")
	app.startRun("Sprint 1")

	want := []executionRow{{"Login works", "Untested"}}
	if got := app.executionsOf(app.onlyRun()); !slices.Equal(got, want) {
		t.Fatalf("executions after creating a case and starting again = %v, want %v", got, want)
	}
}

func TestTwoRunsWithTheSameNameBothExist(t *testing.T) {
	app := startApp(t, newDBPath(t))
	app.createCase("Login works")

	app.startRun("Sprint 1")
	app.startRun("Sprint 1")

	if got, want := app.runNames(), []string{"Sprint 1", "Sprint 1"}; !slices.Equal(got, want) {
		t.Fatalf("runs = %q, want %q", got, want)
	}
	links := app.runLinks()
	if links[0].Path == links[1].Path {
		t.Fatalf("both runs link to %q, want distinct run pages", links[0].Path)
	}
	want := []executionRow{{"Login works", "Untested"}}
	for _, link := range links {
		if got := app.executionsOf(link); !slices.Equal(got, want) {
			t.Fatalf("executions of %s = %v, want %v", link.Path, got, want)
		}
	}
}

func TestCaseCreatedAfterRunStartIsNotInThatRunButIsInANewRun(t *testing.T) {
	app := startApp(t, newDBPath(t))
	app.createCase("Login works")
	app.startRun("Sprint 1")
	app.createCase("Logout works")
	app.startRun("Sprint 2")

	links := app.runLinks()
	if len(links) != 2 {
		t.Fatalf("runs = %v, want two", links)
	}
	first := []executionRow{{"Login works", "Untested"}}
	if got := app.executionsOf(links[0]); !slices.Equal(got, first) {
		t.Fatalf("executions of %q = %v, want %v", links[0].Name, got, first)
	}
	second := []executionRow{{"Login works", "Untested"}, {"Logout works", "Untested"}}
	if got := app.executionsOf(links[1]); !slices.Equal(got, second) {
		t.Fatalf("executions of %q = %v, want %v", links[1].Name, got, second)
	}
}

func TestRunsAndExecutionsSurviveRestartOnTheSameDatabaseFile(t *testing.T) {
	dbPath := newDBPath(t)
	first := startApp(t, dbPath)
	first.createCase("Login works")
	first.createCase("Logout works")
	first.startRun("Sprint 1")
	runsBefore := first.runLinks()
	executionsBefore := first.executionsOf(runsBefore[0])
	first.stop()

	second := startApp(t, dbPath)

	if got := second.runLinks(); !slices.Equal(got, runsBefore) {
		t.Fatalf("runs after restart = %v, want %v", got, runsBefore)
	}
	got := second.executionsOf(runsBefore[0])
	if !slices.Equal(got, executionsBefore) || len(got) != 2 {
		t.Fatalf("executions after restart = %v, want the %v from before", got, executionsBefore)
	}
}

func TestUnknownRunIsNotFound(t *testing.T) {
	app := startApp(t, newDBPath(t))
	app.createCase("Login works")
	app.startRun("Sprint 1")

	for _, path := range []string{"/runs/999", "/runs/0", "/runs/-1", "/runs/sprint-1"} {
		if status, _ := app.get(path); status != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}
}
