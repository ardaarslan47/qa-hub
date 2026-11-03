package web_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func rows(pairs ...string) []executionRow {
	var result []executionRow
	for i := 0; i < len(pairs); i += 2 {
		result = append(result, executionRow{Title: pairs[i], Status: pairs[i+1]})
	}
	return result
}

func startRunOverTwoCases(t *testing.T, a *app) (path string, actions []string) {
	t.Helper()
	a.createCase("Login works")
	a.createCase("Logout works")
	a.startRun("Sprint 1")
	path = a.onlyRun().Path
	actions = a.statusActions(path)
	if len(actions) != 2 {
		t.Fatalf("status forms on %s = %v, want one per execution", path, actions)
	}
	return path, actions
}

func TestEachExecutionOffersPassedAndFailed(t *testing.T) {
	a := startApp(t, tempDB(t))
	path, _ := startRunOverTwoCases(t, a)

	got := elementTexts(a.get(path).Body, "button")

	if want := []string{"Passed", "Failed", "Passed", "Failed"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("buttons = %q, want %q", got, want)
	}
}

func TestRecordedStatusesAreShownAndStayAfterReload(t *testing.T) {
	a := startApp(t, tempDB(t))
	path, actions := startRunOverTwoCases(t, a)

	a.recordStatus(actions[0], "Passed")
	resp := a.recordStatus(actions[1], "Failed")

	want := rows("Login works", "Passed", "Logout works", "Failed")
	if resp.Status != http.StatusOK {
		t.Fatalf("record status = %d, want 200", resp.Status)
	}
	if got := executionRows(resp.Body); !reflect.DeepEqual(got, want) {
		t.Fatalf("executions after recording = %v, want %v", got, want)
	}
	if got := a.executionRows(path); !reflect.DeepEqual(got, want) {
		t.Fatalf("executions after reload = %v, want %v", got, want)
	}
}

func TestNewStatusReplacesTheOldOne(t *testing.T) {
	a := startApp(t, tempDB(t))
	path, actions := startRunOverTwoCases(t, a)

	for _, step := range []struct{ status, want string }{
		{"Failed", "Failed"},
		{"Passed", "Passed"},
		{"Passed", "Passed"},
		{"Failed", "Failed"},
	} {
		a.recordStatus(actions[0], step.status)
		if got, want := a.executionRows(path), rows("Login works", step.want, "Logout works", "Untested"); !reflect.DeepEqual(got, want) {
			t.Fatalf("after recording %s: executions = %v, want %v", step.status, got, want)
		}
	}
}

func TestStatusBelongsToOneExecutionOfOneRun(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.createCase("Login works")
	a.startRun("Sprint 1")
	a.startRun("Sprint 2")
	links := a.runLinks()

	a.recordStatus(a.statusActions(links[0].Path)[0], "Failed")

	if got, want := a.executionRows(links[0].Path), rows("Login works", "Failed"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Sprint 1 executions = %v, want %v", got, want)
	}
	if got, want := a.executionRows(links[1].Path), rows("Login works", "Untested"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Sprint 2 executions = %v, want %v", got, want)
	}
}

func TestStatusCannotBeAnythingButPassedOrFailed(t *testing.T) {
	for _, status := range []string{"Untested", "", "passed", "Blocked"} {
		t.Run("rejects "+status, func(t *testing.T) {
			a := startApp(t, tempDB(t))
			path, actions := startRunOverTwoCases(t, a)
			a.recordStatus(actions[0], "Passed")

			onUntested := a.recordStatus(actions[1], status)
			onPassed := a.recordStatus(actions[0], status)

			if onUntested.Status != http.StatusBadRequest || onPassed.Status != http.StatusBadRequest {
				t.Fatalf("responses = %d and %d, want 400 for status %q", onUntested.Status, onPassed.Status, status)
			}
			if got, want := a.executionRows(path), rows("Login works", "Passed", "Logout works", "Untested"); !reflect.DeepEqual(got, want) {
				t.Fatalf("executions = %v, want %v", got, want)
			}
		})
	}
}

func TestStatusOfAnUnknownExecutionIsNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.createCase("Login works")
	a.startRun("Sprint 1")
	a.startRun("Sprint 2")
	links := a.runLinks()
	executionOfFirstRun := a.statusActions(links[0].Path)[0]
	inSecondRun := strings.Replace(executionOfFirstRun, links[0].Path, links[1].Path, 1)

	for _, action := range []string{inSecondRun, links[0].Path + "/executions/999/status", links[0].Path + "/executions/abc/status", "/runs/999/executions/1/status"} {
		if resp := a.recordStatus(action, "Passed"); resp.Status != http.StatusNotFound {
			t.Errorf("POST %s = %d, want 404", action, resp.Status)
		}
	}
	for _, link := range links {
		if got, want := a.executionRows(link.Path), rows("Login works", "Untested"); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s executions = %v, want %v", link.Path, got, want)
		}
	}
}

func TestStatusesSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	path, actions := startRunOverTwoCases(t, first)
	first.recordStatus(actions[0], "Failed")
	first.recordStatus(actions[0], "Passed")
	first.recordStatus(actions[1], "Failed")
	first.stop()

	second := startApp(t, dbPath)

	if got, want := second.executionRows(path), rows("Login works", "Passed", "Logout works", "Failed"); !reflect.DeepEqual(got, want) {
		t.Fatalf("executions after restart = %v, want %v", got, want)
	}
}
