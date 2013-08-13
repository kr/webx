package main

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestTransportErr(t *testing.T) {
	tr := &Transport{
		Director: func(*http.Request) http.RoundTripper { return nil },
		ErrCode:  500,
		ErrBody:  "foo",
	}
	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	g, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	const wbody = "foo"
	gbody, err := ioutil.ReadAll(g.Body)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if string(gbody) != wbody {
		t.Fatalf("body = %q want %q", string(gbody), wbody)
	}

	w := &http.Response{
		Status:     "500 Internal Server Error",
		StatusCode: 500,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Header: http.Header{
			"Connection":     {"close"},
			"Content-Length": {"3"},
		},
		Close:         true,
		ContentLength: 3,
	}
	g.Body = nil
	diff(t, "Response", g, w)
}

func diff(t *testing.T, prefix string, have, want interface{}) {
	hv := reflect.ValueOf(have).Elem()
	wv := reflect.ValueOf(want).Elem()
	if hv.Type() != wv.Type() {
		t.Errorf("%s: type mismatch %v want %v", prefix, hv.Type(), wv.Type())
	}
	for i := 0; i < hv.NumField(); i++ {
		hf := hv.Field(i).Interface()
		wf := wv.Field(i).Interface()
		if !reflect.DeepEqual(hf, wf) {
			t.Errorf("%s: %s = %v want %v", prefix, hv.Type().Field(i).Name, hf, wf)
		}
	}
}
