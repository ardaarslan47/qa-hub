package web

import (
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
	mux.Handle("GET /{$}", http.RedirectHandler("/cases", http.StatusSeeOther))
	mux.HandleFunc("GET /cases", h.listCases)
	mux.HandleFunc("POST /cases", h.createCase)
	mux.HandleFunc("GET /runs", h.listRuns)
	mux.HandleFunc("POST /runs", h.startRun)
	mux.HandleFunc("GET /runs/{id}", h.showRun)
	mux.HandleFunc("POST /runs/{runID}/executions/{id}/status", h.enterStatus)
	return mux
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
