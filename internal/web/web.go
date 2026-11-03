package web

import (
	"log/slog"
	"net/http"

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
	return mux
}

func internalError(w http.ResponseWriter, action string, err error) {
	slog.Error(action, "err", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
