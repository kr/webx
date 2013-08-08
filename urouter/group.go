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
	g.mu.RLock()
	if len(g.backends) == 0 {
		g.mu.RUnlock()
		return &http.Response{StatusCode: 503, Body: empty}, nil
	}
	// TODO(kr): do something smarter than rand
	c := g.backends[rand.Intn(len(g.backends))]
	g.mu.RUnlock()
	log.Println("spdy roundtrip", r.Host)
	return c.RoundTrip(r)
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
