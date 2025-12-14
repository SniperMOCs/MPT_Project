package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/SniperMOCs/MPT_Project/internal/config"
	"github.com/SniperMOCs/MPT_Project/internal/user"
)

func openDB(dsn string) (*sql.DB, error) {
	log.Println(dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}
	cfg := config.Load()
	db, err := openDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// User
	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewHandler(userService)

	// HTTP mux
	mux := http.NewServeMux()

	// Route register
	userHandler.RegisterRoutes(mux)

	// Server settings
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server started on %v\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
