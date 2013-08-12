package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
)

type Directory struct {
	tab map[string]*Group
	mu  sync.RWMutex
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

func (d *Directory) Lookup(r *http.Request) http.RoundTripper {
	name := strings.TrimSuffix(basehost(r.Host), ".webxapp.io")
	g := d.Get(name)
	if g != nil {
		return g
	}
	return nil // avoid typed nil in RoundTripper interface
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
