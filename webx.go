package webx

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/kr/rspdy"
)

type Server struct {
	Addr    string       // Address to dial
	Token   string       // Fernet token to provide in the handshake
	Op      string       // "web" or "mon"
	Handler http.Handler // Handler to invoke, http.DefaultServeMux if nil

	// TLSClientConfig specifies the TLS configuration to use with
	// tls.Client. If nil, the default configuration is used.
	TLSClientConfig *tls.Config
}

// ParseURL parses url and sets s.Addr, s.Token, and s.Op
// from the URL's password, host, and path fields. The URL
// must use scheme https.
func ParseURL(s *Server, url string) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	if u.User == nil {
		return errors.New("url has no userinfo")
	}
	if u.Scheme != "https" {
		return errors.New("scheme must be https")
	}
	s.Addr = u.Host
	if !strings.Contains(s.Addr, ":") {
		s.Addr += ":https"
	}
	s.Token, _ = u.User.Password()
	s.Op = strings.Trim(u.Path, "/")
	return nil
}

func (s *Server) DialAndServeTLS() error {
	cmd, err := json.Marshal(command{s.Op, s.Token})
	if err != nil {
		return err
	}
	handshake := func(w http.ResponseWriter, r *http.Request) {
		w.Write(cmd)
		select {}
	}
	h := s.Handler
	if h == nil {
		h = http.DefaultServeMux
	}
	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.HandleFunc("backend.webx.io/names", handshake)
	return rspdy.DialAndServeTLS("tcp", s.Addr, s.TLSClientConfig, mux)
}

type command struct {
	Op    string // "web" or "mon"
	Token string
}
