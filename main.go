package main

import (
	"fmt"
	"net/http"
	//	"time"
	"log"
)

func main() {
	fmt.Println("Welcome to chirpy")
	const port = "8080"

	// use the http.NewServerMux() function to create an empty servemux
	mux := http.NewServeMux()

	s := &http.Server{
	Addr:           ":" + port,
	Handler:        mux,
	}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(s.ListenAndServe())	

}
