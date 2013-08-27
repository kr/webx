package main

import (
	"net/http"
	"testing"
)

func TestIDHandler(t *testing.T) {
	var cases = []struct {
		id     string
		header http.Header
	}{
		{"", http.Header{}},
		{"foo", http.Header{"Id": {"foo"}}},
	}

	for _, test := range cases {
		var got string
		var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			got = r.Header.Get("ID")
		}
		h := idHandler(f)
		req := &http.Request{Header: make(http.Header)}
		for k, v := range test.header {
			req.Header[k] = v
		}
		h.ServeHTTP(nil, req)
		if got == "" {
			t.Errorf("%v ID empty", test.header)
		}
		if test.id != "" && got != test.id {
			t.Errorf("%v ID = %q want %q", test.header, got, test.id)
		}
	}
}
