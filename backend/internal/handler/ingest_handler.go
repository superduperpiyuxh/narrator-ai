package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/normalizer"
)

type IngestHandler struct {
	db      *database.DB
	dataDir string
}

func NewIngestHandler(db *database.DB, dataDir string) *IngestHandler {
	return &IngestHandler{db: db, dataDir: dataDir}
}

type IngestEvent struct {
	Timestamp     string                 `json:"timestamp"`
	Hostname      string                 `json:"hostname"`
	EventType     string                 `json:"event_type"`
	EventID       string                 `json:"event_id"`
	UserName      string                 `json:"user_name"`
	SourceIP      string                 `json:"source_ip"`
	DestIP        string                 `json:"dest_ip"`
	ProcessName   string                 `json:"process_name"`
	CommandLine   string                 `json:"command_line"`
	ParentProcess string                 `json:"parent_process"`
	LogType       string                 `json:"log_type"`
	SessionID     string                 `json:"session_id"`
	Department    string                 `json:"department"`
	Location      string                 `json:"location"`
	DeviceType    string                 `json:"device_type"`
	Success       bool                   `json:"success"`
	Port          string                 `json:"port"`
	Protocol      string                 `json:"protocol"`
	FilePath      string                 `json:"file_path"`
	Severity      string                 `json:"severity"`
	Error         string                 `json:"error"`
	RawJSON       map[string]interface{} `json:"raw_json"`
}

type IngestRequest struct {
	Events []IngestEvent `json:"events" binding:"required"`
}

type IngestFileRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}

func (h *IngestHandler) IngestEvents(c *gin.Context) {
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

	if len(req.Events) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max 10000 events per request"})
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
				"timestamp":      ie.Timestamp,
				"hostname":       ie.Hostname,
				"event_type":     ie.EventType,
				"event_id":       ie.EventID,
				"user":           ie.UserName,
				"source_ip":      ie.SourceIP,
				"destination_ip": ie.DestIP,
				"process_name":   ie.ProcessName,
				"command_line":   ie.CommandLine,
				"parent_process": ie.ParentProcess,
				"log_type":       ie.LogType,
				"session_id":     ie.SessionID,
				"department":     ie.Department,
				"location":       ie.Location,
				"device_type":    ie.DeviceType,
				"success":        fmt.Sprintf("%t", ie.Success),
				"port":           ie.Port,
				"protocol":       ie.Protocol,
				"file_path":      ie.FilePath,
				"severity":       ie.Severity,
				"error":          ie.Error,
			}
		}

		e := normalizer.Event{
			Timestamp:     ie.Timestamp,
			Hostname:      ie.Hostname,
			EventType:     ie.EventType,
			EventID:       ie.EventID,
			User:          ie.UserName,
			SourceIP:      ie.SourceIP,
			DestIP:        ie.DestIP,
			ProcessName:   ie.ProcessName,
			CommandLine:   ie.CommandLine,
			ParentProcess: ie.ParentProcess,
			LogType:       ie.LogType,
			SessionID:     ie.SessionID,
			Department:    ie.Department,
			Location:      ie.Location,
			DeviceType:    ie.DeviceType,
			Success:       ie.Success,
			Port:          ie.Port,
			Protocol:      ie.Protocol,
			FilePath:      ie.FilePath,
			Severity:      ie.Severity,
			Error:         ie.Error,
			RawJSON:       rawJSON,
		}
		n := normalizer.NormalizeEvent(e)

		batch = append(batch, database.Event{
			UserID:        userIDStr,
			Timestamp:     n.Timestamp,
			Hostname:      n.Hostname,
			EventType:     n.EventType,
			EventID:       n.EventID,
			UserName:      n.User,
			SourceIP:      n.SourceIP,
			DestIP:        n.DestIP,
			ProcessName:   n.ProcessName,
			CommandLine:   n.CommandLine,
			ParentProcess: n.ParentProcess,
			LogType:       n.LogType,
			SessionID:     n.SessionID,
			Department:    n.Department,
			Location:      n.Location,
			DeviceType:    n.DeviceType,
			Success:       n.Success,
			Port:          n.Port,
			Protocol:      n.Protocol,
			FilePath:      n.FilePath,
			Severity:      n.Severity,
			Error:         n.Error,
			RawJSON:       n.RawJSON,
		})
	}

	if err := h.db.InsertEvents(batch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": len(batch),
		"message":  fmt.Sprintf("successfully imported %d events", len(batch)),
	})
}

func (h *IngestHandler) IngestFile(c *gin.Context) {
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

	var req IngestFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_path required"})
		return
	}

	// Sanitize path
	absPath := filepath.Clean(req.FilePath)
	// Ensure file is within allowed directory
	allowedDir := filepath.Clean(h.dataDir)
	if !strings.HasPrefix(absPath, allowedDir) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_path must be within data directory"})
		return
	}

	file, err := os.Open(absPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("cannot open file: %v", err)})
		return
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

		e := normalizer.Event{RawJSON: raw}
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

		n := normalizer.NormalizeEvent(e)
		batch = append(batch, database.Event{
			UserID:        userIDStr,
			Timestamp:     n.Timestamp,
			Hostname:      n.Hostname,
			EventType:     n.EventType,
			EventID:       n.EventID,
			UserName:      n.User,
			SourceIP:      n.SourceIP,
			DestIP:        n.DestIP,
			ProcessName:   n.ProcessName,
			CommandLine:   n.CommandLine,
			ParentProcess: n.ParentProcess,
			LogType:       n.LogType,
			SessionID:     n.SessionID,
			Department:    n.Department,
			Location:      n.Location,
			DeviceType:    n.DeviceType,
			Success:       n.Success,
			Port:          n.Port,
			Protocol:      n.Protocol,
			FilePath:      n.FilePath,
			Severity:      n.Severity,
			Error:         n.Error,
			RawJSON:       n.RawJSON,
		})
		total++

		if len(batch) >= 5000 {
			if err := h.db.InsertEvents(batch); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			batch = batch[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("read error: %v", err)})
		return
	}

	if len(batch) > 0 {
		if err := h.db.InsertEvents(batch); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": total,
		"message":  fmt.Sprintf("successfully imported %d events from file", total),
	})
}
