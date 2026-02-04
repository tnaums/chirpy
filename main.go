package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
	"strings"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("And a good day to you!\n"))
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

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) >140 {
		respondWithError(w, 400, "Chirp is too long")
	} else {
		cleaned := cleanChirp(params.Body)
		respondWithCleaned(w, 200, cleaned)
	}
}

func cleanChirp(msg string) string{
	wordSlice := strings.Split(msg, " ")

	for idx, word := range wordSlice {
		lowerString := strings.ToLower(word)		
		if lowerString == "kerfuffle" || lowerString == "sharbert" || lowerString == "fornax" {
			wordSlice[idx] = "****"
		}
	}
	return strings.Join(wordSlice, " ")
}

func respondWithCleaned(w http.ResponseWriter, code int, cleaned string) {
	w.WriteHeader(code)
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}
	respBody := returnVals{
		CleanedBody: cleaned,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}
	w.Header().Set("Content-Type", "application/json")	
	w.Write([]byte(dat))		
}
func respondWithJSON(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	type returnVals struct {
		Valid bool `json:"valid"`
	}
	respBody := returnVals{
		Valid: true,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}
	w.Header().Set("Content-Type", "application/json")	
	w.Write([]byte(dat))	
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	type returnVals struct {
		Error string `json:"error"`
	}
	respBody := returnVals{
		Error: msg,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}
	w.Header().Set("Content-Type", "application/json")	
	w.Write([]byte(dat))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) reportMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	output := "<html>\n"
	output = output + "  <body>\n"
	output = output + "    <h1>Welcome, Chirpy Admin</h1>\n"
	counter := fmt.Sprintf("    <p>Chirpy has been visited %d times!</p>\n", cfg.fileserverHits.Load())
	output = output + counter
	output = output + "  </body>\n"
	output = output + "</html>\n"
	w.Write([]byte(output))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
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
	valchirp := http.HandlerFunc(validateChirp)

	// Use the http.FileServer() function to create a handler
	//	fs := http.FileServer(http.Dir(filepathRoot))
	rh := http.RedirectHandler("http://example.org", 307)
	mux.Handle("/app/", http.StripPrefix("/app/", config.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("/api/foo", rh)
	mux.Handle("/time", th)
	mux.Handle("GET /api/healthz", re)
	mux.Handle("/hello", hw)
	mux.Handle("GET /admin/metrics", rm)
	mux.Handle("POST /admin/reset", reset)
	mux.Handle("POST /api/validate_chirp", valchirp)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(s.ListenAndServe())

}
