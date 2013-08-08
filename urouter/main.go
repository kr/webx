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
	defRequestAddr = ":8000" // REQADDR
	defBackendAddr = ":4444" // BKDADDR
	certFile       = "route.webx.io.crt.pem"
	keyFile        = "route.webx.io.key.pem"
)

var (
	empty      emptyReadCloser
	fernetKeys []*fernet.Key
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	fernetKeys = fernet.MustDecodeKeys(mustGetenv("FERNET_KEY"))

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
	handler := &WebsocketProxy{proxy, t}
	log.Println("listen requests", addr)
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatal("error: frontend ListenAndServe:", err)
	}
	panic("unreached")
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
			log.Println("accept spdy", err)
			continue
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
	Op       string // "add" or "remove"
	Name     string // e.g. "foo" for foo.webxapp.io
	Password string
}

type emptyReadCloser int

func (e emptyReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (e emptyReadCloser) Close() error {
	return nil
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("must set env %s", key)
	}
	return val
}
