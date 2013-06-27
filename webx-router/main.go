package main

import (
	"github.com/kr/rspdy"
	"log"
)

const (
	requestAddr = "1.2.3.4:80"
	backendAddr = "5.6.7.8:443"
)

func main() {
	p := NewProxy()
	go listenBackends(p)
	go listenRequests(p)
	select {}
}

func listenBackends(p *Proxy) {
	l, err := rspdy.ListenTLS(backendAddr, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal("error: accept", err)
		}
		go handshakeBackend(c, p)
	}
}

func handshakeBackend(c *spdy.Conn, p *Proxy) {
	resp, err := c.Get("https://backend.webx.io/names")
	if err != nil {
		log.Println("error: get backend names:", err)
		return
	}
	if resp.Status != 200 {
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
			p.tr.Add(cmd.Name, c)
		case "remove":
			p.tr.Remove(cmd.Name, c)
		}
	}
}

type BackendCommand struct {
	Op   string // "add" or "remove"
	Name string // e.g. "foo" for foo.webxapp.io
}

func listenRequests(p *Proxy) {
	// TODO(kr): TLS
	err := http.ListenAndServe(requestAddr, p.rp)
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
	p.tr.tab = make(map[string]*spdy.Conn)
	p.rp.Transport = &p.tr
	return p
}

type Transport struct {
	tab map[string][]*spdy.Conn
	mu  sync.Mutex
}

func (t *Transport) Add(name string, c *spdy.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tab[name] = append(t.tab[name], c)
}

func (t *Transport) Remove(name string, c *spdy.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	a := t.tab[name]
	var b []*spdy.Conn
	for i := range a {
		if a[i] != c {
			b = append(b, a[i])
		}
	}
	t.tab[name] = b
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	c, ok := t.Lookup(r.URL.Host)
}
