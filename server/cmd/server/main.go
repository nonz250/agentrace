package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/satetsu888/agentrace/server/internal/api"
	"github.com/satetsu888/agentrace/server/internal/config"
	"github.com/satetsu888/agentrace/server/internal/repository"
	"github.com/satetsu888/agentrace/server/internal/repository/memory"
	"github.com/satetsu888/agentrace/server/internal/repository/mongodb"
	"github.com/satetsu888/agentrace/server/internal/repository/postgres"
	"github.com/satetsu888/agentrace/server/internal/repository/sqlite"
	"github.com/satetsu888/agentrace/server/internal/repository/turso"
)

func main() {
	cfg := config.Load()

	// Initialize repositories based on DB_TYPE
	repos, closer, err := initRepositories(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize repositories: %v", err)
	}
	if closer != nil {
		defer closer.Close()
	}

	// Create router
	router := api.NewRouter(cfg, repos)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("DB Type: %s", cfg.DBType)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initRepositories(cfg *config.Config) (*repository.Repositories, io.Closer, error) {
	switch cfg.DBType {
	case "memory":
		log.Println("Using in-memory database")
		return memory.NewRepositories(), nil, nil

	case "sqlite":
		log.Printf("Using SQLite database: %s", cfg.DatabaseURL)
		if cfg.DatabaseURL == "" {
			return nil, nil, fmt.Errorf("DATABASE_URL is required for sqlite")
		}
		db, err := sqlite.Open(cfg.DatabaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open SQLite database: %w", err)
		}
		return sqlite.NewRepositories(db), db, nil

	case "postgres":
		log.Printf("Using PostgreSQL database: %s", cfg.DatabaseURL)
		if cfg.DatabaseURL == "" {
			return nil, nil, fmt.Errorf("DATABASE_URL is required for postgres")
		}
		db, err := postgres.Open(cfg.DatabaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
		}
		return postgres.NewRepositories(db), db, nil

	case "mongodb":
		log.Printf("Using MongoDB database: %s", cfg.DatabaseURL)
		if cfg.DatabaseURL == "" {
			return nil, nil, fmt.Errorf("DATABASE_URL is required for mongodb")
		}
		db, err := mongodb.Open(cfg.DatabaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open MongoDB database: %w", err)
		}
		return mongodb.NewRepositories(db), db, nil

	case "turso":
		log.Printf("Using Turso database: %s", cfg.DatabaseURL)
		if cfg.DatabaseURL == "" {
			return nil, nil, fmt.Errorf("DATABASE_URL is required for turso")
		}
		db, err := turso.Open(cfg.DatabaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open Turso database: %w", err)
		}
		return turso.NewRepositories(db), db, nil

	default:
		return nil, nil, fmt.Errorf("unknown database type: %s", cfg.DBType)
	}
}
