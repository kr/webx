package main

import (
	"encoding/json"
	"github.com/fernet/fernet-go"
	"github.com/kr/spdy"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type Backend struct {
	conn   *spdy.Conn
	client http.Client
	proxy  httputil.ReverseProxy
	ws     WebsocketProxy
	ipauth IPAuth // optional in chain of handlers
	http.Handler
}

func NewBackend(c *spdy.Conn) *Backend {
	b := new(Backend)
	b.conn = c
	b.client.Transport = c
	b.proxy.Transport = c
	b.proxy.Director = NopDirector
	b.ws.handler = &b.proxy
	b.ws.transport = c
	b.ipauth.handler = &b.ws
	b.Handler = &b.ws
	return b
}

func NopDirector(*http.Request) {}

func (b *Backend) Handshake(dir *Directory) {
	resp, err := b.client.Get("https://backend.webx.io/names")
	if err != nil {
		log.Println("error: get backend names:", err)
		return
	}
	if resp.StatusCode != 200 {
		log.Println("error: get backend names http status", resp.Status)
		return
	}
	var name string
	defer func() {
		log.Println("remove", name)
		if g := dir.Get(name); g != nil {
			g.Remove(b)
		}
	}()
	d := json.NewDecoder(resp.Body)
	var cmd struct {
		Op    string // e.g. "web" or "mon"
		Token string
		OkIPs []net.IP
	}
	for {
		err := d.Decode(&cmd)
		if err != nil {
			if err != io.EOF {
				log.Println("error: decode backend command:", err)
			}
			return
		}
		msg := fernet.VerifyAndDecrypt([]byte(cmd.Token), time.Hour*24*365, fernetKeys)
		if msg == nil {
			return // unauthorized
		}

		name = string(msg)
		switch cmd.Op {
		case "web":
			log.Println("web", name)
			if len(cmd.OkIPs) > 0 {
				b.ipauth.addrs = cmd.OkIPs
				b.Handler = &b.ipauth
			}
			g := dir.Make(name)
			g.Add(b)
			g.AddRoute(b)
		case "mon":
			log.Println("mon", name)
			dir.Make(name).Add(b)
		}
	}
}

// stringsRemove destructively removes elements of a that
// equal s and returns the resulting slice.
func stringsRemove(a []string, s string) []string {
	i := 0
	for _, t := range a {
		if t != s {
			a[i] = t
			i++
		}
	}
	return a[:i]
}
