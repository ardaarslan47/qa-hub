package web

import (
	"errors"
	"net/http"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

func (h handler) listProjects(w http.ResponseWriter, r *http.Request) {
	h.renderProjects(w, r, http.StatusOK, "")
}

func (h handler) createProject(w http.ResponseWriter, r *http.Request) {
	name, err := qahub.ParseProjectName(r.PostFormValue("name"))
	if err != nil {
		h.renderProjects(w, r, http.StatusUnprocessableEntity, "Name is required.")
		return
	}
	err = h.store.CreateProject(r.Context(), name)
	if errors.Is(err, qahub.ErrProjectNameTaken) {
		h.renderProjects(w, r, http.StatusUnprocessableEntity, "A project with this name already exists.")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

func (h handler) showProject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		http.NotFound(w, r)
		return
	}
	project, err := h.store.FindProject(r.Context(), qahub.ProjectID(id))
	if errors.Is(err, qahub.ErrProjectNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, http.StatusOK, projectPage(project))
}

func (h handler) renderProjects(w http.ResponseWriter, r *http.Request, status int, formError string) {
	projects, err := h.store.Projects(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, projectsPage(projects, formError))
}
