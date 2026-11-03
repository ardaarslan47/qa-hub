package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardaarslan47/qa-hub/internal/qahub"
	"github.com/ardaarslan47/qa-hub/internal/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "address to listen on; port 0 picks a free port")
	dbPath := flag.String("db", "qa-hub.db", "SQLite database file, created if missing")
	flag.Parse()
	if err := serve(*addr, *dbPath); err != nil {
		log.Fatal(err)
	}
}

func serve(addr, dbPath string) error {
	store, err := qahub.Open(dbPath)
	if err != nil {
		return err
	}
	defer store.Close()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("listening on http://%s", ln.Addr())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Handler: web.New(store)}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()

	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
