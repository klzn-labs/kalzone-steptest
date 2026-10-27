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
	db, err := openDatabase(os.Getenv("PLATFORM_DB_DSN"))
	if err != nil {
		return err
	}
	b := burner{db: db, iterations: iterations()}
	f := flooder{publisher: queuePublisher(), key: os.Getenv("FLOOD_KEY")}
	return http.ListenAndServe(":8080", routes(db, b, f))
}

func openDatabase(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	return sql.Open("mysql", dsn)
}

func queuePublisher() publisher {
	client, err := queue.FromEnv()
	if err != nil {
		return nil
	}
	return client
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
		if db == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "database unreachable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
