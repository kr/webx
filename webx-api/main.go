// Heroku app that serves api.webx.io.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
)

const ProvisionMessage = `
Congrats on your new domain name.
See http://git.io/51G0dQ
`

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
	return r
}

func Create(w http.ResponseWriter, r *http.Request) {
	var hreq struct {
		ID      string
		Options struct {
			Name string
		}
	}
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
		log.Println("error send response to heroku:", err)
		http.Error(w, "internal error", 500)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	log.Println("deprovision", id)
	w.WriteHeader(200)
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
