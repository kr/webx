package main

import (
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Directory struct {
	tab map[string]*Group
	mu  sync.RWMutex
}

func (d *Directory) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if g := d.pick(r); g != nil {
		g.ServeHTTP(w, r)
	} else {
		w.WriteHeader(404)
		io.WriteString(w, "no such app")
	}
}

func (d *Directory) pick(r *http.Request) *Group {
	name := strings.TrimSuffix(basehost(r.Host), ".webxapp.io")
	return d.Get(name)
}

func (d *Directory) Get(name string) *Group {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.tab[name]
}

func (d *Directory) Make(name string) *Group {
	if g := d.Get(name); g != nil {
		return g
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if g := d.tab[name]; g != nil {
		return g
	}
	g := new(Group)
	d.tab[name] = g
	return g
}

func basehost(hostport string) string {
	if !strings.Contains(hostport, ":") {
		return hostport
	}
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}
	return host
}
