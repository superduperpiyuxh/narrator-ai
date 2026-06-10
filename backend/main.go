package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/config"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/graylog"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/handler"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer db.Close()

	graylogClient := graylog.NewClient(cfg.GraylogURL, cfg.GraylogUser, cfg.GraylogPass)

	h := handler.New(db, graylogClient, cfg.DataDir)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", h.Health)
	r.GET("/api/events", h.GetEvents)
	r.GET("/api/events/search", h.SearchEvents)
	r.GET("/api/events/host/:hostname", h.GetEventsByHost)
	r.GET("/api/events/type/:eventType", h.GetEventsByType)
	r.GET("/api/stats", h.GetStats)
	r.POST("/api/sync", h.SyncFromGraylog)
	r.POST("/api/import", h.ImportLocal)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
