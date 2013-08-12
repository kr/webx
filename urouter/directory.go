package main

import (
	"github.com/kr/spdy"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Directory struct {
	tab map[string]*Group
	mu  sync.RWMutex
}

func (d *Directory) Make(name string) *Group {
	if g := d.Get(name); g != nil {
		return g
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	g := d.tab[name]
	if g == nil {
		g = new(Group)
		d.tab[name] = g
	}
	return g
}

func (d *Directory) Get(name string) *Group {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.tab[name]
}

func (d *Directory) Add(name string, c *spdy.Conn) {
	d.Make(name).Add(c)
}

func (d *Directory) Remove(name string, c *spdy.Conn) {
	if g := d.Get(name); g != nil {
		g.Remove(c)
	}
}

func (d *Directory) Lookup(r *http.Request) http.RoundTripper {
	name := strings.TrimSuffix(basehost(r.Host), ".webxapp.io")
	return d.Get(name)
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
