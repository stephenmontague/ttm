package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/stephenmontague/ttm-tracker/server/internal/api"
	"github.com/stephenmontague/ttm-tracker/server/internal/config"
	"github.com/stephenmontague/ttm-tracker/server/internal/database"
	"github.com/stephenmontague/ttm-tracker/server/internal/repository"
	ttmtemporal "github.com/stephenmontague/ttm-tracker/server/internal/temporal"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	ctx := context.Background()

	// Initialize PostgreSQL
	dbPool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Connected to PostgreSQL")

	if err := database.RunMigrations(ctx, dbPool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Temporal client
	temporalClient, err := ttmtemporal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()
	log.Println("Connected to Temporal")

	// Initialize repositories
	companyRepo := repository.NewCompanyRepository(dbPool)
	authRepo := repository.NewAuthRepository(dbPool)

	// Seed admin user from env vars
	seedEmail := config.GetAdminSeedEmail()
	seedPassword := config.GetAdminSeedPassword()
	if seedEmail != "" && seedPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash admin password: %v", err)
		}
		if err := authRepo.UpsertAdminUser(ctx, seedEmail, string(hash)); err != nil {
			log.Fatalf("Failed to seed admin user: %v", err)
		}
		log.Printf("Admin user seeded: %s", seedEmail)
	} else {
		log.Println("ADMIN_SEED_EMAIL or ADMIN_SEED_PASSWORD not set — skipping admin seed")
	}

	// Initialize handler
	handler := api.NewHandler(temporalClient, companyRepo, authRepo)

	// Setup router
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/api", func(r chi.Router) {
		// Rate limit all API routes: 100 requests per minute per IP
		r.Use(httprate.LimitByIP(100, 1*time.Minute))

		// Public endpoints
		r.Get("/companies", handler.ListCompanies)
		r.Get("/companies/{slug}", handler.GetCompany)
		r.Get("/companies/{slug}/feed", handler.GetCompanyFeed)

		// Auth endpoints (public — no session required)
		r.With(httprate.LimitByIP(5, 1*time.Minute)).Post("/auth/login", handler.PostLogin)
		r.Post("/auth/logout", handler.PostLogout)

		// Admin endpoints (protected by session middleware)
		r.Route("/admin", func(r chi.Router) {
			r.Use(handler.RequireSession)
			r.Get("/auth/status", handler.GetAuthStatus)
			r.Get("/companies", handler.ListAdminCompanies)
			r.Post("/companies", handler.CreateCompany)
			r.Get("/companies/{slug}", handler.GetAdminCompany)
			r.Post("/companies/{slug}/reconcile", handler.ReconcileCompanyStatus)
			r.Post("/companies/{slug}/signal/outreach", handler.SignalOutreach)
			r.Post("/companies/{slug}/signal/contact", handler.SignalAddContact)
			r.Post("/companies/{slug}/signal/contact/remove", handler.SignalRemoveContact)
			r.Post("/companies/{slug}/signal/agent", handler.SignalRequestAgent)
			r.Post("/companies/{slug}/signal/booked", handler.SignalMeetingBooked)
		})
	})

	// Server
	port := config.GetAPIPort()
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("API server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Server stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
