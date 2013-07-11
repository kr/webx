// Usage: webxd
// Environment:
//   PORT     - port to send requests to
//   WEBX_URL - location and credentials for RSPDY connection
//              e.g. https://foo@route.webx.io/
package main

import (
	"crypto/tls"
	"encoding/json"
	"github.com/kr/rspdy"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var tlsConfig = &tls.Config{
	InsecureSkipVerify: true,
}

func main() {
	log.SetFlags(0)
	log.SetFlags(log.Lshortfile|log.LstdFlags)
	log.SetPrefix("webxd: ")
	innerAddr := ":" + os.Getenv("PORT")
	routerURL, err := url.Parse(os.Getenv("WEBX_URL"))
	if err != nil {
		log.Fatal("parse url:", err)
	}
	mustSanityCheckURL(routerURL)

	handshake := func(w http.ResponseWriter, r *http.Request) {
		webxName := routerURL.User.Username()
		cmd := BackendCommand{"add", webxName}
		err = json.NewEncoder(w).Encode(cmd)
		if err != nil {
			log.Fatal("handshake:", err)
		}
		select {}
	}

	innerURL := &url.URL{Scheme: "http", Host: innerAddr}
	http.Handle("/", httputil.NewSingleHostReverseProxy(innerURL))
	http.HandleFunc("backend.webx.io/names", handshake)
	err = rspdy.DialAndServeTLS("tcp", routerURL.Host, tlsConfig, nil)
	if err != nil {
		log.Fatal("DialAndServe:", err)
	}
}

func mustSanityCheckURL(u *url.URL) {
	if u.User == nil {
		log.Fatal("url has no userinfo")
	}
	if u.Scheme != "https" {
		log.Fatal("scheme must be https")
	}
	if u.Path != "/" {
		log.Fatal("path must be /")
	}
	if u.RawQuery != "" {
		log.Fatal("query must be empty")
	}
	if u.Fragment != "" {
		log.Fatal("fragment must be empty")
	}
}

type BackendCommand struct {
	Op   string // "add" or "remove"
	Name string // e.g. "foo" for foo.webxapp.io
}
