package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
	"github.com/ardaarslan47/qa-hub/internal/web"
)

type app struct {
	t      *testing.T
	store  *qahub.Store
	server *httptest.Server
}

func tempDB(t *testing.T) string {
	return filepath.Join(t.TempDir(), "qa-hub.db")
}

func startApp(t *testing.T, dbPath string) *app {
	t.Helper()
	store, err := qahub.Open(dbPath)
	if err != nil {
		t.Fatalf("open %s: %v", dbPath, err)
	}
	a := &app{t: t, store: store, server: httptest.NewServer(web.New(store))}
	t.Cleanup(a.stop)
	return a
}

func (a *app) stop() {
	a.server.Close()
	if err := a.store.Close(); err != nil {
		a.t.Errorf("close store: %v", err)
	}
}

func (a *app) do(method, path string, form url.Values) (int, string) {
	a.t.Helper()
	req, err := http.NewRequest(method, a.server.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		a.t.Fatal(err)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := a.server.Client().Do(req)
	if err != nil {
		a.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.t.Fatal(err)
	}
	return resp.StatusCode, string(body)
}

func (a *app) createProject(name string) (int, string) {
	a.t.Helper()
	return a.do(http.MethodPost, "/projects", url.Values{"name": {name}})
}

func (a *app) mustCreateProject(name string) {
	a.t.Helper()
	if status, body := a.createProject(name); status != http.StatusOK {
		a.t.Fatalf("create project %q: status %d\n%s", name, status, body)
	}
}

func (a *app) projects() []link {
	a.t.Helper()
	status, body := a.do(http.MethodGet, "/projects", nil)
	if status != http.StatusOK {
		a.t.Fatalf("GET /projects: status %d\n%s", status, body)
	}
	return listLinks(a.t, body, "Projects")
}

func (a *app) projectNames() []string {
	a.t.Helper()
	var names []string
	for _, p := range a.projects() {
		names = append(names, p.Text)
	}
	return names
}

type project struct {
	*app
	path string
}

func (a *app) inProject(name string) project {
	a.t.Helper()
	for _, p := range a.projects() {
		if p.Text == name {
			return project{app: a, path: p.Href}
		}
	}
	a.t.Fatalf("no project named %q", name)
	return project{}
}

func (a *app) inNewProject(name string) project {
	a.t.Helper()
	a.mustCreateProject(name)
	return a.inProject(name)
}

func (p project) createCase(title string) (int, string) {
	p.t.Helper()
	return p.do(http.MethodPost, p.path+"/cases", url.Values{"title": {title}})
}

func (p project) mustCreateCase(title string) {
	p.t.Helper()
	if status, body := p.createCase(title); status != http.StatusOK {
		p.t.Fatalf("create case %q: status %d\n%s", title, status, body)
	}
}

func (p project) caseTitles() []string {
	p.t.Helper()
	status, body := p.do(http.MethodGet, p.path+"/cases", nil)
	if status != http.StatusOK {
		p.t.Fatalf("GET %s/cases: status %d\n%s", p.path, status, body)
	}
	return listItems(p.t, body, "Cases")
}

type link struct {
	Text string
	Href string
}

func (p project) startRun(name string) (int, string) {
	p.t.Helper()
	return p.do(http.MethodPost, p.path+"/runs", url.Values{"name": {name}})
}

func (p project) mustStartRun(name string) {
	p.t.Helper()
	if status, body := p.startRun(name); status != http.StatusOK {
		p.t.Fatalf("start run %q: status %d\n%s", name, status, body)
	}
}

func (p project) runs() []link {
	p.t.Helper()
	status, body := p.do(http.MethodGet, p.path+"/runs", nil)
	if status != http.StatusOK {
		p.t.Fatalf("GET %s/runs: status %d\n%s", p.path, status, body)
	}
	return listLinks(p.t, body, "Runs")
}

func (p project) runNames() []string {
	p.t.Helper()
	var names []string
	for _, r := range p.runs() {
		names = append(names, r.Text)
	}
	return names
}

func (a *app) executions(run link) []string {
	a.t.Helper()
	status, body := a.do(http.MethodGet, run.Href, nil)
	if status != http.StatusOK {
		a.t.Fatalf("GET %s: status %d\n%s", run.Href, status, body)
	}
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		a.t.Fatal(err)
	}
	var summaries []string
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && attr(n, "data-testid") == "execution" {
			summaries = append(summaries, text(n))
		}
	}
	return summaries
}

func (a *app) enterStatus(run link, title, status string) (int, string) {
	a.t.Helper()
	return a.do(http.MethodPost, a.statusAction(run, title), url.Values{"status": {status}})
}

func (a *app) mustEnterStatus(run link, title, status string) {
	a.t.Helper()
	if code, body := a.enterStatus(run, title, status); code != http.StatusOK {
		a.t.Fatalf("enter %s for %q: status %d\n%s", status, title, code, body)
	}
}

func (a *app) statusAction(run link, title string) string {
	a.t.Helper()
	return attr(a.statusForm(run, title), "action")
}

func (a *app) statusChoices(run link, title string) []string {
	a.t.Helper()
	var choices []string
	for n := range a.statusForm(run, title).Descendants() {
		if n.Type == html.ElementNode && n.Data == "button" && attr(n, "type") == "submit" && attr(n, "name") == "status" {
			choices = append(choices, attr(n, "value"))
		}
	}
	return choices
}

func (a *app) statusForm(run link, title string) *html.Node {
	a.t.Helper()
	code, body := a.do(http.MethodGet, run.Href, nil)
	if code != http.StatusOK {
		a.t.Fatalf("GET %s: status %d\n%s", run.Href, code, body)
	}
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		a.t.Fatal(err)
	}
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode || n.Data != "ul" || attr(n, "aria-label") != "Executions" {
			continue
		}
		for li := range n.ChildNodes() {
			if li.Type != html.ElementNode || li.Data != "li" {
				continue
			}
			var form *html.Node
			var hasTitle bool
			for d := range li.Descendants() {
				if d.Type != html.ElementNode {
					continue
				}
				hasTitle = hasTitle || d.Data == "span" && text(d) == title
				if d.Data == "form" {
					form = d
				}
			}
			if hasTitle && form != nil {
				return form
			}
		}
	}
	a.t.Fatalf("no status form for %q on %s\n%s", title, run.Href, body)
	return nil
}

func listLinks(t *testing.T, page, label string) []link {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	var links []link
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "ul" && attr(n, "aria-label") == label {
			for a := range n.Descendants() {
				if a.Type == html.ElementNode && a.Data == "a" {
					links = append(links, link{Text: text(a), Href: attr(a, "href")})
				}
			}
		}
	}
	return links
}

func listItems(t *testing.T, page, label string) []string {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	var items []string
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "ul" && attr(n, "aria-label") == label {
			for li := range n.ChildNodes() {
				if li.Type == html.ElementNode && li.Data == "li" {
					items = append(items, text(li))
				}
			}
		}
	}
	return items
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func text(n *html.Node) string {
	var b strings.Builder
	for d := range n.Descendants() {
		if d.Type == html.TextNode {
			b.WriteString(d.Data)
		}
	}
	return strings.TrimSpace(b.String())
}
