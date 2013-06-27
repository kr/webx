// Heroku app that serves api.webx.io.
package main

import (
	"encoding/hex"
)

const ProvisionMessage = `
Congrats on your new domain name.
See http://git.io/51G0dQ
`

func main() {
	http.HandleFunc("/heroku/resources", Create)
	http.HandleFunc("/heroku/resources/", Delete)
}

func Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", 405)
		return
	}

	var hreq struct {
		ID      string
		Options struct {
			Name string
		}
	}
	json.NewDecoder(r.Body).Decode(&hreq)
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
	log.Println("provision", name)
	var out struct {
		ID     string `json:"id"`
		Config struct {
			WEBX_URL string
		}
		Message string
	}
	out.ID = rands(10)
	out.Config.WEBX_URL = "https://" + name + "@route.webx.io/"
	out.Message = name + ".webxapp.io\n" + ProvisionMessage
	w.Status(201, "created")
	err = json.NewEncoder(w).Encode(out)
	if err != nil {
		log.Println("error send response to heroku:", err)
		http.Error(w, "internal error", 500)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", 405)
		return
	}

	// TODO
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
