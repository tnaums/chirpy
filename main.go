package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/tnaums/chirpy/internal/auth"
	"github.com/tnaums/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

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

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	} else {
		cleaned := cleanChirp(params.Body)
		respondWithCleaned(w, 200, cleaned)
	}
}

func cleanChirp(msg string) string {
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
	queries        *database.Queries
	platform       string
	secretPhrase   string
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

func (cfg *apiConfig) resetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		log.Printf("Operation not allowed")
		w.WriteHeader(403)
	}
	err := cfg.queries.DeleteUsers(context.Background())
	if err != nil {
		log.Printf("couldn't delete users: %w", err)
	}
	w.Write([]byte("Database reset successfully!\n"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("Going through middleware: Total hits: %v\n", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) chirpById(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	id := r.PathValue("chirpID")
	uid, err := uuid.Parse(id)
	if err != nil {
		log.Printf("couldn't create uid from id")
	}

	c, err := cfg.queries.ChirpByID(context.Background(), uid)
	if err != nil {
		log.Printf("Error retrieving chirp by id: %s", err)
		w.WriteHeader(404)
		return
	}

	mainChirp := Chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}

	dat, err := json.MarshalIndent(mainChirp, "", " ")
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dat))

}

func (cfg *apiConfig) chirpGet(w http.ResponseWriter, r *http.Request) {
	var convertedChirps []Chirp
	allChirps, err := cfg.queries.ListChirps(context.Background())
	if err != nil {
		log.Printf("couldn't retrieve chirps: %w", err)
	}

	for _, c := range allChirps {
		fmt.Println(c.CreatedAt)

		mainChirp := Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		}
		convertedChirps = append(convertedChirps, mainChirp)
	}
	fmt.Println(convertedChirps[0])

	dat, err := json.MarshalIndent(convertedChirps, "", " ")
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dat))
}

func (cfg *apiConfig) chirpSave(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	// validate that chirp is not too long
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// Add chirp to the chirps table
	newChirp, err := cfg.queries.CreateChirp(context.Background(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	})
	if err != nil {
		log.Printf("couldn't create feed follow: %w", err)
	}

	mainChirp := Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	}

	dat, err := json.MarshalIndent(mainChirp, "", " ")
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dat))
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	luser, err := cfg.queries.GetUserByEmail(context.Background(), params.Email)
	if err != nil {
		log.Println("Incorrect email or password")
		w.WriteHeader(401)
		return
	}

	check, err2 := auth.CheckPasswordHash(params.Password, luser.HashedPassword)
	if check == false {
		log.Println("Incorrect email or password")
		w.WriteHeader(401)
		return
	}
	if err2 != nil {
		log.Printf("Error checking password hash")
		return
	}

	mainUser := User{
		ID:        luser.ID,
		CreatedAt: luser.CreatedAt,
		UpdatedAt: luser.UpdatedAt,
		Email:     luser.Email,
	}

	dat, err := json.MarshalIndent(mainUser, "", " ")
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dat))

}

func (cfg *apiConfig) registerUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	// change password from plain text to hashed version
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error creating password hash: %w", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.queries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	})
	if err != nil {
		log.Printf("couldn't create user: %w", err)
		w.WriteHeader(500)
		return
	}

	mainUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	dat, err := json.MarshalIndent(mainUser, "", " ")
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dat))

}

func main() {
	const port = "8080"
	const filepathRoot = "."

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	pf := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	config := apiConfig{
		queries:      dbQueries,
		platform:     pf,
		secretPhrase: secret,
	}
	// use the http.NewServerMux() function to create an empty servemux
	mux := http.NewServeMux()

	// Initialise the timeHandler in exactly the same way we would any normal
	// struct.
	th := http.HandlerFunc(timeHandler)
	re := http.HandlerFunc(readinessEndpoint)
	hw := http.HandlerFunc(helloHandler)
	rm := http.HandlerFunc(config.reportMetrics)
	//	reset := http.HandlerFunc(config.resetHits)
	resetdb := http.HandlerFunc(config.resetUsers)
	valchirp := http.HandlerFunc(validateChirp)
	ru := http.HandlerFunc(config.registerUser)
	chirpsv := http.HandlerFunc(config.chirpSave)
	chirpget := http.HandlerFunc(config.chirpGet)
	chirpbyid := http.HandlerFunc(config.chirpById)
	login := http.HandlerFunc(config.loginUser)

	// Use the http.FileServer() function to create a handler
	//	fs := http.FileServer(http.Dir(filepathRoot))
	rh := http.RedirectHandler("http://example.org", 307)
	mux.Handle("/app/", http.StripPrefix("/app/", config.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("/api/foo", rh)
	mux.Handle("/time", th)
	mux.Handle("GET /api/healthz", re)
	mux.Handle("/hello", hw)
	mux.Handle("GET /admin/metrics", rm)
	mux.Handle("POST /admin/reset", resetdb)
	mux.Handle("POST /api/validate_chirp", valchirp)
	mux.Handle("POST /api/users", ru)
	mux.Handle("POST /api/chirps", chirpsv)
	mux.Handle("GET /api/chirps", chirpget)
	mux.Handle("GET /api/chirps/{chirpID}", chirpbyid)
	mux.Handle("POST /api/login", login)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(s.ListenAndServe())

}
