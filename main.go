package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	// ServeMux is an HTTP request router
	mux := http.NewServeMux()
	fileserver := http.FileServer(http.Dir(filepathRoot))

	// HANDLERS:
	mux.Handle("/", fileserver)

	// A Server defines parameters for running an HTTP server
	// We use a pointer to specify the same server instance, instead of working with multiple copies (for each request)
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
