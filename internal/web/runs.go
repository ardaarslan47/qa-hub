package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
)

func (h handler) listRuns(w http.ResponseWriter, r *http.Request) {
	h.renderRuns(w, r, http.StatusOK, "")
}

func (h handler) startRun(w http.ResponseWriter, r *http.Request) {
	name, err := qahub.ParseRunName(r.PostFormValue("name"))
	if err != nil {
		h.renderRuns(w, r, http.StatusUnprocessableEntity, "Name is required.")
		return
	}
	err = h.store.StartRun(r.Context(), name)
	if errors.Is(err, qahub.ErrNoCases) {
		h.renderRuns(w, r, http.StatusUnprocessableEntity, "Create a case before starting a run.")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/runs", http.StatusSeeOther)
}

func (h handler) showRun(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		http.NotFound(w, r)
		return
	}
	run, err := h.store.FindRun(r.Context(), qahub.RunID(id))
	if errors.Is(err, qahub.ErrRunNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	executions, err := h.store.Executions(r.Context(), run.ID)
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, http.StatusOK, runPage(run, executions))
}

func (h handler) enterStatus(w http.ResponseWriter, r *http.Request) {
	runID, runOK := pathID(r, "runID")
	executionID, executionOK := pathID(r, "id")
	if !runOK || !executionOK {
		http.NotFound(w, r)
		return
	}
	status, err := qahub.ParseStatus(r.PostFormValue("status"))
	if err != nil {
		http.Error(w, "Status must be Passed or Failed.", http.StatusBadRequest)
		return
	}
	err = h.store.EnterStatus(r.Context(), qahub.RunID(runID), qahub.ExecutionID(executionID), status)
	if errors.Is(err, qahub.ErrExecutionNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/runs/%d", runID), http.StatusSeeOther)
}

func pathID(r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	return id, err == nil
}

func (h handler) renderRuns(w http.ResponseWriter, r *http.Request, status int, formError string) {
	runs, err := h.store.Runs(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	render(w, r, status, runsPage(runs, formError))
}
