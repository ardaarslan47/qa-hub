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

func (a *app) createCase(title string) (int, string) {
	a.t.Helper()
	return a.do(http.MethodPost, "/cases", url.Values{"title": {title}})
}

func (a *app) mustCreateCase(title string) {
	a.t.Helper()
	if status, body := a.createCase(title); status != http.StatusOK {
		a.t.Fatalf("create case %q: status %d\n%s", title, status, body)
	}
}

func (a *app) caseTitles() []string {
	a.t.Helper()
	status, body := a.do(http.MethodGet, "/cases", nil)
	if status != http.StatusOK {
		a.t.Fatalf("GET /cases: status %d\n%s", status, body)
	}
	return listItems(a.t, body, "Cases")
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
