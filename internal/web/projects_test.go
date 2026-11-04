package web_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestCreatedProjectsAppearInListInCreationOrder(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateProject("Mobile")
	a.mustCreateProject(`<b>Web</b> & "API"`)

	want := []string{"Mobile", `<b>Web</b> & "API"`}
	if got := a.projectNames(); !slices.Equal(got, want) {
		t.Errorf("project names = %q, want %q", got, want)
	}
}

func TestProjectHomeShowsTheProjectName(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateProject("Mobile")
	a.mustCreateProject("Web")

	for _, p := range a.projects() {
		status, body := a.do(http.MethodGet, p.Href, nil)
		if status != http.StatusOK {
			t.Fatalf("GET %s: status %d\n%s", p.Href, status, body)
		}
		if got := heading(t, body); got != p.Text {
			t.Errorf("GET %s: heading = %q, want %q", p.Href, got, p.Text)
		}
	}
}

func TestBlankProjectNameIsRejectedWithAnError(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateProject("Mobile")

	for _, name := range []string{"", "   ", "\t\r\n", "\u00a0\u3000", "\u200b\ufeff\u2060"} {
		status, body := a.createProject(name)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("create project %q: status %d, want %d", name, status, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(body, "Name is required.") {
			t.Errorf("create project %q: page does not show the error\n%s", name, body)
		}
	}

	if got, want := a.projectNames(), []string{"Mobile"}; !slices.Equal(got, want) {
		t.Errorf("project names = %q, want %q", got, want)
	}
}

func TestProjectNameIsUniqueIgnoringASCIICase(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateProject("Mobile")

	for _, name := range []string{"Mobile", "mobile", "MOBILE", "mObIlE", " \t mobile\u200b\n"} {
		status, body := a.createProject(name)
		if status != http.StatusUnprocessableEntity {
			t.Errorf("create project %q: status %d, want %d", name, status, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(body, "A project with this name already exists.") {
			t.Errorf("create project %q: page does not show the error\n%s", name, body)
		}
	}

	if got, want := a.projectNames(), []string{"Mobile"}; !slices.Equal(got, want) {
		t.Errorf("project names = %q, want %q", got, want)
	}
}

func TestProjectNamesDifferingOnlyInNonASCIICaseAreDistinct(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateProject("Éclair")
	a.mustCreateProject("éclair")

	if got, want := a.projectNames(), []string{"Éclair", "éclair"}; !slices.Equal(got, want) {
		t.Errorf("project names = %q, want %q", got, want)
	}
}

func TestProjectNameIsTrimmed(t *testing.T) {
	a := startApp(t, tempDB(t))

	a.mustCreateProject(" \t Mobile\u200b\n")

	if got, want := a.projectNames(), []string{"Mobile"}; !slices.Equal(got, want) {
		t.Errorf("project names = %q, want %q", got, want)
	}
}

func TestProjectsSurviveRestart(t *testing.T) {
	dbPath := tempDB(t)
	first := startApp(t, dbPath)
	first.mustCreateProject("Mobile")
	first.mustCreateProject("Web")
	first.stop()

	second := startApp(t, dbPath)

	if got, want := second.projectNames(), []string{"Mobile", "Web"}; !slices.Equal(got, want) {
		t.Errorf("project names after restart = %q, want %q", got, want)
	}
}

func TestUnknownProjectIsNotFound(t *testing.T) {
	a := startApp(t, tempDB(t))
	a.mustCreateProject("Mobile")

	for _, path := range []string{"/projects/2", "/projects/0", "/projects/-1", "/projects/abc", "/projects/1x"} {
		if status, _ := a.do(http.MethodGet, path, nil); status != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want %d", path, status, http.StatusNotFound)
		}
	}
}

func TestRootRedirectsToProjects(t *testing.T) {
	a := startApp(t, tempDB(t))
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	resp, err := client.Get(a.server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		t.Errorf("GET /: status %d, want a redirect", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "/projects" {
		t.Errorf("GET /: Location = %q, want %q", got, "/projects")
	}
}

func TestEveryPageLinksToProjects(t *testing.T) {
	a := startApp(t, tempDB(t))
	p := a.inNewProject("Mobile")
	p.mustCreateCase("Login works")
	p.mustStartRun("Sprint 1")
	run := p.runs()[0]
	_, rejected := a.createProject("")

	pages := map[string]string{"rejected POST /projects": rejected}
	for _, path := range []string{"/projects", p.path, p.path + "/cases", p.path + "/runs", run.Href} {
		status, body := a.do(http.MethodGet, path, nil)
		if status != http.StatusOK {
			t.Fatalf("GET %s: status %d\n%s", path, status, body)
		}
		pages["GET "+path] = body
	}

	want := link{Text: "Projects", Href: "/projects"}
	for name, body := range pages {
		if got := navLinks(t, body, "Main"); !slices.Contains(got, want) {
			t.Errorf("%s: nav links = %v, want to contain %v", name, got, want)
		}
	}
}

func heading(t *testing.T, page string) string {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "h1" {
			return text(n)
		}
	}
	t.Fatalf("no heading\n%s", page)
	return ""
}

func navLinks(t *testing.T, page, label string) []link {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	var links []link
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "nav" && attr(n, "aria-label") == label {
			for a := range n.Descendants() {
				if a.Type == html.ElementNode && a.Data == "a" {
					links = append(links, link{Text: text(a), Href: attr(a, "href")})
				}
			}
		}
	}
	return links
}
