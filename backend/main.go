package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"shosha_mart/backend/config"
	"shosha_mart/backend/routes"
	"shosha_mart/backend/services"
	syncsvc "shosha_mart/backend/sync"
)

func main() {
	_ = godotenv.Load()
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
