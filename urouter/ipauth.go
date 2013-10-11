package main

import (
	"net"
	"net/http"
)

type IPAuth struct {
	handler http.Handler
	addrs   []net.IP
}

func (h *IPAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(w, r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	h.handler.ServeHTTP(w, r)
}

// isAuthorized returns whether r is authorized by the policy
// defined in h.
func (h *IPAuth) isAuthorized(w http.ResponseWriter, r *http.Request) bool {
	ip := net.ParseIP(r.RemoteAddr)
	for _, a := range h.addrs {
		if a.Equal(ip) {
			return true
		}
	}
	return false
}
