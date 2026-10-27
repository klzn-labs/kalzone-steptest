package main

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/klzn-labs/kalzone-sdk-go/queue"
)

const (
	burnQueue  = "burn"
	maxFlood   = 100000
	publishers = 8
)

type publisher interface {
	Publish(ctx context.Context, queue string, body []byte, opts ...queue.Option) (string, error)
}

type flooder struct {
	publisher publisher
	key       string
}

func (f flooder) flood(w http.ResponseWriter, r *http.Request) {
	if !f.allowed(r.Header.Get("Flood-Key")) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if f.publisher == nil {
		http.Error(w, "no queue to publish to", http.StatusServiceUnavailable)
		return
	}
	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil || n < 1 || n > maxFlood {
		http.Error(w, "n must be 1 to 100000", http.StatusBadRequest)
		return
	}
	go f.publish(context.WithoutCancel(r.Context()), n)
	w.WriteHeader(http.StatusAccepted)
}

func (f flooder) allowed(given string) bool {
	return f.key != "" && subtle.ConstantTimeCompare([]byte(given), []byte(f.key)) == 1
}

func (f flooder) publish(ctx context.Context, n int) {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range publishers {
		wg.Go(func() { f.drain(ctx, jobs) })
	}
	for i := range n {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	slog.Info("flood published", "messages", n)
}

func (f flooder) drain(ctx context.Context, jobs <-chan int) {
	for i := range jobs {
		if _, err := f.publisher.Publish(ctx, burnQueue, []byte(strconv.Itoa(i))); err != nil {
			slog.Warn("publishing failed", "err", err)
		}
	}
}
