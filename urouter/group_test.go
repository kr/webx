package main

import (
	"net/http"
	"testing"
)

var backend = NewBackend(nil)

func TestGroupPick(t *testing.T) {
	var cases = []struct {
		g *Group
		b *Backend
	}{
		{&Group{}, nil},
		{&Group{backends: []*Backend{backend}}, backend},
	}

	req := new(http.Request)
	for _, test := range cases {
		b := test.g.pick(req)
		if b != test.b {
			t.Errorf("b = %v want %v", b, test.b)
		}
	}
}

func TestGroupAdd(t *testing.T) {
	g := new(Group)
	b := NewBackend(nil)
	g.Add(b)
	if n := len(g.backends); n != 1 {
		t.Fatalf("len(g.backends) = %d want 1", n)
	}
	if gotb := g.backends[0]; gotb != b {
		t.Fatalf("gotb = %p want %p", gotb, b)
	}
}

func TestGroupRemove(t *testing.T) {
	b := NewBackend(nil)
	g := &Group{backends: []*Backend{b}}
	g.Remove(b)
	if n := len(g.backends); n != 0 {
		t.Fatalf("len(g.backends) = %d want 0", n)
	}
}

func TestGroupEmptyResp(t *testing.T) {
	req := new(http.Request)
	g := new(Group)
	w := new(resp)
	g.ServeHTTP(w, req)
	if w.code != 503 {
		t.Errorf("code = %d want 503", w.code)
	}
}
