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
	g := new(Group)
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
