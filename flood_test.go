package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/klzn-labs/kalzone-sdk-go/queue"
)

type countingPublisher struct {
	mu    sync.Mutex
	count int
}

func (p *countingPublisher) Publish(context.Context, string, []byte, ...queue.Option) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.count++
	return "id", nil
}

func (p *countingPublisher) published() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.count
}

func floodRequest(key, n string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/flood?n="+n, nil)
	r.Header.Set("Flood-Key", key)
	return r
}

func Test_TheFloodNeedsTheKey(t *testing.T) {
	cases := map[string]flooder{"wrong key": {key: "right"}, "no key configured": {}}
	for name, f := range cases {
		w := httptest.NewRecorder()
		f.flood(w, floodRequest("wrong", "10"))
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s: got %d", name, w.Code)
		}
	}
}

func Test_TheFloodRefusesABadCount(t *testing.T) {
	for _, n := range []string{"0", "100001", "many"} {
		w := httptest.NewRecorder()
		flooder{publisher: &countingPublisher{}, key: "k"}.flood(w, floodRequest("k", n))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("n=%s: got %d", n, w.Code)
		}
	}
}

func Test_TheFloodNeedsAQueue(t *testing.T) {
	w := httptest.NewRecorder()

	flooder{key: "k"}.flood(w, floodRequest("k", "10"))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d", w.Code)
	}
}

func Test_TheFloodPublishesEveryMessage(t *testing.T) {
	p := &countingPublisher{}
	w := httptest.NewRecorder()

	flooder{publisher: p, key: "k"}.flood(w, floodRequest("k", "250"))

	if w.Code != http.StatusAccepted {
		t.Fatalf("got %d", w.Code)
	}
	deadline := time.Now().Add(2 * time.Second)
	for p.published() < 250 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := p.published(); got != 250 {
		t.Fatalf("published %d", got)
	}
}
