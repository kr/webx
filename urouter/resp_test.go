package main

import (
	"net/http"
)

type resp struct {
	code   int
	header http.Header
	body   []byte
}

func (c *resp) WriteHeader(code int) {
	c.code = code
}

func (c *resp) Header() http.Header {
	if c.header == nil {
		c.header = make(http.Header)
	}
	return c.header
}

func (c *resp) Write(p []byte) (int, error) {
	c.body = append(c.body, p...)
	return len(p), nil
}
