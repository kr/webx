package main

import (
	"github.com/kr/spdy"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type Group struct {
	backends []*spdy.Conn
	mu       sync.RWMutex
}

func (g *Group) RoundTrip(r *http.Request) (*http.Response, error) {
	rt := g.Lookup(r)
	if rt == nil {
		return &http.Response{StatusCode: 503, Body: empty}, nil
	}
	log.Println("spdy roundtrip", r.Host)
	return rt.RoundTrip(r)
}

func (g *Group) Lookup(r *http.Request) http.RoundTripper {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if len(g.backends) == 0 {
		return nil
	}
	// TODO(kr): do something smarter than rand
	return g.backends[rand.Intn(len(g.backends))]
}

func (g *Group) Add(c *spdy.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, b := range g.backends {
		if b == c {
			return
		}
	}
	g.backends = append(g.backends, c)
}

func (g *Group) Remove(c *spdy.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()
	a := g.backends
	var b []*spdy.Conn
	for i := range a {
		if a[i] != c {
			b = append(b, a[i])
		}
	}
	g.backends = b
}

type Backend struct {
	c *spdy.Conn
}
