// Heroku app that serves api.webx.io.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/kr/fernet"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	ProvisionMessage = `
Congrats on your new domain name.
See http://git.io/51G0dQ
`
	NoNameMessage = `
Missing flag --name.
See http://git.io/51G0dQ for help.
`
)

const username = "webx"

var (
	fernetKey *fernet.Key
	password string
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	password = mustGetenv("HEROKU_PASSWORD")

	var err error
	fernetKey, err = fernet.DecodeKey(mustGetenv("FERNET_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	r := NewRouter()
	http.ListenAndServe(":"+port, r)
}

func NewRouter() *mux.Router {
	webxdPath, err := exec.LookPath("webxd")
	if err != nil {
		log.Fatal(err)
	}
	routerPath, err := exec.LookPath("urouter")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/heroku/resources", Create).Methods("POST")
	r.HandleFunc("/heroku/resources/{id}", Put).Methods("PUT")
	r.HandleFunc("/heroku/resources/{id}", Delete).Methods("DELETE")
	r.HandleFunc("/", Home).Methods("GET", "HEAD")
	r.Handle("/dyno-profile.sh", fileHandler("webxd/dyno-profile.sh")).Methods("GET", "HEAD")
	r.Handle("/webxd", fileHandler(webxdPath)).Methods("GET", "HEAD")
	r.Handle("/urouter", fileHandler(routerPath)).Methods("GET", "HEAD")
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
		jsonError(w, NoNameMessage, 422)
		return
	}
	sig, err := fernet.EncryptAndSign([]byte(hreq.Options.Name), fernetKey)
	if err != nil {
		log.Println("error signing app name:", err)
		w.WriteHeader(500)
	}

	log.Println("provision", hreq.Options.Name)
	var out struct {
		ID      string                    `json:"id"`
		Config  struct{ WEBX_URL string } `json:"config"`
		Message string                    `json:"message"`
	}
	out.ID = rands(10)
	out.Config.WEBX_URL = "https://" + hreq.Options.Name + ":" + string(sig) + "@route.webx.io/"
	out.Message = hreq.Options.Name + ".webxapp.io\n" + ProvisionMessage
	w.WriteHeader(201)
	err = json.NewEncoder(w).Encode(out)
	if err != nil {
		log.Println("error sending response to heroku:", err)
		http.Error(w, "internal error", 500)
	}
}

func Put(w http.ResponseWriter, r *http.Request) {
	if !authenticate(r) {
		log.Println("auth failure")
		w.WriteHeader(401)
		return
	}

	id := mux.Vars(r)["id"]
	log.Println("update", id)
	w.WriteHeader(200)
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

type J map[string]interface{}

func jsonError(w http.ResponseWriter, message string, code int) {
	message = strings.TrimSpace(message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(J{"message": message})
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
	if len(s) < 1 {
		return false
	}
	if c := s[0]; !('a' <= c && c <= 'z' || '0' <= c && c <= '9') {
		return false
	}
	for _, c := range s[1:] {
		if !('a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-') {
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
