package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	auth "github.com/ericsprague3/chirpy/internal/auth"
	"github.com/ericsprague3/chirpy/internal/database"
	godotenv "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	tokenSecret    string
	polkaKey       string
}

const DEFAULT_EXPIRES int = int(time.Hour / time.Second)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("middlewareMetricsInc hit")
		cfg.fileserverHits.Add(1)

		// No Cache force browser to revalidate
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		next.ServeHTTP(w, r)
	})
}

type contextKey string

const userIDKey contextKey = "userID"

func (cfg *apiConfig) middlewareAuthenticatedRoute(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract Bearer token from Authentication Header
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. Validate the JWT and retrieve the user's UUID
		userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 3. Attach the authenticated userID to the request context
		ctx := context.WithValue(r.Context(), userIDKey, userID)

		// 4. Pass control to the protected handler with the updated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cfg *apiConfig) HandleResets(resp http.ResponseWriter, req *http.Request) {
	// Delete Users
	err := cfg.HandleDeleteUsers(resp, req)
	if err != nil {
		log.Printf("Error deleting users %v\n", err)
		resp.WriteHeader(500)
		return
	}

	cfg.fileserverHits.Store(0)
	resp.Header().Add("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(200)
	fmt.Fprintf(resp, "Hits: %d", cfg.fileserverHits.Load())
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to postgres database: %v", err)
		os.Exit(1)
	}
	tokenSecret := os.Getenv("TOKEN_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
		tokenSecret:    tokenSecret,
		polkaKey:       polkaKey,
	}
	serveMux := http.NewServeMux()
	port := ":8080"
	server := http.Server{
		Addr:    port,
		Handler: serveMux,
	}

	fsHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))

	// File App Routes
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fsHandler))
	// API Routes
	serveMux.HandleFunc("GET /api/healthz", HandleHealth)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.HandleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.HandleGetChirpById)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.middlewareAuthenticatedRoute(apiCfg.HandleDeleteChirp))
	serveMux.HandleFunc("POST /api/chirps", apiCfg.middlewareAuthenticatedRoute(apiCfg.HandleCreateChirp))
	serveMux.HandleFunc("POST /api/login", apiCfg.HandleLoginUser)
	serveMux.HandleFunc("POST /api/polka/webhooks", apiCfg.HandleWebHookUserUpgraded)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.HandleRefresh)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.HandleRevoke)
	serveMux.HandleFunc("POST /api/users", apiCfg.HandleCreateUser)
	serveMux.HandleFunc("PUT /api/users", apiCfg.middlewareAuthenticatedRoute(apiCfg.HandleUpdateUser))
	// Admin Routes
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.HandleResets)

	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("API Started. Listening on localhost%s\n", port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Fatalf("Server error: %v", err)
	case <-ctx.Done():
		log.Println("Shutdown signal received....")
		stop()
	}

	shudownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shudownCtx); err != nil {
		log.Printf("Forced shutdown due to timeout: %v", err)
	}

	log.Println("Closing Postgres connection pool...")
	if err := db.Close(); err != nil {
		log.Printf("Error closing Postgres: %v", err)
	}

	log.Println("Server exited cleanly")
}
