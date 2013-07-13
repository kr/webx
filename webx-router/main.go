package main

import (
	"encoding/json"
	"github.com/kr/rspdy"
	"github.com/kr/spdy"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

const (
	defRequestAddr = ":8000" // REQADDR
	defBackendAddr = ":4444" // BKDADDR
	certFile       = "route.webx.io.crt.pem"
	keyFile        = "route.webx.io.key.pem"
)

var empty emptyReadCloser

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	t := &Transport{tab: make(map[string]*Group)}
	go listenBackends(t)
	go listenRequests(t)
	select {}
}

func listenRequests(t *Transport) {
	// TODO(kr): TLS
	addr := os.Getenv("REQADDR")
	if addr == "" {
		addr = defRequestAddr
	}
	proxy := &httputil.ReverseProxy{
		Transport: t,
		Director:  func(*http.Request) {},
	}
	log.Println("listen requests", addr)
	err := http.ListenAndServe(addr, proxy)
	if err != nil {
		log.Fatal("error: frontend ListenAndServe:", err)
	}
}

func listenBackends(t *Transport) {
	addr := os.Getenv("BKDADDR")
	if addr == "" {
		addr = defBackendAddr
	}
	log.Println("listen backends", addr)
	l, err := rspdy.ListenTLS(addr, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.AcceptSPDY()
		if err != nil {
			log.Fatal("error: accept", err)
		}
		go handshakeBackend(c, t)
	}
}

func handshakeBackend(c *spdy.Conn, t *Transport) {
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
	var names []string
	defer func() {
		for _, s := range names {
			log.Println("remove", s)
			t.Remove(s, c)
		}
	}()
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
			t.Make(cmd.Name).Add(c)
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
			t.Remove(cmd.Name, c)
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

type BackendCommand struct {
	Op   string // "add" or "remove"
	Name string // e.g. "foo" for foo.webxapp.io
}

type emptyReadCloser int

func (e emptyReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (e emptyReadCloser) Close() error {
	return nil
}
