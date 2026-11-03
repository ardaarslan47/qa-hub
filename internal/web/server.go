package web

import (
	"bytes"
	"errors"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

var formErrors = map[error]string{
	qahub.ErrTitleRequired: "Title is required.",
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
	return mux
}

func (s *server) listCases(w http.ResponseWriter, r *http.Request) {
	s.renderCases(w, r, http.StatusOK, caseForm{})
}

func (s *server) createCase(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	_, err := s.store.CreateCase(r.Context(), title)
	if message, ok := formError(err); ok {
		s.renderCases(w, r, http.StatusUnprocessableEntity, caseForm{Title: title, Error: message})
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/cases", http.StatusSeeOther)
}

func (s *server) renderCases(w http.ResponseWriter, r *http.Request, status int, form caseForm) {
	cases, err := s.store.Cases(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, casesPage(cases, form))
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
