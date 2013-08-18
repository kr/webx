package main

import (
	"io"
	"math/rand"
	"net/http"
	"sync"
)

type Group struct {
	backends []*Backend
	mu       sync.RWMutex
}

func (g *Group) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b := g.pick(r); b != nil {
		b.ServeHTTP(w, r)
	} else {
		w.WriteHeader(503)
		io.WriteString(w, "no backends")
	}
}

func (g *Group) pick(r *http.Request) *Backend {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if len(g.backends) == 0 {
		return nil
	}
	// TODO(kr): do something smarter than rand
	return g.backends[rand.Intn(len(g.backends))]
}

func (g *Group) Add(b *Backend) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, b1 := range g.backends {
		if b1 == b {
			return
		}
	}
	g.backends = append(g.backends, b)
}

func (g *Group) Remove(b *Backend) {
	g.mu.Lock()
	defer g.mu.Unlock()
	var a []*Backend
	for _, b1 := range g.backends {
		if b1 != b {
			a = append(a, b1)
		}
	}
	g.backends = a
}
