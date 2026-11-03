package web

import (
	"net/http"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

func (h handler) listCases(w http.ResponseWriter, r *http.Request) {
	h.renderCases(w, r, http.StatusOK, "")
}

func (h handler) createCase(w http.ResponseWriter, r *http.Request) {
	title, err := qahub.ParseTitle(r.PostFormValue("title"))
	if err != nil {
		h.renderCases(w, r, http.StatusUnprocessableEntity, "Title is required.")
		return
	}
	if err := h.store.CreateCase(r.Context(), title); err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/cases", http.StatusSeeOther)
}

func (h handler) renderCases(w http.ResponseWriter, r *http.Request, status int, titleError string) {
	cases, err := h.store.Cases(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, casesPage(cases, titleError))
}
