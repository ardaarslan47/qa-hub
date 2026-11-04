package web

import (
	"net/http"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

const libraryPath = "/library"

func (h handler) showLibrary(w http.ResponseWriter, r *http.Request) {
	h.renderLibrary(w, r, http.StatusOK, "")
}

func (h handler) createCase(w http.ResponseWriter, r *http.Request) {
	title, err := qahub.ParseTitle(r.PostFormValue("title"))
	if err != nil {
		h.renderLibrary(w, r, http.StatusUnprocessableEntity, "Title is required.")
		return
	}
	if err := h.store.CreateCase(r.Context(), title); err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, libraryPath, http.StatusSeeOther)
}

func (h handler) renderLibrary(w http.ResponseWriter, r *http.Request, status int, titleError string) {
	cases, err := h.store.Cases(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, libraryPage(cases, titleError))
}
