package web

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

var formErrors = map[error]string{
	qahub.ErrTitleRequired: "Title is required.",
	qahub.ErrNameRequired:  "Name is required.",
	qahub.ErrNoCases:       "Create a case before starting a run.",
}

func formError(err error) (string, bool) {
	for known, message := range formErrors {
		if errors.Is(err, known) {
			return message, true
		}
	}
	return "", false
}

type server struct {
	store *qahub.Store
}

func New(store *qahub.Store) http.Handler {
	s := &server{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/cases", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /cases", s.listCases)
	mux.HandleFunc("POST /cases", s.createCase)
	mux.HandleFunc("GET /runs", s.listRuns)
	mux.HandleFunc("POST /runs", s.startRun)
	mux.HandleFunc("GET /runs/{id}", s.showRun)
	return mux
}

func (s *server) listCases(w http.ResponseWriter, r *http.Request) {
	s.renderCases(w, r, http.StatusOK, formState{})
}

func (s *server) createCase(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	_, err := s.store.CreateCase(r.Context(), title)
	if message, ok := formError(err); ok {
		s.renderCases(w, r, http.StatusUnprocessableEntity, formState{Value: title, Error: message})
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/cases", http.StatusSeeOther)
}

func (s *server) renderCases(w http.ResponseWriter, r *http.Request, status int, form formState) {
	cases, err := s.store.Cases(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, casesPage(cases, form))
}

func (s *server) listRuns(w http.ResponseWriter, r *http.Request) {
	s.renderRuns(w, r, http.StatusOK, formState{})
}

func (s *server) startRun(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	_, err := s.store.StartRun(r.Context(), name)
	if message, ok := formError(err); ok {
		s.renderRuns(w, r, http.StatusUnprocessableEntity, formState{Value: name, Error: message})
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/runs", http.StatusSeeOther)
}

func (s *server) renderRuns(w http.ResponseWriter, r *http.Request, status int, form formState) {
	runs, err := s.store.Runs(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, runsPage(runs, form))
}

func (s *server) showRun(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	run, executions, err := s.store.FindRun(r.Context(), id)
	if errors.Is(err, qahub.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, http.StatusOK, runPage(run, executions))
}

func render(w http.ResponseWriter, r *http.Request, status int, page templ.Component) {
	var body bytes.Buffer
	if err := page.Render(r.Context(), &body); err != nil {
		serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	body.WriteTo(w)
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
