package main

import (
	"net/http"
)

type Transport struct {
	Director func(*http.Request) http.RoundTripper
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	rt := t.Director(r)
	if rt == nil {
		return &http.Response{StatusCode: 503, Body: empty}, nil
	}
	return rt.RoundTrip(r)
}
