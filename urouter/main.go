package main

import (
	"github.com/kr/fernet"
	"github.com/kr/rspdy"
	"log"
	"net/http"
	"os"
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
