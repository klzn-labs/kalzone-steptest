package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"

	"github.com/klzn-labs/kalzone-sdk-go/queue"
)

const defaultIterations = 200000

var errNoDatabase = errors.New("no database attached")

type burner struct {
	db         *sql.DB
	iterations int
}

func iterations() int {
	n, err := strconv.Atoi(os.Getenv("BURN_ITERATIONS"))
	if err != nil || n < 1 {
		return defaultIterations
	}
	return n
}

func (b burner) burn(ctx context.Context, m queue.Message) error {
	if b.db == nil {
		return queue.Permanent(errNoDatabase)
	}
	var result int
	return b.db.QueryRowContext(ctx, "SELECT BENCHMARK(?, SHA2(?, 256))", b.iterations, m.ID).Scan(&result)
}
