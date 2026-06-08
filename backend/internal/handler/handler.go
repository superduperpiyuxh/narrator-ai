package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/graylog"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/normalizer"
)

type Handler struct {
	db          *database.DB
	graylogClient *graylog.Client
}

func New(db *database.DB, gc *graylog.Client) *Handler {
	return &Handler{db: db, graylogClient: gc}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetEvents(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 1000 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	events, total, err := h.db.GetEvents(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) SearchEvents(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}
	if len(query) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too long (max 200 chars)"})
		return
	}

	events, err := h.db.SearchEvents(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

func (h *Handler) GetEventsByHost(c *gin.Context) {
	hostname := c.Param("hostname")
	if hostname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hostname is required"})
		return
	}

	events, err := h.db.GetEventsByHost(hostname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

func (h *Handler) GetEventsByType(c *gin.Context) {
	eventType := c.Param("eventType")
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventType is required"})
		return
	}

	events, err := h.db.GetEventsByType(eventType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.db.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) SyncFromGraylog(c *gin.Context) {
	batchSize := 500
	totalSynced := 0

	err := h.graylogClient.FetchAllEvents("*", batchSize, func(events []graylog.Event) error {
		normalized := make([]database.Event, len(events))
		for i, e := range events {
			n := normalizer.NormalizeEvent(e)
			normalized[i] = toDBEvent(n)
		}

		if err := h.db.InsertEvents(normalized); err != nil {
			return err
		}
		totalSynced += len(normalized)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"synced": totalSynced})
}

func toDBEvent(e graylog.Event) database.Event {
	return database.Event{
		Timestamp:     e.Timestamp,
		Hostname:      e.Hostname,
		EventType:     e.EventType,
		EventID:       e.EventID,
		UserName:      e.User,
		SourceIP:      e.SourceIP,
		DestIP:        e.DestIP,
		ProcessName:   e.ProcessName,
		CommandLine:   e.CommandLine,
		ParentProcess: e.ParentProcess,
		LogType:       e.LogType,
		SessionID:     e.SessionID,
		Department:    e.Department,
		Location:      e.Location,
		DeviceType:    e.DeviceType,
		Success:       e.Success,
		Port:          e.Port,
		Protocol:      e.Protocol,
		FilePath:      e.FilePath,
		Severity:      e.Severity,
		Error:         e.Error,
		RawJSON:       e.RawJSON,
	}
}
