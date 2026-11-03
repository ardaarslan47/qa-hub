package web

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"

	"github.com/ardaarslan47/qa-hub/internal/domain"
)

func (s *server) listCases(w http.ResponseWriter, r *http.Request) {
	s.renderCases(w, r, http.StatusOK, caseForm{})
}

func (s *server) createCase(w http.ResponseWriter, r *http.Request) {
	rawTitle := r.PostFormValue("title")
	title, err := domain.ParseCaseTitle(rawTitle)
	if err == nil {
		err = s.store.CreateCase(r.Context(), title)
	}
	var userErr domain.UserError
	if errors.As(err, &userErr) {
		s.renderCases(w, r, http.StatusUnprocessableEntity, caseForm{Title: rawTitle, Error: userErr.Error()})
		return
	}
	if err != nil {
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
