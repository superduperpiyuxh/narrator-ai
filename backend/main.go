package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/auth"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/config"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/handler"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer db.Close()

	authSvc := auth.NewService(db.Conn(), cfg.JWTSecret)
	adminPass := getEnvOrDefault("ADMIN_PASSWORD", "admin123")
	if err := authSvc.CreateDefaultAdmin("admin@nexus.ai", adminPass); err == nil {
		log.Printf("Created default admin: admin@nexus.ai / %s", adminPass)
	}
	authH := handler.NewAuthHandler(authSvc)

	h := handler.New(db, cfg.DataDir)
	ih := handler.NewIncidentHandler(db)
	nh := handler.NewNarrativeHandler(db, authSvc, cfg.OpenRouterKey)
	fh := handler.NewFeedbackHandler(db)
	ingestH := handler.NewIngestHandler(db, cfg.DataDir)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	r.GET("/health", h.Health)
	r.POST("/api/auth/signup", authH.Signup)
	r.POST("/api/auth/login", authH.Login)

	// Rate limiter: 5 narrative generations per minute per user
	narrativeRateLimiter := auth.NewRateLimiter(5, time.Minute)

	// Protected routes
	protected := r.Group("")
	protected.Use(auth.AuthMiddleware(authSvc))
	{
		protected.GET("/api/auth/me", authH.Me)
		protected.GET("/api/auth/settings", authH.GetSettings)
		protected.PUT("/api/auth/settings", authH.UpdateSettings)

		protected.GET("/api/events", h.GetEvents)
		protected.GET("/api/events/search", h.SearchEvents)
		protected.GET("/api/events/host/:hostname", h.GetEventsByHost)
		protected.GET("/api/events/type/:eventType", h.GetEventsByType)
		protected.GET("/api/stats", h.GetStats)
		protected.POST("/api/import", h.ImportLocal)

		protected.POST("/api/incidents/group", ih.GroupIncidents)
		protected.GET("/api/incidents/stats", ih.GetIncidentStats)
		protected.GET("/api/incidents", ih.GetIncidents)
		protected.GET("/api/incidents/:id", ih.GetIncident)
		protected.GET("/api/incidents/:id/events", ih.GetIncidentEvents)
		protected.GET("/api/techniques", ih.GetTechniques)
		protected.GET("/api/techniques/counts", ih.GetTechniqueCounts)

		protected.POST("/api/incidents/:id/narrative", auth.RateLimitMiddleware(narrativeRateLimiter), nh.GenerateNarrative)
		protected.GET("/api/incidents/:id/narrative", nh.GetNarrative)
		protected.GET("/api/narratives/:id", nh.GetNarrativeSourceEvents)

		protected.POST("/api/feedback", fh.SubmitFeedback)
		protected.GET("/api/feedback/:narrative_id", fh.GetFeedback)

		protected.POST("/api/v1/ingest", ingestH.IngestEvents)
		protected.POST("/api/v1/ingest/file", ingestH.IngestFile)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 130 * time.Second,
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

func generateRandomPassword(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate random password: %v", err)
	}
	return hex.EncodeToString(b)[:length]
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
