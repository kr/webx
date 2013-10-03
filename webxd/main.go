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
	"github.com/kr/webx"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"time"
)

const (
	redialPause = 2 * time.Second
)

var (
	verbose bool
	dyno    = os.Getenv("DYNO")
)

var tlsConfig = &tls.Config{
	InsecureSkipVerify: true,
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("webxd: ")
	mode := "web"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "web":
		innerURL := &url.URL{Scheme: "http", Host: ":" + os.Getenv("PORT")}
		rp := httputil.NewSingleHostReverseProxy(innerURL)
		rp.Transport = new(WebsocketTransport)
		http.Handle("/", LogHandler{rp})
		http.HandleFunc("backend.webx.io/mon/ps", ListProc)
	case "mon":
		http.HandleFunc("backend.webx.io/mon/ps", ListProc)
	}
	if os.Getenv("WEBX_VERBOSE") != "" {
		verbose = true
	}
	for {
		err := webx.DialAndServeTLS(os.Getenv("WEBX_URL"), tlsConfig, nil)
		if err != nil {
			log.Println("DialAndServe:", err)
			log.Println("DialAndServe:", os.Getenv("WEBX_URL"))
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

func ListProc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Dyno", dyno)
	out, err := exec.Command("ps", "-ef").Output()
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error()+"\n")
		return
	}
	w.WriteHeader(200)
	w.Write(out)
}
