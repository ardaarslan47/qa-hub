package web

import (
	"net/http"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

const libraryPath = "/library"

type caseForm struct {
	Title         string
	Description   string
	Priority      qahub.Priority
	TitleError    string
	PriorityError string
}

func (h handler) showLibrary(w http.ResponseWriter, r *http.Request) {
	h.renderLibrary(w, r, http.StatusOK, caseForm{Priority: qahub.DefaultPriority()})
}

func (h handler) createCase(w http.ResponseWriter, r *http.Request) {
	form := caseForm{Title: r.PostFormValue("title"), Description: r.PostFormValue("description"), Priority: qahub.DefaultPriority()}
	title, titleErr := qahub.ParseTitle(form.Title)
	priority, priorityErr := qahub.ParsePriority(r.PostFormValue("priority"))
	if titleErr != nil {
		form.TitleError = "Title is required."
	}
	if priorityErr != nil {
		form.PriorityError = "Priority must be Low, Medium or High."
	} else {
		form.Priority = priority
	}
	if titleErr != nil || priorityErr != nil {
		h.renderLibrary(w, r, http.StatusUnprocessableEntity, form)
		return
	}
	if err := h.store.CreateCase(r.Context(), title, qahub.ParseDescription(form.Description), priority); err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, libraryPath, http.StatusSeeOther)
}

func (h handler) renderLibrary(w http.ResponseWriter, r *http.Request, status int, form caseForm) {
	cases, err := h.store.Cases(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, libraryPage(cases, form))
}
