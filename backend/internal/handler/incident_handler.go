package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/attck"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/auth"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/incident"
)

type IncidentHandler struct {
	db *database.DB
}

func NewIncidentHandler(db *database.DB) *IncidentHandler {
	return &IncidentHandler{db: db}
}

func (h *IncidentHandler) GroupIncidents(c *gin.Context) {
	start := time.Now()

	events, err := h.db.GetUnprocessedEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	groups := incident.GroupEventsIntoIncidents(events, 15)

	techniques := attck.AllTechniques()
	techRefs := make([]database.TechniqueRef, len(techniques))
	for i, t := range techniques {
		techRefs[i] = database.TechniqueRef{
			TechniqueID: t.ID,
			Name:        t.Name,
			Tactic:      t.Tactic,
		}
	}
	if err := h.db.SeedTechniques(techRefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	incidentsCreated := 0
	for _, group := range groups {
		inc := incident.BuildIncidentFromGroup(group)
		eventIDs := make([]int64, len(group))
		for i, e := range group {
			eventIDs[i] = e.ID
		}
		if err := h.db.CreateIncident(&inc, eventIDs); err != nil {
			continue
		}
		incidentsCreated++
	}

	duration := time.Since(start)
	c.JSON(http.StatusOK, gin.H{
		"grouped":          len(events),
		"incidents_created": incidentsCreated,
		"duration":         duration.String(),
	})
}

func (h *IncidentHandler) GetIncidents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	severity := c.Query("severity")
	status := c.Query("status")
	sourceIP := c.Query("source_ip")

	userID := auth.GetUserID(c)
	incidents, total, err := h.db.GetIncidentsByUserID(userID, limit, offset, severity, status, sourceIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"incidents": incidents,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *IncidentHandler) GetIncident(c *gin.Context) {
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

	// Check ownership
	userID := auth.GetUserID(c)
	if inc.UserID != "" && inc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incident": inc})
}

func (h *IncidentHandler) GetIncidentEvents(c *gin.Context) {
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

	// Check ownership
	userID := auth.GetUserID(c)
	if inc.UserID != "" && inc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	events, err := h.db.GetIncidentEvents(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":      events,
		"total":       len(events),
		"incident_id": id,
	})
}

func (h *IncidentHandler) GetIncidentStats(c *gin.Context) {
	userID := auth.GetUserID(c)
	stats, err := h.db.GetIncidentStatsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *IncidentHandler) GetTechniques(c *gin.Context) {
	techniques := attck.AllTechniques()
	c.JSON(http.StatusOK, gin.H{"techniques": techniques})
}
