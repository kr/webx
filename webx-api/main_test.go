package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fakehandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func BenchmarkDefaultServeMux(b *testing.B) {
	b.StopTimer()
	http.DefaultServeMux = http.NewServeMux()
	http.HandleFunc("/heroku/resources", fakehandler)
	http.HandleFunc("/heroku/delete", fakehandler)
	ts := httptest.NewServer(http.DefaultServeMux)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		res, err := http.Post(ts.URL+"/heroku/resources", "text", strings.NewReader(""))
		if err != nil {
			b.Fatal("Post:", err)
		}
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			b.Fatal("ReadAll:", err)
		}
	}

	b.StopTimer()
}
