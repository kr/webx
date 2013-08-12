package main

import (
	"net/http"
	"testing"
)

func TestTransportErr(t *testing.T) {
	tr := &Transport{func(*http.Request) http.RoundTripper {
		return nil
	}}
	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if resp.StatusCode != 503 {
		t.Fatalf("StatusCode = %d want 503", resp.StatusCode)
	}
}
