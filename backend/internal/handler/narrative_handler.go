package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/llm"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/narrative"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/security"
)

type NarrativeHandler struct {
	db      *database.DB
	llmKey  string
}

func NewNarrativeHandler(db *database.DB, llmKey string) *NarrativeHandler {
	return &NarrativeHandler{db: db, llmKey: llmKey}
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

	llmClient := llm.NewClient(h.llmKey)
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
		Summary:          narr.Summary,
		Confidence:       narr.Confidence,
		Sentences:        string(sentencesJSON),
		ModelUsed:        "anthropic/claude-3-opus",
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
