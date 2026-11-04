package web

import (
	"net/http"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

func (h handler) listCases(w http.ResponseWriter, r *http.Request, project qahub.Project) {
	h.renderCases(w, r, http.StatusOK, project, "")
}

func (h handler) createCase(w http.ResponseWriter, r *http.Request, project qahub.Project) {
	title, err := qahub.ParseTitle(r.PostFormValue("title"))
	if err != nil {
		h.renderCases(w, r, http.StatusUnprocessableEntity, project, "Title is required.")
		return
	}
	if err := h.store.CreateCase(r.Context(), project.ID, title); err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, casesPath(project.ID), http.StatusSeeOther)
}

func (h handler) renderCases(w http.ResponseWriter, r *http.Request, status int, project qahub.Project, titleError string) {
	cases, err := h.store.Cases(r.Context(), project.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, casesPage(project, cases, titleError))
}
