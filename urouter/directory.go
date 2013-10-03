package main

import (
	"crypto/tls"
	"encoding/base64"
	"github.com/fernet/fernet-go"
	"github.com/kr/spdy"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
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

// ServeRSPDY serves an incoming RSPDY connection.
// It makes a new backend to represent the conn,
// then starts the handshake process.
func (d *Directory) ServeRSPDY(s *http.Server, c *tls.Conn, h http.Handler) {
	b := NewBackend(&spdy.Conn{Conn: c})
	b.Handshake(d)
}

// pick chooses the appropriate Group for r, based on the Host
// header field. If there is no such Group, pick returns nil.
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

func (d *Directory) Monitor(w http.ResponseWriter, r *http.Request) {
	_, tok := basicAuth(r.Header.Get("Authorization"))
	name := fernet.VerifyAndDecrypt([]byte(tok), time.Hour*24*365, fernetKeys)
	if name == nil {
		http.Error(w, "yo unauthorized", http.StatusUnauthorized)
		return
	}
	g := d.Get(string(name))
	if g == nil {
		http.NotFound(w, r)
		return
	}
	r.Host = "backend.webx.io"
	r.URL.Host = r.Host
	g.Monitor(w, r)
}

func basicAuth(h string) (username, password string) {
	if !strings.HasPrefix(h, "Basic ") {
		return
	}
	h = strings.TrimPrefix(h, "Basic ")
	b, err := base64.StdEncoding.DecodeString(h)
	if err != nil {
		return
	}
	s := string(b)
	if p := strings.Index(s, ":"); p > -1 {
		return s[:p], s[p+1:]
	}
	return s, ""
}
