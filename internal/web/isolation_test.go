package web_test

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestRunsStayInTheirOwnProject(t *testing.T) {
	a := startApp(t, tempDB(t))
	mobile := a.inNewProject("Mobile")
	desktop := a.inNewProject("Desktop")
	a.mustCreateCase("Login works")

	mobile.mustStartRun("Sprint 1")
	desktop.mustStartRun("Sprint 2")

	if got, want := mobile.runNames(), []string{"Sprint 1"}; !slices.Equal(got, want) {
		t.Errorf("Mobile run names = %q, want %q", got, want)
	}
	if got, want := desktop.runNames(), []string{"Sprint 2"}; !slices.Equal(got, want) {
		t.Errorf("Desktop run names = %q, want %q", got, want)
	}
}

func TestRunsOfEveryProjectHaveExecutionsForEveryCaseOfTheLibrary(t *testing.T) {
	a := startApp(t, tempDB(t))
	mobile := a.inNewProject("Mobile")
	desktop := a.inNewProject("Desktop")
	a.mustCreateCase("Checkout works")
	a.mustCreateCase("Login works")

	mobile.mustStartRun("Sprint 1")
	desktop.mustStartRun("Sprint 2")

	want := []string{"Checkout works Untested", "Login works Untested"}
	if got := mobile.executions(mobile.runs()[0]); !slices.Equal(got, want) {
		t.Errorf("Mobile executions = %q, want %q", got, want)
	}
	if got := desktop.executions(desktop.runs()[0]); !slices.Equal(got, want) {
		t.Errorf("Desktop executions = %q, want %q", got, want)
	}
}

func TestTwoProjectsMayShareRunNames(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateCase("Login works")

	for _, p := range []project{a.inNewProject("Mobile"), a.inNewProject("Desktop")} {
		p.mustStartRun("Sprint 1")

		if got, want := p.runNames(), []string{"Sprint 1"}; !slices.Equal(got, want) {
			t.Errorf("%s run names = %q, want %q", p.path, got, want)
		}
	}
}

func TestRunCanStartInAProjectWhoseLibraryCasesWereCreatedFromAnywhere(t *testing.T) {
	a := startApp(t, tempDB(t))
	mobile := a.inNewProject("Mobile")
	desktop := a.inNewProject("Desktop")
	mobile.mustCreateCase("Checkout works")

	desktop.mustStartRun("Sprint 1")

	if got, want := desktop.executions(desktop.runs()[0]), []string{"Checkout works Untested"}; !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestRunOpenedThroughAnotherProjectIsNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))
	mobile := a.inNewProject("Mobile")
	desktop := a.inNewProject("Desktop")
	mobile.mustCreateCase("Login works")
	mobile.mustStartRun("Sprint 1")
	run := mobile.runs()[0]

	path := strings.Replace(run.Href, mobile.path, desktop.path, 1)

	if status, _ := a.do(http.MethodGet, path, nil); status != http.StatusNotFound {
		t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
	}
	if status, _ := a.do(http.MethodGet, run.Href, nil); status != http.StatusOK {
		t.Errorf("GET %s: status %d, want %d", run.Href, status, http.StatusOK)
	}
}

func TestStatusCannotBeEnteredThroughAnotherProject(t *testing.T) {
	a := startApp(t, tempDB(t))
	mobile := a.inNewProject("Mobile")
	desktop := a.inNewProject("Desktop")
	a.mustCreateCase("Login works")
	a.mustCreateCase("Checkout works")
	mobile.mustStartRun("Sprint 1")
	desktop.mustStartRun("Sprint 2")

	paths := []string{
		"/projects/2/runs/1/executions/1/status",
		"/projects/2/runs/2/executions/1/status",
		"/projects/1/runs/2/executions/3/status",
		"/projects/1/runs/1/executions/3/status",
		"/projects/3/runs/1/executions/1/status",
	}
	for _, path := range paths {
		for _, status := range []string{"Passed", "Failed", "Skipped", ""} {
			if code, _ := a.do(http.MethodPost, path, url.Values{"status": {status}}); code != http.StatusNotFound {
				t.Errorf("POST %s with %q: status %d, want %d", path, status, code, http.StatusNotFound)
			}
		}
	}

	want := []string{"Login works Untested", "Checkout works Untested"}
	if got := mobile.executions(mobile.runs()[0]); !slices.Equal(got, want) {
		t.Errorf("Mobile executions = %q, want %q", got, want)
	}
	if got := desktop.executions(desktop.runs()[0]); !slices.Equal(got, want) {
		t.Errorf("Desktop executions = %q, want %q", got, want)
	}
}

func TestProjectCasesAddressesAreNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))
	p := a.inNewProject("Mobile")
	a.mustCreateCase("Login works")

	for _, path := range []string{p.path + "/cases", p.path + "/cases/", p.path + "/cases/1", "/projects/99/cases"} {
		if status, _ := a.do(http.MethodGet, path, nil); status != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
		}
		if status, _ := a.do(http.MethodPost, path, url.Values{"title": {"Logout works"}}); status != http.StatusNotFound {
			t.Errorf("POST %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}

	if got, want := a.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Errorf("case titles = %q, want %q", got, want)
	}
}

