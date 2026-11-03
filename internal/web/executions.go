package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"

	"github.com/ardaarslan47/qa-hub/internal/domain"
)

func executionAnchor(id int64) string {
	return fmt.Sprintf("execution-%d", id)
}

func executionStatusPath(id int64) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/executions/%d/status", id))
}

func statusLabel(e domain.Execution) string {
	if e.Status == nil {
		return "Untested"
	}
	return e.Status.String()
}

func (s *server) setExecutionStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var runID int64
	status, err := domain.ParseStatus(r.PostFormValue("status"))
	if err == nil {
		runID, err = s.store.SetExecutionStatus(r.Context(), id, status)
	}
	var userErr domain.UserError
	if errors.As(err, &userErr) {
		http.Error(w, userErr.Error(), http.StatusUnprocessableEntity)
		return
	}
	if errors.Is(err, domain.ErrExecutionNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalError(w, "set execution status", err)
		return
	}
	http.Redirect(w, r, string(runPath(runID))+"#"+executionAnchor(id), http.StatusSeeOther)
}
