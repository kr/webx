package main

import (
	"encoding/json"
	"github.com/fernet/fernet-go"
	"github.com/kr/spdy"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

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
			return // unauthorized
		}

		name := string(msg)
		switch cmd.Op {
		case "add":
			log.Println("add", name)
			dir.Make(name).Add(b)
			found := false
			for _, s := range names {
				if s == name {
					found = true
				}
			}
			if !found {
				names = append(names, name)
			}
		case "remove":
			log.Println("remove", name)
			if g := dir.Get(name); g != nil {
				g.Remove(b)
			}
			var a []string
			for i := range names {
				if names[i] != name {
					a = append(a, names[i])
				}
			}
			names = a
		}
	}
}
