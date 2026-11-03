package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"

	"github.com/ardaarslan47/qa-hub/internal/domain"
)

func runPath(id int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/runs/%d", id))
}

func (s *server) listRuns(w http.ResponseWriter, r *http.Request) {
	s.renderRuns(w, r, http.StatusOK, runForm{})
}

func (s *server) startRun(w http.ResponseWriter, r *http.Request) {
	rawName := r.PostFormValue("name")
	name, err := domain.ParseRunName(rawName)
	if err == nil {
		err = s.store.StartRun(r.Context(), name)
	}
	var userErr domain.UserError
	if errors.As(err, &userErr) {
		s.renderRuns(w, r, http.StatusUnprocessableEntity, runForm{Name: rawName, Error: userErr.Error()})
		return
	}
	if err != nil {
		internalError(w, "start run", err)
		return
	}
	http.Redirect(w, r, "/runs", http.StatusSeeOther)
}

func (s *server) showRun(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	run, err := s.store.GetRun(r.Context(), id)
	if errors.Is(err, domain.ErrRunNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalError(w, "get run", err)
		return
	}
	executions, err := s.store.ListExecutions(r.Context(), id)
	if err != nil {
		internalError(w, "list executions", err)
		return
	}
	templ.Handler(runPage(run, executions)).ServeHTTP(w, r)
}

func (s *server) renderRuns(w http.ResponseWriter, r *http.Request, status int, form runForm) {
	runs, err := s.store.ListRuns(r.Context())
	if err != nil {
		internalError(w, "list runs", err)
		return
	}
	templ.Handler(runsPage(runs, form), templ.WithStatus(status)).ServeHTTP(w, r)
}
