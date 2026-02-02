package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"
	"fmt"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
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

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) reportMetrics(w http.ResponseWriter, r *http.Request){
	count := fmt.Sprintf("%d", cfg.fileserverHits.Load())
	w.Write([]byte("Hits: " + count + "\n"))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
	count := fmt.Sprintf("%d", cfg.fileserverHits.Load())	
	w.Write([]byte("Hits: " + count + "\n"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("Going through middleware: Total hits: %v\n", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

func main() {
	const port = "8080"
	const filepathRoot = "."

	config := apiConfig{}
	// use the http.NewServerMux() function to create an empty servemux
	mux := http.NewServeMux()

	// Initialise the timeHandler in exactly the same way we would any normal
	// struct.
	th := http.HandlerFunc(timeHandler)
	re := http.HandlerFunc(readinessEndpoint)
	hw := http.HandlerFunc(helloHandler)
	rm := http.HandlerFunc(config.reportMetrics)
	reset := http.HandlerFunc(config.resetHits)
	
	// Use the http.FileServer() function to create a handler
	//	fs := http.FileServer(http.Dir(filepathRoot))
	rh := http.RedirectHandler("http://example.org", 307)
	mux.Handle("/app/", http.StripPrefix("/app/", config.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("/foo", rh)
	mux.Handle("/time", th)
	mux.Handle("/healthz", re)
	mux.Handle("/hello", hw)
	mux.Handle("/metrics", rm)
	mux.Handle("/reset", reset)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(s.ListenAndServe())

}