func TestSubpagesOfAnUnknownProjectAreNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))
	p := a.inNewProject("Mobile")
	p.mustCreateCase("Login works")
	p.mustStartRun("Sprint 1")

	for _, id := range []string{"2", "0", "-1", "abc", "1x", "99999999999999999999"} {
		base := "/projects/" + id
		for _, path := range []string{"/runs", "/runs/1"} {
			if status, _ := a.do(http.MethodGet, base+path, nil); status != http.StatusNotFound {
				t.Errorf("GET %s%s: status %d, want %d", base, path, status, http.StatusNotFound)
			}
		}
		runForm := url.Values{"name": {"Sprint 2"}}
		if status, _ := a.do(http.MethodPost, base+"/runs", runForm); status != http.StatusNotFound {
			t.Errorf("POST %s/runs: status %d, want %d", base, status, http.StatusNotFound)
		}
		if status, _ := a.do(http.MethodPost, base+"/runs", url.Values{}); status != http.StatusNotFound {
			t.Errorf("POST %s/runs with a blank form: status %d, want %d", base, status, http.StatusNotFound)
		}
	}

	if got, want := p.runNames(), []string{"Sprint 1"}; !slices.Equal(got, want) {
		t.Errorf("run names = %q, want %q", got, want)
	}
}

func TestWritesWithoutAnyProjectAreNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))

	if status, _ := a.do(http.MethodPost, "/projects/1/runs", url.Values{"name": {"Sprint 1"}}); status != http.StatusNotFound {
		t.Errorf("start run: status %d, want %d", status, http.StatusNotFound)
	}
}

func TestOldAddressesAreNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))
	p := a.inNewProject("Mobile")
	p.mustCreateCase("Login works")
	p.mustStartRun("Sprint 1")

	for _, path := range []string{"/cases", "/runs", "/runs/1"} {
		if status, _ := a.do(http.MethodGet, path, nil); status != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}
	writes := map[string]url.Values{
		"/cases":                      {"title": {"Logout works"}},
		"/runs":                       {"name": {"Sprint 2"}},
		"/runs/1/executions/1/status": {"status": {"Passed"}},
	}
	for path, form := range writes {
		if status, _ := a.do(http.MethodPost, path, form); status != http.StatusNotFound {
			t.Errorf("POST %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}

	if got, want := p.executions(p.runs()[0]), []string{"Login works Untested"}; !slices.Equal(got, want) {
		t.Errorf("executions = %q, want %q", got, want)
	}
}

func TestProjectHomeLinksOnlyToItsRuns(t *testing.T) {
	a := startApp(t, tempDB(t))

	for _, p := range []project{a.inNewProject("Mobile"), a.inNewProject("Desktop")} {
		status, body := a.do(http.MethodGet, p.path, nil)
		if status != http.StatusOK {
			t.Fatalf("GET %s: status %d\n%s", p.path, status, body)
		}

		want := []link{{Text: "Runs", Href: p.path + "/runs"}}
		got := navLinks(t, body, "Project")
		if !slices.Equal(got, want) {
			t.Fatalf("GET %s: links = %v, want %v", p.path, got, want)
		}
		status, body = a.do(http.MethodGet, got[0].Href, nil)
		if status != http.StatusOK {
			t.Fatalf("GET %s: status %d\n%s", got[0].Href, status, body)
		}
		if heading := heading(t, body); heading != got[0].Text {
			t.Errorf("GET %s: heading = %q, want %q", got[0].Href, heading, got[0].Text)
		}
	}
}

func TestProjectSubpagesShowProjectsLibraryAndTheProjectNameInTheMenu(t *testing.T) {
	a := startApp(t, tempDB(t))
	p := a.inNewProject("Mobile")
	p.mustCreateCase("Login works")
	p.mustStartRun("Sprint 1")
	_, rejectedRun := p.startRun("")

	pages := map[string]string{"rejected run": rejectedRun}
	for _, path := range []string{p.path + "/runs", p.runs()[0].Href} {
		status, body := a.do(http.MethodGet, path, nil)
		if status != http.StatusOK {
			t.Fatalf("GET %s: status %d\n%s", path, status, body)
		}
		pages["GET "+path] = body
	}

	want := []link{{Text: "Projects", Href: "/projects"}, {Text: "Library", Href: "/library"}, {Text: "Mobile", Href: p.path}}
	for name, body := range pages {
		if got := navLinks(t, body, "Main"); !slices.Equal(got, want) {
			t.Errorf("%s: menu = %v, want %v", name, got, want)
		}
		if got := navLinks(t, body, "Project"); len(got) != 0 {
			t.Errorf("%s: project links = %v, want none", name, got)
		}
	}
}

func TestProjectListHomeAndLibraryDoNotShowTheProjectNameInTheMenu(t *testing.T) {
	a := startApp(t, tempDB(t))
	p := a.inNewProject("Mobile")
	_, rejectedCase := a.createCase("")

	pages := map[string]string{"rejected case": rejectedCase}
	for _, path := range []string{"/projects", p.path, "/library"} {
		_, body := a.do(http.MethodGet, path, nil)
		pages["GET "+path] = body
	}

	want := []link{{Text: "Projects", Href: "/projects"}, {Text: "Library", Href: "/library"}}
	for name, body := range pages {
		if got := navLinks(t, body, "Main"); !slices.Equal(got, want) {
			t.Errorf("%s: menu = %v, want %v", name, got, want)
		}
	}
}
