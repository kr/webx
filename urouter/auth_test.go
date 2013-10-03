package main

import (
	"testing"
)

func TestGetBasicAuth(t *testing.T) {
	var cases = []struct{ h, u, p string }{
		{"", "", ""},
		{"YTpi", "", ""},
		{"Basic %", "", ""},      // invalid base64
		{"Basic ", "", ""},       // valid empty string
		{"Basic Og==", "", ""},   // ":"
		{"Basic YQ==", "a", ""},  // "a"
		{"Basic YTo=", "a", ""},  // "a:"
		{"Basic OmI=", "", "b"},  // ":b"
		{"Basic YTpi", "a", "b"}, // "a:b"
		{"Basic  YTpi", "", ""},  // two spaces
	}
	for _, tt := range cases {
		u, p := basicAuth(tt.h)
		if u != tt.u || p != tt.p {
			t.Errorf("basicAuth(%q) = %q, %q want %q, %q", tt.h, u, p, tt.u, tt.p)
		}
	}
}
