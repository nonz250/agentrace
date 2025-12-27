package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/satetsu888/agentrace/server/internal/api"
	"github.com/satetsu888/agentrace/server/internal/config"
	"github.com/satetsu888/agentrace/server/internal/repository/memory"
)

func main() {
	cfg := config.Load()

	// Initialize repositories (memory for Step 1)
	repos := memory.NewRepositories()

	// Create router
	router := api.NewRouter(cfg, repos)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("DB Type: %s", cfg.DBType)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
