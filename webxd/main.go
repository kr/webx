// Usage: webxd
// Environment:
//   PORT     - port to send requests to
//   WEBX_URL - location and credentials for RSPDY connection
//              e.g. https://foo@route.webx.io/
//
// Optional Environment:
//   WEBX_VERBOSE - log extra information
package main

import (
	"crypto/tls"
	"encoding/json"
	"github.com/kr/rspdy"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	redialPause = 2 * time.Second
)

var verbose bool

var tlsConfig = &tls.Config{
	InsecureSkipVerify: true,
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("webxd: ")
	innerAddr := ":" + os.Getenv("PORT")
	routerURL, err := url.Parse(os.Getenv("WEBX_URL"))
	if err != nil {
		log.Fatal("parse url:", err)
	}
	mustSanityCheckURL(routerURL)

	handshake := func(w http.ResponseWriter, r *http.Request) {
		webxName := routerURL.User.Username()
		password, _ := routerURL.User.Password()
		cmd := BackendCommand{"add", webxName, password}
		err = json.NewEncoder(w).Encode(cmd)
		if err != nil {
			log.Fatal("handshake:", err)
		}
		select {}
	}

	innerURL := &url.URL{Scheme: "http", Host: innerAddr}
	rp := httputil.NewSingleHostReverseProxy(innerURL)
	rp.Transport = new(WebsocketTransport)
	http.Handle("/", LogHandler{rp})
	http.HandleFunc("backend.webx.io/names", handshake)
	if os.Getenv("WEBX_VERBOSE") != "" {
		verbose = true
	}
	for {
		addr := routerURL.Host
		if p := strings.LastIndex(addr, ":"); p == -1 {
			addr += ":https"
		}
		err = rspdy.DialAndServeTLS("tcp", addr, tlsConfig, nil)
		if err != nil {
			log.Println("DialAndServe:", err)
			time.Sleep(redialPause)
		}
	}
}

type LogHandler struct {
	http.Handler
}

func (h LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w = &LogResponseWriter{ResponseWriter: w, r: r}
	h.Handler.ServeHTTP(w, r)
}

type LogResponseWriter struct {
	http.ResponseWriter
	r      *http.Request
	logged bool
}

func (w *LogResponseWriter) Write(p []byte) (int, error) {
	w.log(http.StatusOK)
	return w.ResponseWriter.Write(p)
}

func (w *LogResponseWriter) WriteHeader(code int) {
	w.log(code)
	w.ResponseWriter.WriteHeader(code)
}

func (w *LogResponseWriter) log(code int) {
	if !w.logged {
		w.logged = true
		Infoln(code, w.r.Host, w.r.URL.Path)
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
	Op       string // "add" or "remove"
	Name     string // e.g. "foo" for foo.webxapp.io
	Password string
}

func Infoln(v ...interface{}) {
	if verbose {
		log.Println(v...)
	}
}

type WebsocketTransport struct{}

func (w WebsocketTransport) Proxy(req *http.Request) (*http.Response, error) {
	conn, err := net.DialTimeout("tcp", req.URL.Host, 50*time.Millisecond)
	if err != nil {
		return nil, err
	}
	go io.Copy(conn, req.Body)
	resp := &http.Response{
		StatusCode: 200,
		Body:       conn,
	}
	return resp, nil
}

func (w WebsocketTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == "WEBSOCKET" {
		return w.Proxy(req)
	}
	return DefaultTransport.RoundTrip(req)
}

// Transport is like http.Transport but it sets Close to true on all requests.
type Transport struct{}

var DefaultTransport = new(Transport)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Close = true
	return http.DefaultTransport.RoundTrip(req)
}
