package main

import (
	"fmt"
	"net/http"
)

func main() {
	// ServeMux is an HTTP request multiplexer.
	// It matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL.
	mux := http.NewServeMux()

	// A Server defines parameters for running an HTTP server. The zero value for Server is a valid configuration.
	// Use the mux we created as the handler for our server
	server := http.Server{
		Handler: mux,
		Addr:    "localhost:8080",
	}

	fmt.Printf("Serving at %s", server.Addr)
	server.ListenAndServe()
}
