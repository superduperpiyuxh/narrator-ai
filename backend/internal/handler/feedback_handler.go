package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

type FeedbackHandler struct {
	db *database.DB
}

func NewFeedbackHandler(db *database.DB) *FeedbackHandler {
	return &FeedbackHandler{db: db}
}

func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	var req struct {
		NarrativeID int64  `json:"narrative_id"`
		IncidentID  int64  `json:"incident_id"`
		Rating      int    `json:"rating"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if req.NarrativeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "narrative_id is required and must be positive"})
		return
	}
	if req.IncidentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incident_id is required and must be positive"})
		return
	}
	if req.Rating != -1 && req.Rating != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be -1 or 1"})
		return
	}

	// Sanitize notes
	req.Notes = strings.TrimSpace(req.Notes)
	if len(req.Notes) > 1000 {
		req.Notes = req.Notes[:1000]
	}

	feedback := &database.Feedback{
		NarrativeID: req.NarrativeID,
		IncidentID:  req.IncidentID,
		Rating:      req.Rating,
		Notes:       req.Notes,
	}

	if err := h.db.CreateFeedback(feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"feedback": feedback})
}

func (h *FeedbackHandler) GetFeedback(c *gin.Context) {
	narrativeID, err := strconv.ParseInt(c.Param("narrative_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid narrative id"})
		return
	}

	feedback, err := h.db.GetFeedbackByNarrativeID(narrativeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"feedback": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve feedback"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"feedback": feedback})
}
