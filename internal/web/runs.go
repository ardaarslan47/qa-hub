package web

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

type runHandler func(http.ResponseWriter, *http.Request, qahub.Project, qahub.Run)

func (h handler) inRun(next runHandler) projectHandler {
	return func(w http.ResponseWriter, r *http.Request, project qahub.Project) {
		id, ok := pathID(r, "runID")
		if !ok {
			http.NotFound(w, r)
			return
		}
		run, err := h.store.FindRun(r.Context(), project.ID, qahub.RunID(id))
		if errors.Is(err, qahub.ErrRunNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			serverError(w, r, err)
			return
		}
		next(w, r, project, run)
	}
}

func (h handler) listRuns(w http.ResponseWriter, r *http.Request, project qahub.Project) {
	h.renderRuns(w, r, http.StatusOK, project, "")
}

func (h handler) startRun(w http.ResponseWriter, r *http.Request, project qahub.Project) {
	name, err := qahub.ParseRunName(r.PostFormValue("name"))
	if err != nil {
		h.renderRuns(w, r, http.StatusUnprocessableEntity, project, "Name is required.")
		return
	}
	err = h.store.StartRun(r.Context(), project.ID, name)
	if errors.Is(err, qahub.ErrNoCases) {
		h.renderRuns(w, r, http.StatusUnprocessableEntity, project, "Create a case before starting a run.")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, runsPath(project.ID), http.StatusSeeOther)
}

func (h handler) showRun(w http.ResponseWriter, r *http.Request, project qahub.Project, run qahub.Run) {
	executions, err := h.store.Executions(r.Context(), project.ID, run.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, http.StatusOK, runPage(project, run, executions))
}

func (h handler) enterStatus(w http.ResponseWriter, r *http.Request, project qahub.Project, run qahub.Run) {
	id, ok := pathID(r, "executionID")
	if !ok {
		http.NotFound(w, r)
		return
	}
	executionID := qahub.ExecutionID(id)
	executions, err := h.store.Executions(r.Context(), project.ID, run.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	if !slices.ContainsFunc(executions, func(e qahub.Execution) bool { return e.ID == executionID }) {
		http.NotFound(w, r)
		return
	}
	status, err := qahub.ParseStatus(r.PostFormValue("status"))
	if err != nil {
		http.Error(w, "Status must be Passed or Failed.", http.StatusBadRequest)
		return
	}
	err = h.store.EnterStatus(r.Context(), project.ID, run.ID, executionID, status)
	if errors.Is(err, qahub.ErrExecutionNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, runPath(project.ID, run.ID), http.StatusSeeOther)
}

func pathID(r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	return id, err == nil
}

func (h handler) renderRuns(w http.ResponseWriter, r *http.Request, status int, project qahub.Project, formError string) {
	runs, err := h.store.Runs(r.Context(), project.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, runsPage(project, runs, formError))
}
