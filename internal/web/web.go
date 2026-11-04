package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

type handler struct {
	store *qahub.Store
}

func New(store *qahub.Store) http.Handler {
	h := handler{store: store}
	mux := http.NewServeMux()
	mux.Handle("GET /{$}", http.RedirectHandler("/projects", http.StatusSeeOther))
	mux.HandleFunc("GET "+libraryPath, h.showLibrary)
	mux.HandleFunc("POST "+libraryPath, h.createCase)
	mux.HandleFunc("GET /projects", h.listProjects)
	mux.HandleFunc("POST /projects", h.createProject)
	mux.HandleFunc("GET /projects/{id}", h.inProject(h.showProject))
	mux.HandleFunc("GET /projects/{id}/runs", h.inProject(h.listRuns))
	mux.HandleFunc("POST /projects/{id}/runs", h.inProject(h.startRun))
	mux.HandleFunc("GET /projects/{id}/runs/{runID}", h.inProject(h.inRun(h.showRun)))
	mux.HandleFunc("POST /projects/{id}/runs/{runID}/executions/{executionID}/status", h.inProject(h.inRun(h.enterStatus)))
	return mux
}

func projectPath(id qahub.ProjectID) string {
	return fmt.Sprintf("/projects/%d", id)
}

func runsPath(id qahub.ProjectID) string {
	return projectPath(id) + "/runs"
}

func runPath(project qahub.ProjectID, run qahub.RunID) string {
	return fmt.Sprintf("%s/%d", runsPath(project), run)
}

func statusPath(project qahub.ProjectID, run qahub.RunID, execution qahub.ExecutionID) string {
	return fmt.Sprintf("%s/executions/%d/status", runPath(project, run), execution)
}

func render(w http.ResponseWriter, r *http.Request, status int, page templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := page.Render(r.Context(), w); err != nil {
		log.Printf("render %s %s: %v", r.Method, r.URL.Path, err)
	}
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	http.Error(w, "Something went wrong.", http.StatusInternalServerError)
}
