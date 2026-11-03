package web_test

import (
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
	"github.com/ardaarslan47/qa-hub/internal/web"
)

type app struct {
	t      *testing.T
	server *httptest.Server
	store  *qahub.Store
}

type response struct {
	Status int
	Body   string
}

func tempDB(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "qahub.db")
}

func startApp(t *testing.T, dbPath string) *app {
	t.Helper()
	store, err := qahub.Open(t.Context(), dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	a := &app{t: t, server: httptest.NewServer(web.New(store)), store: store}
	t.Cleanup(a.stop)
	return a
}

func (a *app) stop() {
	a.server.Close()
	a.store.Close()
}

func (a *app) get(path string) response {
	a.t.Helper()
	resp, err := a.server.Client().Get(a.server.URL + path)
	if err != nil {
		a.t.Fatalf("GET %s: %v", path, err)
	}
	return readResponse(a.t, resp)
}

func (a *app) post(path string, form url.Values) response {
	a.t.Helper()
	resp, err := a.server.Client().PostForm(a.server.URL+path, form)
	if err != nil {
		a.t.Fatalf("POST %s: %v", path, err)
	}
	return readResponse(a.t, resp)
}

func readResponse(t *testing.T, resp *http.Response) response {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return response{Status: resp.StatusCode, Body: string(body)}
}

func (a *app) createCase(title string) response {
	a.t.Helper()
	return a.post("/cases", url.Values{"title": {title}})
}

func (a *app) caseTitles() []string {
	a.t.Helper()
	return elementTexts(a.get("/cases").Body, "li")
}

var tags = regexp.MustCompile(`<[^>]*>`)

func elementTexts(body, element string) []string {
	pattern := regexp.MustCompile(`(?s)<` + element + `[ >].*?</` + element + `>`)
	var texts []string
	for _, match := range pattern.FindAllString(body, -1) {
		texts = append(texts, html.UnescapeString(tags.ReplaceAllString(match, "")))
	}
	return texts
}
