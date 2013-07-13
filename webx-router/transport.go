package main

import (
	"github.com/kr/spdy"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Transport struct {
	tab map[string]*Group
	mu  sync.RWMutex
}

func (t *Transport) Make(name string) *Group {
	if g := t.Lookup(name); g != nil {
		return g
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	g := t.tab[name]
	if g == nil {
		g = new(Group)
		t.tab[name] = g
	}
	return g
}

func (t *Transport) Lookup(name string) *Group {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tab[name]
}

func (t *Transport) Add(name string, c *spdy.Conn) {
	t.Make(name).Add(c)
}

func (t *Transport) Remove(name string, c *spdy.Conn) {
	if g := t.Lookup(name); g != nil {
		g.Remove(c)
	}
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	name := strings.TrimSuffix(r.Host, ".webxapp.io")
	g := t.Lookup(name)
	if g == nil {
		return &http.Response{StatusCode: 503, Body: empty}, nil
	}
	log.Println("roundtrip", name)
	return g.RoundTrip(r)
}
