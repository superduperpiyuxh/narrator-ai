package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/auth"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/llm"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/narrative"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/security"
)

type NarrativeHandler struct {
	db      *database.DB
	authSvc *auth.Service
	llmKey  string
}

func NewNarrativeHandler(db *database.DB, authSvc *auth.Service, llmKey string) *NarrativeHandler {
	return &NarrativeHandler{db: db, authSvc: authSvc, llmKey: llmKey}
}

func (h *NarrativeHandler) GenerateNarrative(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident id"})
		return
	}

	inc, err := h.db.GetIncidentByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
		return
	}

	existing, _ := h.db.GetNarrativeByIncidentID(id)
	if existing != nil {
		c.JSON(http.StatusOK, gin.H{"narrative": existing, "cached": true})
		return
	}

	if h.llmKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OPENROUTER_API_KEY not configured"})
		return
	}

	if security.DetectInjection(inc.Title) || security.DetectInjection(inc.Description) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "potentially malicious input detected"})
		return
	}

	events, err := h.db.GetIncidentEvents(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use user's OpenRouter key if available, fall back to global
	userID := auth.GetUserID(c)
	llmKey := h.llmKey
	if user, _ := h.authSvc.GetUserByID(userID); user != nil && user.OpenRouterKey != "" {
		llmKey = user.OpenRouterKey
	}
	llmClient := llm.NewClient(llmKey)
	gen := narrative.NewGenerator(llmClient)

	start := time.Now()
	narr, rawResponse, tokens, err := gen.Generate(*inc, events)
	duration := time.Since(start)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sentencesJSON, _ := json.Marshal(narr.Sentences)

	dbNarr := &database.Narrative{
		IncidentID:       id,
		UserID:           userID,
		Summary:          narr.Summary,
		Confidence:       narr.Confidence,
		Sentences:        string(sentencesJSON),
		ModelUsed:        "openrouter-rotating",
		Temperature:      0.2,
		TokensUsed:       tokens,
		GenerationTimeMs: duration.Milliseconds(),
	}

	if err := h.db.CreateNarrative(dbNarr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = rawResponse

	c.JSON(http.StatusOK, gin.H{
		"narrative": dbNarr,
		"cached":    false,
	})
}

func (h *NarrativeHandler) GetNarrative(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid incident id"})
		return
	}

	// Check incident ownership
	inc, err := h.db.GetIncidentByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "incident not found"})
		return
	}
	userID := auth.GetUserID(c)
	if inc.UserID != "" && inc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	narr, err := h.db.GetNarrativeByIncidentID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if narr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "narrative not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"narrative": narr})
}

func (h *NarrativeHandler) GetNarrativeSourceEvents(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid narrative id"})
		return
	}

	// Check narrative ownership through incident
	narr, err := h.db.GetNarrativeByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if narr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "narrative not found"})
		return
	}
	inc, err := h.db.GetIncidentByID(narr.IncidentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inc != nil {
		userID := auth.GetUserID(c)
		if inc.UserID != "" && inc.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	events, err := h.db.GetNarrativeSourceEvents(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":      events,
		"total":       len(events),
		"narrative_id": id,
	})
}
