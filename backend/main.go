package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"shosha_mart/backend/config"
	"shosha_mart/backend/routes"
	"shosha_mart/backend/services"
	syncsvc "shosha_mart/backend/sync"
)

func main() {
	// Load .env from current directory or executable directory
	if err := godotenv.Load(); err != nil {
		// Fallback: try loading from executable's directory (production)
		exePath, _ := os.Executable()
		envPath := filepath.Join(filepath.Dir(exePath), ".env")
		_ = godotenv.Load(envPath)
	}

	cfg := config.Load()

	db, err := services.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	worker := syncsvc.NewWorker(db, cfg)
	worker.StartBackground()

	r := gin.Default()
	routes.Register(r, db, cfg, worker)

	log.Printf("sidecar listening on %s (db: %s, sync interval ~%v)", cfg.BindAddr, cfg.DBPath, 5*time.Minute)
	if err := r.Run(cfg.BindAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
