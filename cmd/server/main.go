package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"expenses-backend/internal/config"
	"expenses-backend/internal/db"
	"expenses-backend/internal/http/router"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		log.Fatalf("db init failed: %v", err)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	router.Setup(r, cfg)

	addr := ":" + cfg.AppPort
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
