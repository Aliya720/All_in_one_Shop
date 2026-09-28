package main

import (
	"log"
	"net/http"
	"os"

	"ecommerce/backend/internal/auth"
	"ecommerce/backend/internal/config"
	"ecommerce/backend/internal/database"
	"ecommerce/backend/internal/handlers"
	custommw "ecommerce/backend/internal/middleware"
	"ecommerce/backend/internal/models"
	"ecommerce/backend/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	if err := database.EnsureUploadDir(cfg.UploadDir); err != nil {
		log.Fatalf("could not create upload directory: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("connected to database")

	migrationsDir := getEnvOrDefault("MIGRATIONS_DIR", "./migrations")
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	if cfg.AppEnv != "production" || os.Getenv("SEED_ON_START") == "true" {
		seedFile := getEnvOrDefault("SEED_FILE", "./seed/seed.sql")
		if err := database.Seed(db, seedFile); err != nil {
			log.Printf("warning: seeding failed: %v", err)
		}
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db, cartRepo)

	ensureAdminUser(userRepo, cfg.AdminEmail, cfg.AdminPassword)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	productHandler := handlers.NewProductHandler(productRepo, cfg)
	cartHandler := handlers.NewCartHandler(cartRepo, productRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo)
	adminHandler := handlers.NewAdminHandler(userRepo)

	router := mux.NewRouter()
	router.Use(custommw.Recover)
	router.Use(custommw.Logging)
	router.Use(custommw.CORS(cfg.AllowedOrigins))

	router.HandleFunc("/healthz", handlers.Health).Methods(http.MethodGet)

	// Serve uploaded product images.
	router.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))),
	)

	api := router.PathPrefix("/api").Subrouter()

	// --- Public auth routes ---
	api.HandleFunc("/auth/register", authHandler.Register).Methods(http.MethodPost)
	api.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)

	// --- Public catalog routes ---
	api.HandleFunc("/categories", categoryHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/categories/{slug}", categoryHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/products", productHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/products/{slug}", productHandler.Get).Methods(http.MethodGet)

	// --- Authenticated routes ---
	authRequired := custommw.RequireAuth(cfg.JWTSecret)

	authSub := api.PathPrefix("").Subrouter()
	authSub.Use(authRequired)

	authSub.HandleFunc("/auth/me", authHandler.Me).Methods(http.MethodGet)

	authSub.HandleFunc("/cart", cartHandler.Get).Methods(http.MethodGet)
	authSub.HandleFunc("/cart/items", cartHandler.AddItem).Methods(http.MethodPost)
	authSub.HandleFunc("/cart/items/{productId}", cartHandler.UpdateItem).Methods(http.MethodPut)
	authSub.HandleFunc("/cart/items/{productId}", cartHandler.RemoveItem).Methods(http.MethodDelete)
	authSub.HandleFunc("/cart", cartHandler.Clear).Methods(http.MethodDelete)

	authSub.HandleFunc("/orders/checkout", orderHandler.Checkout).Methods(http.MethodPost)
	authSub.HandleFunc("/orders", orderHandler.ListMine).Methods(http.MethodGet)
	authSub.HandleFunc("/orders/{id}", orderHandler.GetMine).Methods(http.MethodGet)

	// --- Admin routes ---
	adminSub := api.PathPrefix("/admin").Subrouter()
	adminSub.Use(authRequired)
	adminSub.Use(custommw.RequireAdmin)

	adminSub.HandleFunc("/categories", categoryHandler.Create).Methods(http.MethodPost)
	adminSub.HandleFunc("/categories/{id}", categoryHandler.Update).Methods(http.MethodPut)
	adminSub.HandleFunc("/categories/{id}", categoryHandler.Delete).Methods(http.MethodDelete)

	adminSub.HandleFunc("/products", productHandler.ListAdmin).Methods(http.MethodGet)
	adminSub.HandleFunc("/products", productHandler.Create).Methods(http.MethodPost)
	adminSub.HandleFunc("/products/{id}", productHandler.Update).Methods(http.MethodPut)
	adminSub.HandleFunc("/products/{id}", productHandler.Delete).Methods(http.MethodDelete)
	adminSub.HandleFunc("/products/{id}/image", productHandler.UploadImage).Methods(http.MethodPost)

	adminSub.HandleFunc("/orders", orderHandler.ListAll).Methods(http.MethodGet)
	adminSub.HandleFunc("/orders/{id}", orderHandler.GetAny).Methods(http.MethodGet)
	adminSub.HandleFunc("/orders/{id}/status", orderHandler.UpdateStatus).Methods(http.MethodPut)

	adminSub.HandleFunc("/users", adminHandler.ListUsers).Methods(http.MethodGet)
	adminSub.HandleFunc("/stats", orderHandler.Stats).Methods(http.MethodGet)

	// Optional: serve the static frontend directly from the Go server.
	// Handy for local development without Docker/Nginx. Must be
	// registered last so it never shadows /api or /uploads.
	if frontendDir := os.Getenv("FRONTEND_DIR"); frontendDir != "" {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir(frontendDir)))
		log.Printf("serving frontend from %s", frontendDir)
	}

	addr := ":" + cfg.AppPort
	log.Printf("ecommerce API listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ensureAdminUser creates a default admin account on first boot if one
// does not already exist, using ADMIN_EMAIL / ADMIN_PASSWORD from config.
func ensureAdminUser(users *repository.UserRepository, email, password string) {
	if _, err := users.FindByEmail(email); err == nil {
		return // already exists
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("warning: could not hash admin password: %v", err)
		return
	}

	if _, err := users.Create("Administrator", email, hash, models.RoleAdmin); err != nil {
		log.Printf("warning: could not create default admin user: %v", err)
		return
	}

	log.Printf("default admin user created: %s", email)
}
