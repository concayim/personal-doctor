package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"personal-doctor/backend/internal/agent"
	"personal-doctor/backend/internal/api"
	"personal-doctor/backend/internal/config"
	"personal-doctor/backend/internal/store"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	doctor, err := agent.NewDoctorAgent(cfg.Agent)
	if err != nil {
		log.Fatalf("create doctor agent: %v", err)
	}

	handler := api.NewServer(db, doctor)
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("personal doctor API listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
