package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/graylog"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/normalizer"
)

var (
	db          *database.DB
	graylogClient *graylog.Client
)

func main() {
	var err error
	db, err = database.New("./narratorai.db")
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	graylogClient = graylog.NewClient("http://localhost:9000", "admin", "admin")

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/api/events", getEvents)
	r.GET("/api/events/search", searchEvents)
	r.GET("/api/events/host/:hostname", getEventsByHost)
	r.GET("/api/events/type/:eventType", getEventsByType)
	r.GET("/api/stats", getStats)
	r.POST("/api/sync", syncFromGraylog)

	r.Run(":8080")
}

func getEvents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	events, total, err := db.GetEvents(limit, offset)
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

func searchEvents(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	events, err := db.SearchEvents(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

func getEventsByHost(c *gin.Context) {
	hostname := c.Param("hostname")
	events, err := db.GetEventsByHost(hostname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

func getEventsByType(c *gin.Context) {
	eventType := c.Param("eventType")
	events, err := db.GetEventsByType(eventType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events, "total": len(events)})
}

func getStats(c *gin.Context) {
	stats, err := db.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func syncFromGraylog(c *gin.Context) {
	batchSize := 500
	totalSynced := 0

	err := graylogClient.FetchAllEvents("*", batchSize, func(events []graylog.Event) error {
		normalized := make([]database.Event, len(events))
		for i, e := range events {
			n := normalizer.NormalizeEvent(e)
			normalized[i] = toDBEvent(n)
		}

		if err := db.InsertEvents(normalized); err != nil {
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

func init() {
	_ = fmt.Sprintf
}
