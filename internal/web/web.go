package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ardaarslan47/qa-hub/internal/store"
)

type server struct {
	store *store.Store
}

func New(st *store.Store) http.Handler {
	s := &server{store: st}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/cases", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /cases", s.listCases)
	mux.HandleFunc("POST /cases", s.createCase)
	mux.HandleFunc("GET /runs", s.listRuns)
	mux.HandleFunc("POST /runs", s.startRun)
	mux.HandleFunc("GET /runs/{id}", s.showRun)
	mux.HandleFunc("POST /executions/{id}/status", s.setExecutionStatus)
	return mux
}

func internalError(w http.ResponseWriter, action string, err error) {
	slog.Error(action, "err", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return 0, false
	}
	return id, true
}
