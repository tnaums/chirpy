package main

import (
	"net/http"
	"log"
)

func main() {
	const port = "8080"

	// use the http.NewServerMux() function to create an empty servemux
	mux := http.NewServeMux()

	// Use the http.FileServer() function to create a handler
	fh := http.FileServer(http.Dir("."))
	rh := http.RedirectHandler("http://example.org", 307)	
	mux.Handle("/", fh)
	mux.Handle("/foo", rh)
	
	s := &http.Server{
	Addr:           ":" + port,
	Handler:        mux,
	}
	
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(s.ListenAndServe())	

}
