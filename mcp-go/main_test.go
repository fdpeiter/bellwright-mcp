package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrigins(t *testing.T) {
	for _, s := range []string{"https://127.0.0.1", "http://example.com", "http://u:p@localhost", "http://localhost/secret", "http://localhost?x=1"} {
		if _, e := newBackend(s); e == nil {
			t.Fatalf("accepted %s", s)
		}
	}
}
func TestRead(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/world" {
			t.Error("unexpected request")
		}
		w.Write([]byte(`{"snapshot_id":7}`))
	}))
	defer s.Close()
	b, e := newBackend(s.URL)
	if e != nil {
		t.Fatal(e)
	}
	_, out, e := b.read(context.Background(), "/api/world", "save", "not live")
	if e != nil || out.Source != "save" || out.RetrievedAt == "" {
		t.Fatalf("%+v %v", out, e)
	}
}
func TestRejectRedirect(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "http://example.com", 302) }))
	defer s.Close()
	b, _ := newBackend(s.URL)
	if _, _, e := b.read(context.Background(), "/api/world", "save", ""); e == nil {
		t.Fatal("followed redirect")
	}
}
