package main

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestTransportErr(t *testing.T) {
	tr := &Transport{
		Director: func(*http.Request) http.RoundTripper {
			return nil
		},
		ErrCode: 456,
		ErrBody: "foo",
	}
	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if resp.StatusCode != tr.ErrCode {
		t.Errorf("StatusCode = %d want %d", resp.StatusCode, tr.ErrCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if g := string(body); g != tr.ErrBody {
		t.Fatalf("body = %q want %q", g, tr.ErrBody)
	}
}
