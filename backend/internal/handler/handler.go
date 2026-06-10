package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/graylog"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/normalizer"
)

type Handler struct {
	db            *database.DB
	graylogClient *graylog.Client
	dataDir       string
}

func New(db *database.DB, gc *graylog.Client, dataDir string) *Handler {
	return &Handler{db: db, graylogClient: gc, dataDir: dataDir}
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
	batchSize := 5000
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

func (h *Handler) ImportLocal(c *gin.Context) {
	pattern := filepath.Join(h.dataDir, "**", "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(matches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no JSON files found in " + h.dataDir})
		return
	}

	totalSynced := 0
	for _, filePath := range matches {
		count, err := h.importFile(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed %s: %v", filepath.Base(filePath), err)})
			return
		}
		totalSynced += count
	}

	c.JSON(http.StatusOK, gin.H{"synced": totalSynced, "files": len(matches)})
}

func (h *Handler) importFile(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	total := 0
	batch := make([]database.Event, 0, 5000)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		e := mapToEvent(raw)
		n := normalizer.NormalizeEvent(e)
		batch = append(batch, toDBEvent(n))
		total++

		if len(batch) >= 5000 {
			if err := h.db.InsertEvents(batch); err != nil {
				return total, err
			}
			batch = batch[:0]
			fmt.Printf("  %s: %d events imported\n", filepath.Base(filePath), total)
		}
	}

	if len(batch) > 0 {
		if err := h.db.InsertEvents(batch); err != nil {
			return total, err
		}
	}

	return total, scanner.Err()
}

func mapToEvent(raw map[string]interface{}) graylog.Event {
	e := graylog.Event{RawJSON: raw}

	getStr := func(key string) string {
		if v, ok := raw[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	e.Timestamp = getStr("timestamp")
	e.Hostname = getStr("hostname")
	e.EventType = getStr("event_type")
	e.EventID = getStr("event_id")
	e.User = getStr("user")
	e.SourceIP = getStr("source_ip")
	e.DestIP = getStr("destination_ip")
	e.ProcessName = getStr("process_name")
	e.CommandLine = getStr("command_line")
	e.ParentProcess = getStr("parent_process")
	e.LogType = getStr("log_type")
	e.SessionID = getStr("session_id")
	e.Department = getStr("department")
	e.Location = getStr("location")
	e.DeviceType = getStr("device_type")
	e.Port = getStr("port")
	e.Protocol = getStr("protocol")
	e.FilePath = getStr("file_path")
	e.Severity = getStr("severity")
	e.Error = getStr("error")

	if v, ok := raw["success"].(string); ok {
		e.Success = v == "true"
	}

	return e
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
