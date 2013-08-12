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
	rt := d.Lookup(req)
	if rt != nil {
		t.Fatalf("rt = %v want nil", rt)
	}

	w := d.Make(name)
	g = d.Get(name)
	if g != w {
		t.Fatalf("g = %v want %v", g, w)
	}
	rt = d.Lookup(req)
	if g, ok := rt.(*Group); !ok || g != w {
		t.Errorf("ok = %v want true", ok)
		t.Fatalf("g = %v want %v", g, w)
	}
}
