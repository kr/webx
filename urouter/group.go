package main

import (
	"github.com/kr/spdy"
	"math/rand"
	"net/http"
	"sync"
)

type Group struct {
	backends []*spdy.Conn
	mu       sync.RWMutex

	Transport
}

func NewGroup() *Group {
	g := new(Group)
	g.Transport = Transport{g.Lookup, 503, "no backends"}
	return g
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
