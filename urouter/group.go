package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"sync"
)

type Group struct {
	routable []*Backend
	backends []*Backend
	mu       sync.RWMutex
}

func (g *Group) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b := g.route(r); b != nil {
		b.ServeHTTP(w, r)
	} else {
		w.WriteHeader(503)
		io.WriteString(w, "no backends")
	}
}

// route chooses a single Backend in g for r.
// If there are no routable backends, route returns nil.
func (g *Group) route(r *http.Request) *Backend {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if len(g.routable) == 0 {
		return nil
	}
	// TODO(kr): do something smarter than rand
	return g.routable[rand.Intn(len(g.routable))]
}

func (g *Group) Add(b *Backend) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.backends = append(g.backends, b)
}

func (g *Group) AddRoute(b *Backend) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.routable = append(g.routable, b)
}

func (g *Group) Remove(b *Backend) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.backends = backendsRemove(g.backends, b)
	g.routable = backendsRemove(g.routable, b)
}

// backendsRemove destructively removes elements of a that
// equal b and returns the resulting slice.
func backendsRemove(a []*Backend, b *Backend) []*Backend {
	i := 0
	for _, t := range a {
		if t != b {
			a[i] = t
			i++
		}
	}
	return a[:i]
}

func (g *Group) Monitor(w http.ResponseWriter, r *http.Request) {
	// TODO(kr): use spdy.Pusher when available
	all := make([]bufResp, len(g.backends))
	var wg sync.WaitGroup
	g.mu.Lock()
	for i, b := range g.backends {
		wg.Add(1)
		wp := &all[i]
		go func(b *Backend) {
			b.ServeHTTP(wp, r)
			wg.Done()
		}(b)
	}
	g.mu.Unlock()
	wg.Wait()
	json.NewEncoder(w).Encode(all)
}

type bufResp struct {
	Code int
	Body []byte
	Head http.Header `json:"Header"`
}

func (c *bufResp) WriteHeader(code int) {
	c.Code = code
}

func (c *bufResp) Header() http.Header {
	if c.Head == nil {
		c.Head = make(http.Header)
	}
	return c.Head
}

func (c *bufResp) Write(p []byte) (int, error) {
	c.Body = append(c.Body, p...)
	return len(p), nil
}
