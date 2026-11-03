package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/ardaarslan47/qa-hub/internal/store"
	"github.com/ardaarslan47/qa-hub/internal/web"
)

type runningApp struct {
	t      *testing.T
	server *httptest.Server
	store  *store.Store
}

func startApp(t *testing.T, dbPath string) *runningApp {
	t.Helper()
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	app := &runningApp{t: t, server: httptest.NewServer(web.New(st)), store: st}
	t.Cleanup(app.stop)
	return app
}

func (a *runningApp) stop() {
	if a.server == nil {
		return
	}
	a.server.Close()
	if err := a.store.Close(); err != nil {
		a.t.Errorf("close store: %v", err)
	}
	a.server = nil
}

func (a *runningApp) createCase(title string) (status int, body string) {
	a.t.Helper()
	resp, err := http.PostForm(a.server.URL+"/cases", url.Values{"title": {title}})
	if err != nil {
		a.t.Fatalf("post case: %v", err)
	}
	return resp.StatusCode, readBody(a.t, resp)
}

func (a *runningApp) caseTitles() []string {
	a.t.Helper()
	return caseTitlesIn(a.t, a.page("/cases"))
}

type runLink struct {
	Name string
	Path string
}

type executionRow struct {
	Title string
	Label string
}

func (a *runningApp) startRun(name string) (status int, body string) {
	a.t.Helper()
	resp, err := http.PostForm(a.server.URL+"/runs", url.Values{"name": {name}})
	if err != nil {
		a.t.Fatalf("post run: %v", err)
	}
	return resp.StatusCode, readBody(a.t, resp)
}

func (a *runningApp) get(path string) (status int, body string) {
	a.t.Helper()
	resp, err := http.Get(a.server.URL + path)
	if err != nil {
		a.t.Fatalf("get %s: %v", path, err)
	}
	return resp.StatusCode, readBody(a.t, resp)
}

func (a *runningApp) page(path string) string {
	a.t.Helper()
	status, body := a.get(path)
	if status != http.StatusOK {
		a.t.Fatalf("get %s: status %d", path, status)
	}
	return body
}

func (a *runningApp) runLinks() []runLink {
	a.t.Helper()
	return runLinksIn(a.t, a.page("/runs"))
}

func (a *runningApp) runNames() []string {
	a.t.Helper()
	names := []string{}
	for _, link := range a.runLinks() {
		names = append(names, link.Name)
	}
	return names
}

func (a *runningApp) executionsOf(run runLink) []executionRow {
	a.t.Helper()
	return executionRowsIn(a.t, a.page(run.Path))
}

func (a *runningApp) onlyRun() runLink {
	a.t.Helper()
	links := a.runLinks()
	if len(links) != 1 {
		a.t.Fatalf("runs = %v, want exactly one", links)
	}
	return links[0]
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(body)
}

func parsePage(t *testing.T, page string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}
	return doc
}

func listItemsIn(t *testing.T, page, label string) []*html.Node {
	t.Helper()
	var items []*html.Node
	for _, list := range findAll(parsePage(t, page), func(n *html.Node) bool {
		return n.Data == "ul" && attr(n, "aria-label") == label
	}) {
		items = append(items, findAll(list, func(n *html.Node) bool { return n.Data == "li" })...)
	}
	return items
}

func caseTitlesIn(t *testing.T, page string) []string {
	t.Helper()
	titles := []string{}
	for _, item := range listItemsIn(t, page, "Cases") {
		titles = append(titles, strings.TrimSpace(textOf(item)))
	}
	return titles
}

func runLinksIn(t *testing.T, page string) []runLink {
	t.Helper()
	links := []runLink{}
	for _, item := range listItemsIn(t, page, "Runs") {
		anchors := findAll(item, func(n *html.Node) bool { return n.Data == "a" })
		if len(anchors) != 1 {
			t.Fatalf("run item has %d links, want one", len(anchors))
		}
		links = append(links, runLink{Name: strings.TrimSpace(textOf(anchors[0])), Path: attr(anchors[0], "href")})
	}
	return links
}

func executionRowsIn(t *testing.T, page string) []executionRow {
	t.Helper()
	rows := []executionRow{}
	for _, item := range listItemsIn(t, page, "Executions") {
		cells := findAll(item, func(n *html.Node) bool { return n.Data == "span" })
		if len(cells) != 2 {
			t.Fatalf("execution item has %d spans, want title and label", len(cells))
		}
		rows = append(rows, executionRow{Title: strings.TrimSpace(textOf(cells[0])), Label: strings.TrimSpace(textOf(cells[1]))})
	}
	return rows
}

func alertTextIn(t *testing.T, page string) string {
	t.Helper()
	alerts := findAll(parsePage(t, page), func(n *html.Node) bool { return attr(n, "role") == "alert" })
	if len(alerts) == 0 {
		return ""
	}
	return strings.TrimSpace(textOf(alerts[0]))
}

func findAll(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && match(n) {
			found = append(found, n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return found
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textOf(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textOf(c))
	}
	return sb.String()
}

func newDBPath(t *testing.T) string {
	return filepath.Join(t.TempDir(), "qa-hub.db")
}

func TestCreatedCaseAppearsInList(t *testing.T) {
	app := startApp(t, newDBPath(t))

	status, _ := app.createCase("Login works")

	if status != http.StatusOK {
		t.Fatalf("create case: status %d, want %d", status, http.StatusOK)
	}
	if got, want := app.caseTitles(), []string{"Login works"}; !slices.Equal(got, want) {
		t.Fatalf("cases = %q, want %q", got, want)
	}
}

func TestTwoCasesWithTheSameTitleBothExist(t *testing.T) {
	app := startApp(t, newDBPath(t))

	app.createCase("Checkout")
	app.createCase("Checkout")

	if got, want := app.caseTitles(), []string{"Checkout", "Checkout"}; !slices.Equal(got, want) {
		t.Fatalf("cases = %q, want %q", got, want)
	}
}

func TestCasesSurviveRestartOnTheSameDatabaseFile(t *testing.T) {
	dbPath := newDBPath(t)
	first := startApp(t, dbPath)
	first.createCase("Login works")
	first.createCase("Logout works")
	before := first.caseTitles()
	first.stop()

	second := startApp(t, dbPath)

	after := second.caseTitles()
	if !slices.Equal(after, before) || len(after) != 2 {
		t.Fatalf("cases after restart = %q, want the %q from before", after, before)
	}
}

func TestBlankTitleIsRejectedWithErrorMessageAndNothingIsStored(t *testing.T) {
	blankTitles := map[string]string{
		"empty":           "",
		"spaces":          "   ",
		"tab and newline": "\t \n",
	}
	for name, blank := range blankTitles {
		t.Run(name, func(t *testing.T) {
			app := startApp(t, newDBPath(t))
			app.createCase("Existing case")

			status, body := app.createCase(blank)

			if status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want %d", status, http.StatusUnprocessableEntity)
			}
			if got := alertTextIn(t, body); !strings.Contains(got, "Title is required") {
				t.Fatalf("error message = %q, want it to contain %q", got, "Title is required")
			}
			if got, want := app.caseTitles(), []string{"Existing case"}; !slices.Equal(got, want) {
				t.Fatalf("cases = %q, want %q", got, want)
			}
		})
	}
}
