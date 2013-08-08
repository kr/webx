package main

import (
	"testing"
)

func TestBasehost(t *testing.T) {
	var cases = []struct {
		hostport string
		w        string
	}{
		{"example.com", "example.com"},
		{"example.com:80", "example.com"},
	}
	for _, test := range cases {
		g := basehost(test.hostport)
		if test.w != g {
			t.Errorf("hostport(%q) = %q want %q", test.hostport, g, test.w)
		}
	}
}
