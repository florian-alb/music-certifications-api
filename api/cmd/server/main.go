// @title           Music Certifications API
// @version         1.0
// @description     Query music certifications (gold, platinum, diamond) from SNEP and other sources.
// @host            localhost:8000
// @BasePath        /
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            X-API-Key
package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/florian-alb/music-certifications/docs"
	"github.com/florian-alb/music-certifications/internal/cache"
	"github.com/florian-alb/music-certifications/internal/handler"
	"github.com/florian-alb/music-certifications/internal/middleware"
	"github.com/florian-alb/music-certifications/internal/repository"
	"github.com/florian-alb/music-certifications/internal/service"
	"github.com/florian-alb/music-certifications/pkg/database"
	httpswagger "github.com/swaggo/http-swagger"
)

func main() {
	dbURL := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/musiccerts?sslmode=disable")
	redisURL := env("REDIS_URL", "redis://localhost:6379")
	port := env("PORT", "8000")

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	rdb, err := cache.New(redisURL)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}

	// Repositories
	certRepo := repository.NewCertificationRepo(db)
	userRepo := repository.NewUserRepo(db)
	apiKeyRepo := repository.NewAPIKeyRepo(db)

	// Services
	certSvc := service.NewCertificationService(certRepo)
	authSvc := service.NewAuthService(userRepo, apiKeyRepo)

	// Handlers
	certH := handler.NewCertificationHandler(certSvc)
	authH := handler.NewAuthHandler(authSvc)

	// Middleware chain: auth → ratelimit → handler
	authMw := middleware.NewAuthMiddleware(authSvc)
	rlMw := middleware.NewRateLimitMiddleware(rdb)
	protect := func(h http.Handler) http.Handler {
		return authMw.Wrap(rlMw.Wrap(h))
	}

	mux := http.NewServeMux()

	// Swagger UI
	mux.Handle("/swagger/", httpswagger.WrapHandler)

	// Public
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /v1/auth/register", authH.Register)
	mux.HandleFunc("POST /v1/auth/login", authH.Login)

	// Protected (require X-API-Key + rate limit)
	mux.Handle("GET /v1/certifications", protect(http.HandlerFunc(certH.List)))
	mux.Handle("GET /v1/certifications/{id}", protect(http.HandlerFunc(certH.GetByID)))
	mux.Handle("GET /v1/artists/{id}/certifications", protect(http.HandlerFunc(certH.ListByArtist)))
	mux.Handle("GET /v1/releases/{id}/certifications", protect(http.HandlerFunc(certH.ListByRelease)))
	mux.Handle("GET /v1/search", protect(http.HandlerFunc(certH.Search)))
	mux.Handle("GET /v1/sources", protect(http.HandlerFunc(certH.ListSources)))
	mux.Handle("GET /v1/me/apikey", protect(http.HandlerFunc(authH.GetMyKey)))

	log.Printf("Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(mux)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
