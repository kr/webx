package main

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type Transport struct {
	Director func(*http.Request) http.RoundTripper

	ErrCode int
	ErrBody string
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	rt := t.Director(r)
	if rt == nil {
		body := ioutil.NopCloser(strings.NewReader(t.ErrBody))
		return &http.Response{StatusCode: t.ErrCode, Body: body}, nil
	}
	return rt.RoundTrip(r)
}
