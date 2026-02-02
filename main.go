package main

import (
	"net/http"
	"log"
	"time"
)


func helloHandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Hello, world!\n"))
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(time.RFC1123)
	w.Write([]byte("The time is: " + tm + "\n"))
}

func readinessEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func main() {
	const port = "8080"
	const filepathRoot = "."
	// use the http.NewServerMux() function to create an empty servemux
	mux := http.NewServeMux()


	// Initialise the timeHandler in exactly the same way we would any normal
	// struct.
	th := http.HandlerFunc(timeHandler)
	re := http.HandlerFunc(readinessEndpoint)
	hw := http.HandlerFunc(helloHandler)

	// Use the http.FileServer() function to create a handler
	//	fs := http.FileServer(http.Dir(filepathRoot))
	rh := http.RedirectHandler("http://example.org", 307)	
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/foo", rh)
	mux.Handle("/time", th)	
	mux.Handle("/healthz", re)
	mux.Handle("/hello", hw)
	
	s := &http.Server{
	Addr:           ":" + port,
	Handler:        mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)	
	log.Fatal(s.ListenAndServe())	

}
