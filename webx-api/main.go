// Heroku app that serves api.webx.io.
package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/heroku/webx/keys"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const ProvisionMessage = `
Congrats on your new domain name.
See http://git.io/51G0dQ
`

const username = "webx"

var (
	password   string
	signingKey *ecdsa.PrivateKey
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	password = mustGetenv("HEROKU_PASSWORD")

	// load elliptic curve
	ecdsaBits, err := strconv.Atoi(mustGetenv("ECDSA_BITS"))
	if err != nil {
		log.Fatalf("invalid ECDSA_BITS")
	}
	curve, err := keys.ParseCurveBits(ecdsaBits)
	if err != nil {
		log.Fatalf("invalid ECDSA curve size: %s", err)
	}

	// load private signing key
	privKeyStr := mustGetenv("ECDSA_PRIVATE_KEY")
	signingKey, err = keys.DecodePrivateKey(privKeyStr, curve)
	if err != nil {
		log.Fatalf("error loading signing key: %s", err)
	}

	r := NewRouter()
	http.ListenAndServe(":"+port, r)
}

func NewRouter() *mux.Router {
	webxdPath, err := exec.LookPath("webxd")
	if err != nil {
		log.Fatal(err)
	}
	routerPath, err := exec.LookPath("webx-router")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/heroku/resources", Create).
		Methods("POST")
	r.HandleFunc("/heroku/resources/{id}", Delete).
		Methods("DELETE")
	r.HandleFunc("/", Home).Methods("GET", "HEAD")
	r.Handle("/dyno-profile.sh", fileHandler("webxd/dyno-profile.sh")).Methods("GET", "HEAD")
	r.Handle("/webxd", fileHandler(webxdPath)).Methods("GET", "HEAD")
	r.Handle("/webx-router", fileHandler(routerPath)).Methods("GET", "HEAD")
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

	auth, err := keys.SignUser(hreq.Options.Name, signingKey)
	if err != nil {
		log.Println("signing error:", err)
		w.WriteHeader(500)
		return
	}

	var out struct {
		ID     string `json:"id"`
		Config struct {
			WEBX_URL string
		}
		Message string
	}
	out.ID = rands(10)
	out.Config.WEBX_URL = "https://" + hreq.Options.Name + ":" + auth + "@route.webx.io/"
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

func Home(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "webx\n")
}

type fileHandler string

func (h fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, string(h))
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

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("must set env %s", key)
	}
	return val
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
