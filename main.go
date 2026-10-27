package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/klzn-labs/kalzone-sdk-go/queue"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("steptest stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	db, err := sql.Open("mysql", os.Getenv("PLATFORM_DB_DSN"))
	if err != nil {
		return err
	}
	publisher, err := queue.FromEnv()
	if err != nil {
		return err
	}
	b := burner{db: db, iterations: iterations()}
	f := flooder{publisher: publisher, key: os.Getenv("FLOOD_KEY")}
	return http.ListenAndServe(":8080", routes(db, b, f))
}

func routes(db *sql.DB, b burner, f flooder) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health(db))
	mux.Handle("POST /hooks/burn", queue.Handler(os.Getenv("PLATFORM_QUEUE_SIGNING_KEY"), b.burn))
	mux.HandleFunc("POST /flood", f.flood)
	return mux
}

func health(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "database unreachable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
