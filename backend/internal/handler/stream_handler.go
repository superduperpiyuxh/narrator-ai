package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/auth"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/incident"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/normalizer"
)

type StreamHandler struct {
	db          *database.DB
	mu          sync.RWMutex
	clients     map[chan database.Event]struct{}
	eventBuffer []database.Event
}

func NewStreamHandler(db *database.DB) *StreamHandler {
	return &StreamHandler{
		db:      db,
		clients: make(map[chan database.Event]struct{}),
	}
}

// SSEStream handles Server-Sent Events for real-time event streaming
func (h *StreamHandler) SSEStream(c *gin.Context) {
	userID := auth.GetUserID(c)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	eventChan := make(chan database.Event, 100)

	h.mu.Lock()
	h.clients[eventChan] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, eventChan)
		h.mu.Unlock()
		close(eventChan)
	}()

	c.SSEvent("connected", gin.H{
		"user_id": userID,
		"message": "connected to event stream",
	})
	flusher.Flush()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			c.SSEvent("event", json.RawMessage(data))
			flusher.Flush()
		}
	}
}

// BroadcastEvent sends an event to all connected SSE clients
func (h *StreamHandler) BroadcastEvent(event database.Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- event:
		default:
			// client too slow, drop event
		}
	}
}

// IngestStream accepts real-time events via POST and broadcasts to SSE clients
func (h *StreamHandler) IngestStream(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "events array required"})
		return
	}

	if len(req.Events) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no events provided"})
		return
	}

	if len(req.Events) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max 1000 events per stream request"})
		return
	}

	batch := make([]database.Event, 0, len(req.Events))
	for _, ie := range req.Events {
		if ie.Timestamp == "" || ie.Hostname == "" || ie.EventType == "" {
			continue
		}

		rawJSON := ie.RawJSON
		if rawJSON == nil {
			rawJSON = map[string]interface{}{
				"timestamp":  ie.Timestamp,
				"hostname":   ie.Hostname,
				"event_type": ie.EventType,
				"event_id":   ie.EventID,
				"user":       ie.UserName,
				"source_ip":  ie.SourceIP,
			}
		}

		e := normalizer.Event{
			Timestamp:   ie.Timestamp,
			Hostname:    ie.Hostname,
			EventType:   ie.EventType,
			EventID:     ie.EventID,
			User:        ie.UserName,
			SourceIP:    ie.SourceIP,
			DestIP:      ie.DestIP,
			ProcessName: ie.ProcessName,
			CommandLine: ie.CommandLine,
			RawJSON:     rawJSON,
		}
		n := normalizer.NormalizeEvent(e)

		dbEvent := database.Event{
			UserID:      userIDStr,
			Timestamp:   n.Timestamp,
			Hostname:    n.Hostname,
			EventType:   n.EventType,
			EventID:     n.EventID,
			UserName:    n.User,
			SourceIP:    n.SourceIP,
			DestIP:      n.DestIP,
			ProcessName: n.ProcessName,
			CommandLine: n.CommandLine,
			RawJSON:     n.RawJSON,
		}
		batch = append(batch, dbEvent)
	}

	if len(batch) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid events"})
		return
	}

	if err := h.db.InsertEvents(batch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast each event to SSE clients
	for _, event := range batch {
		h.BroadcastEvent(event)
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": len(batch),
		"message":  fmt.Sprintf("streamed %d events", len(batch)),
	})
}

// IngestSingle accepts a single real-time event (simpler API for integrations)
func (h *StreamHandler) IngestSingle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var ie IngestEvent
	if err := c.ShouldBindJSON(&ie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	if ie.Timestamp == "" || ie.Hostname == "" || ie.EventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp, hostname, event_type are required"})
		return
	}

	rawJSON := ie.RawJSON
	if rawJSON == nil {
		rawJSON = map[string]interface{}{
			"timestamp":  ie.Timestamp,
			"hostname":   ie.Hostname,
			"event_type": ie.EventType,
		}
	}

	e := normalizer.Event{
		Timestamp:   ie.Timestamp,
		Hostname:    ie.Hostname,
		EventType:   ie.EventType,
		EventID:     ie.EventID,
		User:        ie.UserName,
		SourceIP:    ie.SourceIP,
		CommandLine: ie.CommandLine,
		RawJSON:     rawJSON,
	}
	n := normalizer.NormalizeEvent(e)

	dbEvent := database.Event{
		UserID:      userIDStr,
		Timestamp:   n.Timestamp,
		Hostname:    n.Hostname,
		EventType:   n.EventType,
		EventID:     n.EventID,
		UserName:    n.User,
		SourceIP:    n.SourceIP,
		CommandLine: n.CommandLine,
		RawJSON:     n.RawJSON,
	}

	if err := h.db.InsertEvents([]database.Event{dbEvent}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.BroadcastEvent(dbEvent)

	c.JSON(http.StatusOK, gin.H{
		"imported": 1,
		"message":  "event ingested",
	})
}

// AutoGroupIncidents groups recent unprocessed events into incidents
// Called after batch ingestion or periodically
func (h *StreamHandler) AutoGroupIncidents(userID string) (int, error) {
	events, err := h.db.GetUnprocessedEvents()
	if err != nil {
		return 0, fmt.Errorf("get unprocessed events: %w", err)
	}

	if len(events) == 0 {
		return 0, nil
	}

	groups := incident.GroupEventsIntoIncidents(events, 15)

	created := 0
	for _, group := range groups {
		inc := incident.BuildIncidentFromGroup(group)
		inc.UserID = userID
		eventIDs := make([]int64, len(group))
		for i, e := range group {
			eventIDs[i] = e.ID
		}
		if err := h.db.CreateIncident(&inc, eventIDs); err != nil {
			continue
		}
		created++
	}

	return created, nil
}
