package main

import (
	"context"
	"database/sql"
	"os"
	"strconv"

	"github.com/klzn-labs/kalzone-sdk-go/queue"
)

const defaultIterations = 200000

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
	var result int
	return b.db.QueryRowContext(ctx, "SELECT BENCHMARK(?, SHA2(?, 256))", b.iterations, m.ID).Scan(&result)
}
