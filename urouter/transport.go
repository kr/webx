package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
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
		n := len(t.ErrBody)
		resp := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(t.ErrBody)),
			Status:     strconv.Itoa(t.ErrCode) + " " + http.StatusText(t.ErrCode),
			StatusCode: t.ErrCode,
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			ProtoMinor: 0,
			Header: http.Header{
				"Connection":     {"close"},
				"Content-Length": {strconv.Itoa(n)},
			},
			Close:         true,
			ContentLength: int64(n),
		}
		return resp, nil
	}
	return rt.RoundTrip(r)
}
