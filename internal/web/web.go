package web

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"

	"github.com/ardaarslan47/qa-hub/internal/domain"
	"github.com/ardaarslan47/qa-hub/internal/store"
)

type server struct {
	store *store.Store
}

func New(st *store.Store) http.Handler {
	s := &server{store: st}
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
	rawTitle := r.PostFormValue("title")
	title, err := domain.ParseCaseTitle(rawTitle)
	if err != nil {
		s.renderCases(w, r, http.StatusUnprocessableEntity, caseForm{Title: rawTitle, Error: err.Error()})
		return
	}
	if err := s.store.CreateCase(r.Context(), title); err != nil {
		internalError(w, "create case", err)
		return
	}
	http.Redirect(w, r, "/cases", http.StatusSeeOther)
}

func (s *server) renderCases(w http.ResponseWriter, r *http.Request, status int, form caseForm) {
	cases, err := s.store.ListCases(r.Context())
	if err != nil {
		internalError(w, "list cases", err)
		return
	}
	templ.Handler(casesPage(cases, form), templ.WithStatus(status)).ServeHTTP(w, r)
}

func internalError(w http.ResponseWriter, action string, err error) {
	slog.Error(action, "err", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
