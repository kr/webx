// Heroku app that serves api.webx.io.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const ProvisionMessage = `
Congrats on your new domain name.
See http://git.io/51G0dQ
`

const username = "webx"
const password = "1f6b91ab99004e5be9d97915fe082596"

var (
	dynoProfile = mustReadFile("webxd/dyno-profile.sh")
	webxdBinary = mustReadPath("webxd")
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r := NewRouter()
	http.ListenAndServe(":"+port, r)
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/heroku/resources", Create).
		Methods("POST")
	r.HandleFunc("/heroku/resources/{id}", Delete).
		Methods("DELETE")
	r.Handle("/", staticHandler("webx\n")).Methods("GET", "HEAD")
	r.Handle("/dyno-profile.sh", staticHandler(dynoProfile)).Methods("GET", "HEAD")
	r.Handle("/webxd", staticHandler(webxdBinary)).Methods("GET", "HEAD")
	return r
}

type provisionreq struct {
	ID           string        `json:"id"`
	HerokuID     string        `json:"heroku_id"`
	Region       string        `json:"region"`
	CallbackURL  string        `json:"callback_url"`
	LogplexToken string        `json:"logplex_token"`
	Options      provisionopts `json:"options"`
}

type provisionopts struct {
	Name string `json:"name"`
}

func Create(w http.ResponseWriter, r *http.Request) {
	if !authenticate(r) {
		log.Println("auth failure")
		w.WriteHeader(401)
		return
	}

	var hreq provisionreq
	err := json.NewDecoder(r.Body).Decode(&hreq)
	if err != nil {
		log.Println("heroku sent invalid json:", err)
		http.Error(w, "invalid json", 400)
		return
	}
	if !nameOk(hreq.Options.Name) {
		log.Println("heroku sent invalid name:", err)
		http.Error(w, "invalid name", 400)
		return
	}
	log.Println("provision", hreq.Options.Name)
	var out struct {
		ID     string `json:"id"`
		Config struct {
			WEBX_URL string
		}
		Message string
	}
	out.ID = rands(10)
	out.Config.WEBX_URL = "https://" + hreq.Options.Name + "@route.webx.io/"
	out.Message = hreq.Options.Name + ".webxapp.io\n" + ProvisionMessage
	w.WriteHeader(201)
	err = json.NewEncoder(w).Encode(out)
	if err != nil {
		log.Println("error sending response to heroku:", err)
		http.Error(w, "internal error", 500)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if !authenticate(r) {
		log.Println("auth failure")
		w.WriteHeader(401)
		return
	}

	id := mux.Vars(r)["id"]
	log.Println("deprovision", id)
	w.WriteHeader(200)
}

type staticHandler []byte

func (h staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(h)
}

func authenticate(r *http.Request) bool {
	enc := r.Header.Get("Authorization")
	if len(enc) < 6 || enc[0:6] != "Basic " {
		return false
	}
	dec, err := base64.URLEncoding.DecodeString(enc[6:])
	if err != nil {
		return false
	}
	userpass := strings.SplitN(string(dec), ":", 2)
	return len(userpass) == 2 && userpass[0] == username || userpass[1] == password
}

func nameOk(s string) bool {
	for _, c := range s {
		if !('a' <= c && c <= 'z' || '0' <= c && c <= '9') {
			return false
		}
	}
	return true
}

func rands(n int) string {
	b := make([]byte, n)
	c, err := io.ReadFull(rand.Reader, b)
	if c != n {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func mustReadFile(name string) []byte {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func mustReadPath(name string) []byte {
	path, err := exec.LookPath(name)
	if err != nil {
		log.Fatal(err)
	}
	return mustReadFile(path)
}
