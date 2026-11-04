package web

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

const (
	libraryPath              = "/library"
	preconditionUnknownError = "A Precondition must be a case in the Library."
	preconditionCycleError   = "A case cannot be its own Precondition, directly or through other cases."
)

func casePath(id qahub.CaseID) string {
	return fmt.Sprintf("%s/%d", libraryPath, id)
}

type caseForm struct {
	Title              string
	Description        string
	Priority           qahub.Priority
	Preconditions      []qahub.CaseID
	TitleError         string
	PriorityError      string
	PreconditionsError string
}

type caseInput struct {
	title         qahub.Title
	description   qahub.Description
	priority      qahub.Priority
	preconditions []qahub.CaseID
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
	preconditions, preconditionsOK := parsePreconditions(r.PostForm["preconditions"])
	form.Preconditions = preconditions
	if !preconditionsOK {
		form.PreconditionsError = preconditionUnknownError
	}
	input := caseInput{
		title:         title,
		description:   qahub.ParseDescription(form.Description),
		priority:      priority,
		preconditions: preconditions,
	}
	return form, input, titleErr == nil && priorityErr == nil && preconditionsOK
}

func parsePreconditions(values []string) ([]qahub.CaseID, bool) {
	ok := true
	var ids []qahub.CaseID
	for _, value := range values {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			ok = false
			continue
		}
		ids = append(ids, qahub.CaseID(id))
	}
	return ids, ok
}

func preconditionRejection(err error) string {
	switch {
	case errors.Is(err, qahub.ErrPreconditionNotFound):
		return preconditionUnknownError
	case errors.Is(err, qahub.ErrPreconditionCycle):
		return preconditionCycleError
	}
	return ""
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
	err := h.store.CreateCase(r.Context(), input.title, input.description, input.priority, input.preconditions)
	if form.PreconditionsError = preconditionRejection(err); form.PreconditionsError != "" {
		h.renderLibrary(w, r, http.StatusUnprocessableEntity, form)
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, libraryPath, http.StatusSeeOther)
}

func (h handler) showCase(w http.ResponseWriter, r *http.Request, c qahub.Case) {
	preconditions, err := h.store.Preconditions(r.Context(), c.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	form := caseForm{Title: c.Title.String(), Description: c.Description.String(), Priority: c.Priority}
	for _, p := range preconditions {
		form.Preconditions = append(form.Preconditions, p.ID)
	}
	h.renderCase(w, r, http.StatusOK, c, form)
}

func (h handler) editCase(w http.ResponseWriter, r *http.Request, c qahub.Case) {
	form, input, ok := readCaseForm(r, c.Priority)
	if !ok {
		h.renderCase(w, r, http.StatusUnprocessableEntity, c, form)
		return
	}
	err := h.store.UpdateCase(r.Context(), c.ID, input.title, input.description, input.priority, input.preconditions)
	if form.PreconditionsError = preconditionRejection(err); form.PreconditionsError != "" {
		h.renderCase(w, r, http.StatusUnprocessableEntity, c, form)
		return
	}
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

func (h handler) renderCase(w http.ResponseWriter, r *http.Request, status int, c qahub.Case, form caseForm) {
	preconditions, err := h.store.Preconditions(r.Context(), c.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	cases, err := h.store.Cases(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	choices := slices.DeleteFunc(cases, func(other qahub.Case) bool { return other.ID == c.ID })
	render(w, r, status, casePage(c, preconditions, choices, form))
}
