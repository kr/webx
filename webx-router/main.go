package main

import (
	"encoding/json"
	"github.com/kr/rspdy"
	"github.com/kr/spdy"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
)

const (
	defRequestAddr = ":8000" // REQADDR
	defBackendAddr = ":4444" // BKDADDR
	certFile       = "route.webx.io.crt.pem"
	keyFile        = "route.webx.io.key.pem"
)

var empty emptyReadCloser

type emptyReadCloser int

func (e emptyReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (e emptyReadCloser) Close() error {
	return nil
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	p := NewProxy()
	go listenBackends(p)
	go listenRequests(p)
	select {}
}

func listenBackends(p *Proxy) {
	addr := os.Getenv("BKDADDR")
	if addr == "" {
		addr = defBackendAddr
	}
	l, err := rspdy.ListenTLS(addr, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.AcceptSPDY()
		if err != nil {
			log.Fatal("error: accept", err)
		}
		go handshakeBackend(c, p)
	}
}

func handshakeBackend(c *spdy.Conn, p *Proxy) {
	client := &http.Client{Transport: c}
	resp, err := client.Get("https://backend.webx.io/names")
	if err != nil {
		log.Println("error: get backend names:", err)
		return
	}
	if resp.StatusCode != 200 {
		log.Println("error: get backend names http status", resp.Status)
		return
	}
	// TODO(kr): defer remove c from p
	d := json.NewDecoder(resp.Body)
	var cmd BackendCommand
	for {
		err := d.Decode(&cmd)
		if err != nil {
			if err != io.EOF {
				log.Println("error: decode backend command:", err)
			}
			return
		}
		// TODO(kr): authenticate command as discussed
		switch cmd.Op {
		case "add":
			log.Println("add", cmd.Name)
			p.tr.Lookup(cmd.Name, true).Add(c)
		case "remove":
			log.Println("remove", cmd.Name)
			p.tr.Lookup(cmd.Name, true).Remove(c)
		}
	}
}

type BackendCommand struct {
	Op   string // "add" or "remove"
	Name string // e.g. "foo" for foo.webxapp.io
}

func listenRequests(p *Proxy) {
	// TODO(kr): TLS
	addr := os.Getenv("REQADDR")
	if addr == "" {
		addr = defRequestAddr
	}
	err := http.ListenAndServe(addr, &p.rp)
	if err != nil {
		log.Fatal("error: frontend ListenAndServe:", err)
	}
}

type Proxy struct {
	rp httputil.ReverseProxy
	tr Transport
}

func NewProxy() (p *Proxy) {
	p = new(Proxy)
	p.tr.tab = make(map[string]*Group)
	p.rp.Director = func(*http.Request) {}
	p.rp.Transport = &p.tr
	return p
}

type Transport struct {
	tab map[string]*Group
	mu  sync.Mutex
}

func (t *Transport) Lookup(name string, create bool) *Group {
	log.Println("lookup", name)
	t.mu.Lock()
	defer t.mu.Unlock()
	g := t.tab[name]
	if g == nil && create {
		g = new(Group)
		t.tab[name] = g
	}
	return g
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	name := strings.TrimSuffix(r.Host, ".webxapp.io")
	g := t.Lookup(name, false)
	if g == nil {
		return &http.Response{StatusCode: 503, Body: empty}, nil
	}
	log.Println("roundtrip", name)
	return g.RoundTrip(r)
}

type Group struct {
	backends []*spdy.Conn
	mu       sync.Mutex
}

func (g *Group) RoundTrip(r *http.Request) (*http.Response, error) {
	g.mu.Lock()
	if len(g.backends) == 0 {
		g.mu.Unlock()
		return &http.Response{StatusCode: 503, Body: empty}, nil
	}
	// TODO(kr): do something smarter than rand
	c := g.backends[rand.Intn(len(g.backends))]
	g.mu.Unlock()
	log.Println("spdy roundtrip", r.Host)
	return c.RoundTrip(r)
}

func (g *Group) Add(c *spdy.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()
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
