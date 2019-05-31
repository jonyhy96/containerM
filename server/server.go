package server

import (
	"net/http"
	"time"
)

// NewServer return new http server
func NewServer(mux *http.ServeMux, addr string) *http.Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return srv
}
