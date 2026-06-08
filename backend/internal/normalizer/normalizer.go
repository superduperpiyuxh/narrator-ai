package normalizer

import (
	"regexp"
	"strings"
	"time"

	"github.com/superduperpiyuxh/narrator-ai/backend/internal/graylog"
)

var eventTypeMap = map[string]string{
	"process_start":            "process_activity",
	"process_termination":      "process_activity",
	"network_connection":       "network_activity",
	"network_connection_failed": "network_activity",
	"login_attempt":            "authentication",
	"login_success":            "authentication",
	"login_failure":            "authentication",
	"admin_action":             "privilege_escalation",
	"file_access":              "file_activity",
	"web_browsing":             "web_activity",
	"kerberos_auth_success":    "authentication",
	"api_key_auth_success":     "authentication",
	"api_key_auth_failure":     "authentication",
	"certificate_auth_success": "authentication",
	"oauth_token_success":      "authentication",
	"system_event":             "system",
}

var ipRegex = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)

func NormalizeEvent(raw graylog.Event) graylog.Event {
	e := raw

	e.Timestamp = normalizeTimestamp(e.Timestamp)
	e.EventType = normalizeEventType(e.EventType)
	e.Success = normalizeBool(e.Success)
	e.SourceIP = normalizeIP(e.SourceIP)
	e.DestIP = normalizeIP(e.DestIP)
	e.User = normalizeString(e.User)
	e.Hostname = normalizeString(e.Hostname)
	e.ProcessName = normalizeString(e.ProcessName)
	e.CommandLine = normalizeString(e.CommandLine)
	e.ParentProcess = normalizeString(e.ParentProcess)
	e.EventID = normalizeString(e.EventID)
	e.LogType = normalizeString(e.LogType)
	e.SessionID = normalizeString(e.SessionID)
	e.Department = normalizeString(e.Department)
	e.Location = normalizeString(e.Location)
	e.DeviceType = normalizeString(e.DeviceType)
	e.Port = normalizeString(e.Port)
	e.Protocol = normalizeString(e.Protocol)
	e.FilePath = normalizeString(e.FilePath)
	e.Severity = normalizeString(e.Severity)
	e.Error = normalizeString(e.Error)

	return e
}

func normalizeTimestamp(ts string) string {
	if ts == "" {
		return time.Now().UTC().Format(time.RFC3339)
	}

	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}

	return ts
}

func normalizeEventType(raw string) string {
	if mapped, ok := eventTypeMap[raw]; ok {
		return mapped
	}
	return raw
}

func normalizeBool(val bool) bool {
	return val
}

func normalizeIP(ip string) string {
	if ip == "" {
		return ""
	}
	if ipRegex.MatchString(ip) {
		return ip
	}
	return ""
}

func normalizeString(s string) string {
	if s == "" || s == "null" || s == "undefined" {
		return ""
	}
	return strings.TrimSpace(s)
}
