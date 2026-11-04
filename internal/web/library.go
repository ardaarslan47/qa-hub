package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

const libraryPath = "/library"

func casePath(id qahub.CaseID) string {
	return fmt.Sprintf("%s/%d", libraryPath, id)
}

type caseForm struct {
	Title         string
	Description   string
	Priority      qahub.Priority
	TitleError    string
	PriorityError string
}

type caseInput struct {
	title       qahub.Title
	description qahub.Description
	priority    qahub.Priority
}

type caseHandler func(http.ResponseWriter, *http.Request, qahub.Case)

func (h handler) inCase(next caseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(r, "id")
		if !ok {
			http.NotFound(w, r)
			return
		}
		c, err := h.store.FindCase(r.Context(), qahub.CaseID(id))
		if errors.Is(err, qahub.ErrCaseNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			serverError(w, r, err)
			return
		}
		next(w, r, c)
	}
}

func readCaseForm(r *http.Request, fallback qahub.Priority) (caseForm, caseInput, bool) {
	form := caseForm{Title: r.PostFormValue("title"), Description: r.PostFormValue("description"), Priority: fallback}
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
	input := caseInput{title: title, description: qahub.ParseDescription(form.Description), priority: priority}
	return form, input, titleErr == nil && priorityErr == nil
}

func (h handler) showLibrary(w http.ResponseWriter, r *http.Request) {
	h.renderLibrary(w, r, http.StatusOK, caseForm{Priority: qahub.DefaultPriority()})
}

func (h handler) createCase(w http.ResponseWriter, r *http.Request) {
	form, input, ok := readCaseForm(r, qahub.DefaultPriority())
	if !ok {
		h.renderLibrary(w, r, http.StatusUnprocessableEntity, form)
		return
	}
	if err := h.store.CreateCase(r.Context(), input.title, input.description, input.priority); err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, libraryPath, http.StatusSeeOther)
}

func (h handler) showCase(w http.ResponseWriter, r *http.Request, c qahub.Case) {
	form := caseForm{Title: c.Title.String(), Description: c.Description.String(), Priority: c.Priority}
	render(w, r, http.StatusOK, casePage(c, form))
}

func (h handler) editCase(w http.ResponseWriter, r *http.Request, c qahub.Case) {
	form, input, ok := readCaseForm(r, c.Priority)
	if !ok {
		render(w, r, http.StatusUnprocessableEntity, casePage(c, form))
		return
	}
	err := h.store.UpdateCase(r.Context(), c.ID, input.title, input.description, input.priority)
	if errors.Is(err, qahub.ErrCaseNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
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
