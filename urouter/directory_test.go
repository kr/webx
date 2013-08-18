package main

import (
	"net/http"
	"testing"
)

func TestDirectory(t *testing.T) {
	const name = "foo"
	req := &http.Request{Host: name + ".webxapp.io"}
	d := &Directory{tab: make(map[string]*Group)}
	g := d.Get(name)
	if g != nil {
		t.Fatalf("g = %v want nil", g)
	}
	g = d.pick(req)
	if g != nil {
		t.Fatalf("g = %v want nil", g)
	}

	w := d.Make(name)
	g = d.Get(name)
	if g != w {
		t.Fatalf("g = %v want %v", g, w)
	}
	g = d.pick(req)
	if g != w {
		t.Fatalf("g = %v want %v", g, w)
	}
}

func TestDirectoryEmptyResp(t *testing.T) {
	d := &Directory{tab: make(map[string]*Group)}
	w := new(resp)
	d.ServeHTTP(w, &http.Request{Host: "foo.webxapp.io"})
	if w.code != 404 {
		t.Errorf("code = %d want 404", w.code)
	}
}
