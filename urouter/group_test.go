package main

import (
	"github.com/kr/spdy"
	"net/http"
	"testing"
)

func TestGroup(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	g := NewGroup()
	rt := g.Lookup(req)
	if rt != nil {
		t.Fatalf("rt = %v want nil", rt)
	}
	c := new(spdy.Conn)
	g.Add(c)
	rt = g.Lookup(req)
	if gc, ok := rt.(*spdy.Conn); !ok || gc != c {
		t.Errorf("ok = false want true")
		t.Fatalf("gc = %p want %p", gc, c)
	}
	g.Remove(c)
	rt = g.Lookup(req)
	if rt != nil {
		t.Fatalf("rt = %v want nil", rt)
	}
}

func TestEmptyGroup(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	g := NewGroup()
	resp, err := g.RoundTrip(req)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("StatusCode = %d want 503", resp.StatusCode)
	}
}
