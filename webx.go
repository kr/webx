package webx

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	neturl "net/url"
	"os"
	"strings"

	"github.com/kr/rspdy"
)

func DialAndServeTLS(url string, tlsConfig *tls.Config, h http.Handler) error {
	u, err := neturl.Parse(os.Getenv("RUNX_URL"))
	if err != nil {
		return err
	}
	if u.User == nil {
		return errors.New("url has no userinfo")
	}
	if u.Scheme != "https" {
		return errors.New("scheme must be https")
	}

	name := u.User.Username()
	password, _ := u.User.Password()
	cmd, err := json.Marshal(Command{"add", name, password})
	if err != nil {
		return err
	}
	handshake := func(w http.ResponseWriter, r *http.Request) {
		w.Write(cmd)
		select {}
	}

	if h == nil {
		h = http.DefaultServeMux
	}
	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.HandleFunc("backend.webx.io/names", handshake)
	addr := u.Host
	if !strings.Contains(addr, ":") {
		addr += ":https"
	}
	return rspdy.DialAndServeTLS("tcp", addr, tlsConfig, mux)
}

type Command struct {
	Op       string // "add" or "remove"
	Name     string // e.g. "foo" for foo.webxapp.io
	Password string
}
