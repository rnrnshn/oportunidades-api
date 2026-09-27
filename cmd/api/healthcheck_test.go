package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunHealthcheck(t *testing.T) {
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(status)
	}))
	defer server.Close()
	_, port, _ := net.SplitHostPort(server.Listener.Addr().String())

	if err := runHealthcheck(port); err != nil {
		t.Fatalf("expected healthy, got %v", err)
	}

	status = http.StatusInternalServerError
	if err := runHealthcheck(port); err == nil {
		t.Fatal("expected an error for a failing health endpoint")
	}
}
