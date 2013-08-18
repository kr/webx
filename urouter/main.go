package main

import (
	"encoding/json"
	"github.com/kr/fernet"
	"github.com/kr/rspdy"
	"github.com/kr/spdy"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const (
	defRequestAddr    = ":8000" // REQADDR
	defRequestTLSAddr = ":4443" // REQTLSADDR
	defBackendAddr    = ":1111" // BKDADDR

	innerCertFile = "inner.crt"
	innerKeyFile  = "inner.key"
	outerCertFile = "outer.crt"
	outerKeyFile  = "outer.key"
)

var (
	fernetKeys []*fernet.Key // FERNET_KEY
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var err error
	fernetKeys, err = fernet.DecodeKeys(mustGetenv("FERNET_KEY"))
	if err != nil {
		log.Fatal("FERNET_KEY contains invalid keys: ", err)
	}

	d := &Directory{tab: make(map[string]*Group)}
	go listenBackends(d)
	go listenHTTP(d)
	go listenHTTPS(d)
	select {}
}

func listenHTTP(handler http.Handler) {
	addr := os.Getenv("REQADDR")
	if addr == "" {
		addr = defRequestAddr
	}
	log.Println("listen requests", addr)
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatal("error: frontend ListenAndServe:", err)
	}
	panic("unreached")
}

func listenHTTPS(handler http.Handler) {
	addr := os.Getenv("REQTLSADDR")
	if addr == "" {
		addr = defRequestTLSAddr
	}
	log.Println("listen requests tls", addr)
	err := http.ListenAndServeTLS(addr, outerCertFile, outerKeyFile, handler)
	if err != nil {
		log.Fatal("error: frontend ListenAndServeTLS:", err)
	}
	panic("unreached")
}

func listenBackends(dir *Directory) {
	addr := os.Getenv("BKDADDR")
	if addr == "" {
		addr = defBackendAddr
	}
	log.Println("listen backends", addr)
	l, err := rspdy.ListenTLS(addr, innerCertFile, innerKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.AcceptSPDY()
		if err != nil {
			log.Println("accept spdy", err)
			continue
		}
		b := NewBackend(c)
		go b.Handshake(dir)
	}
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("must set env %s", key)
	}
	return val
}

type Backend struct {
	conn   *spdy.Conn
	client http.Client
	proxy  httputil.ReverseProxy
	WebsocketProxy
}

func NewBackend(c *spdy.Conn) *Backend {
	b := new(Backend)
	b.conn = c
	b.client.Transport = c
	b.proxy.Transport = c
	b.proxy.Director = NopDirector
	b.WebsocketProxy.handler = &b.proxy
	b.WebsocketProxy.transport = c
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
	var names []string
	defer func() {
		for _, s := range names {
			log.Println("remove", s)
			if g := dir.Get(s); g != nil {
				g.Remove(b)
			}
		}
	}()
	d := json.NewDecoder(resp.Body)
	var cmd struct {
		Op       string // "add" or "remove"
		Name     string // e.g. "foo" for foo.webxapp.io
		Password string
	}
	for {
		err := d.Decode(&cmd)
		if err != nil {
			if err != io.EOF {
				log.Println("error: decode backend command:", err)
			}
			return
		}
		msg := fernet.VerifyAndDecrypt([]byte(cmd.Password), time.Hour*24*365, fernetKeys)
		if msg == nil {
			// auth failed
			log.Println("error: auth verification failure, name:", cmd.Name)
			return
		} else if token := string(msg); token != cmd.Name {
			// Name does not match verified Fernet token
			log.Printf("error: auth verification mismatch: name=%q token=%q\n", cmd.Name, token)
			return
		}

		switch cmd.Op {
		case "add":
			log.Println("add", cmd.Name)
			dir.Make(cmd.Name).Add(b)
			found := false
			for _, s := range names {
				if s == cmd.Name {
					found = true
				}
			}
			if !found {
				names = append(names, cmd.Name)
			}
		case "remove":
			log.Println("remove", cmd.Name)
			if g := dir.Get(cmd.Name); g != nil {
				g.Remove(b)
			}
			var a []string
			for i := range names {
				if names[i] != cmd.Name {
					a = append(a, names[i])
				}
			}
			names = a
		}
	}
}
