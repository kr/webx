package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type WebsocketProxy struct {
	handler   http.Handler
	transport http.RoundTripper
}

func (p *WebsocketProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We don't do the websocket handshake, we just decide whether
	// to proxy it as a websocket. The backend does the handshake.
	if r.Header.Get("Upgrade") == "websocket" {
		p.Proxy(w, r)
		return
	}
	p.handler.ServeHTTP(w, r)
}

func (p *WebsocketProxy) Proxy(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Println("hijack assertion failed", r.Host, r.URL.Path)
		p.handler.ServeHTTP(w, r) // last-ditch effort as plain http
		return
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		log.Println("hijack failed", r.Host, r.URL.Path, err)
		p.handler.ServeHTTP(w, r) // last-ditch effort as plain http
		return
	}
	defer conn.Close()
	rw.Flush()

	wrapreq := new(http.Request)
	wrapreq.Proto = "HTTP/1.1"
	wrapreq.ProtoMajor, wrapreq.ProtoMinor = 1, 1
	wrapreq.Method = "WEBSOCKET"
	wrapreq.Host = r.Host
	const dummy = "/"
	wrapreq.URL = &url.URL{Path: dummy}
	var buf bytes.Buffer
	r.Write(&buf)
	wrapreq.Body = ioutil.NopCloser(io.MultiReader(&buf, conn))
	resp, err := p.transport.RoundTrip(wrapreq)
	if err != nil || resp.StatusCode != 200 {
		io.WriteString(conn, "HTTP/1.0 503 Gateway Failed\r\n")
		io.WriteString(conn, "Connection: close\r\n\r\n")
		return
	}
	defer resp.Body.Close()
	io.Copy(conn, resp.Body)
}
