package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base32"
	"github.com/fernet/fernet-go"
	"github.com/kr/spdy"
	"io"
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
	h := idHandler(d)
	go listenHTTP(h)
	go listenHTTPS(h)
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
	var srv spdy.Server
	srv.Addr = os.Getenv("BKDADDR")
	if srv.Addr == "" {
		srv.Addr = defBackendAddr
	}
	log.Println("listen backends", srv.Addr)
	srv.Handler = http.NewServeMux()
	srv.TLSConfig = &tls.Config{
		NextProtos: []string{"spdy/3", "rspdy/3", "http/1.1"},
	}
	srv.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){
		"rspdy/3": dir.ServeRSPDY,
	}
	err := srv.ListenAndServeTLS(innerCertFile, innerKeyFile)
	if err != nil {
		log.Fatal(err)
	}
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("must set env %s", key)
	}
	return val
}

// idHandler ensures each incoming request has a request ID
// in header field ID. If the field is already present, it is
// left alone; otherwise, idHandler generates a new random
// string.
func idHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Header["Id"]; !ok {
			r.Header.Set("Id", randID())
		}
		h.ServeHTTP(w, r)
	})
}

func randID() string {
	const n = 15
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		log.Fatal("randID", err)
	}
	return base32.StdEncoding.EncodeToString(b)
}
