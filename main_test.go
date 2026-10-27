package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_AServiceWithoutADatabaseIsHealthy(t *testing.T) {
	db, err := openDatabase("")
	w := httptest.NewRecorder()

	health(db)(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if err != nil || w.Code != http.StatusOK {
		t.Fatalf("err %v, got %d", err, w.Code)
	}
}
