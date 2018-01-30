package liltunnel

import (
	"bytes"
	"net/http"
)

// _sigh_. Go? Why u no make status code visible?
type httpResponseWriter struct {
	orig http.ResponseWriter
	code int
	body bytes.Buffer
}

func (h *httpResponseWriter) Header() http.Header {
	return h.orig.Header()
}

func (h *httpResponseWriter) Write(b []byte) (int, error) {
	if h.code == 0 {
		h.code = http.StatusOK
	}

	wrote, err := h.body.Write(b)
	if err != nil {
		return wrote, err
	}

	return h.orig.Write(b)
}

func (h *httpResponseWriter) WriteHeader(c int) {
	h.code = c
	h.orig.WriteHeader(c)
}
