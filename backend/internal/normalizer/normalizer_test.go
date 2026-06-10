package normalizer

import (
	"testing"
)

func TestNormalizeTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"rfc3339", "2025-12-21T10:00:00Z", "2025-12-21T10:00:00Z"},
		{"space format", "2025-12-21 10:00:00", "2025-12-21T10:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("normalizeTimestamp(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeEventType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"process_start", "process_activity"},
		{"login_success", "authentication"},
		{"network_connection", "network_activity"},
		{"unknown_type", "unknown_type"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeEventType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeEventType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"10.0.0.1", "10.0.0.1"},
		{"192.168.1.1", "192.168.1.1"},
		{"invalid", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeIP(tt.input)
			if got != tt.want {
				t.Errorf("normalizeIP(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"", ""},
		{"null", ""},
		{"undefined", ""},
		{"  spaces  ", "spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeString(tt.input)
			if got != tt.want {
				t.Errorf("normalizeString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeEvent(t *testing.T) {
	e := Event{
		Timestamp: "2025-12-21 10:00:00",
		EventType: "process_start",
		SourceIP:  "10.0.0.1",
		User:      "admin",
	}

	got := NormalizeEvent(e)
	if got.Timestamp != "2025-12-21T10:00:00Z" {
		t.Errorf("expected normalized timestamp, got %s", got.Timestamp)
	}
	if got.EventType != "process_activity" {
		t.Errorf("expected process_activity, got %s", got.EventType)
	}
}
