package incident

import (
	"testing"
	"time"

	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func TestGroupEvents(t *testing.T) {
	events := make([]database.Event, 10)
	baseTime := time.Date(2025, 12, 21, 10, 0, 0, 0, time.UTC)
	for i := range events {
		events[i] = database.Event{
			SourceIP:  "10.0.0.1",
			Timestamp: baseTime.Add(time.Duration(i) * 2 * time.Minute).Format(time.RFC3339),
		}
	}

	groups := GroupEventsIntoIncidents(events, 15)
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0]) != 10 {
		t.Errorf("expected 10 events in group, got %d", len(groups[0]))
	}
}

func TestGroupEventsTimeGap(t *testing.T) {
	events := []database.Event{
		{SourceIP: "10.0.0.1", Timestamp: "2025-12-21T10:00:00Z"},
		{SourceIP: "10.0.0.1", Timestamp: "2025-12-21T10:05:00Z"},
		{SourceIP: "10.0.0.1", Timestamp: "2025-12-21T10:30:00Z"},
	}

	groups := GroupEventsIntoIncidents(events, 15)
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestGroupEventsDifferentIPs(t *testing.T) {
	events := []database.Event{
		{SourceIP: "10.0.0.1", Timestamp: "2025-12-21T10:00:00Z"},
		{SourceIP: "10.0.0.2", Timestamp: "2025-12-21T10:01:00Z"},
		{SourceIP: "10.0.0.3", Timestamp: "2025-12-21T10:02:00Z"},
	}

	groups := GroupEventsIntoIncidents(events, 15)
	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}
}

func TestGroupEventsEmpty(t *testing.T) {
	groups := GroupEventsIntoIncidents(nil, 15)
	if groups != nil {
		t.Errorf("expected nil, got %v", groups)
	}
}

func TestBuildIncidentFromGroup(t *testing.T) {
	events := []database.Event{
		{
			SourceIP:    "10.0.0.1",
			Timestamp:   "2025-12-21T10:00:00Z",
			UserName:    "admin",
			Hostname:    "DC01",
			EventID:     "4625",
			EventType:   "authentication",
			ProcessName: "lsass.exe",
		},
		{
			SourceIP:    "10.0.0.1",
			Timestamp:   "2025-12-21T10:05:00Z",
			UserName:    "admin",
			Hostname:    "DC01",
			EventID:     "4625",
			EventType:   "authentication",
			ProcessName: "lsass.exe",
		},
	}

	inc := BuildIncidentFromGroup(events)

	if inc.SourceIP != "10.0.0.1" {
		t.Errorf("expected source_ip 10.0.0.1, got %s", inc.SourceIP)
	}
	if inc.EventCount != 2 {
		t.Errorf("expected 2 events, got %d", inc.EventCount)
	}
	if len(inc.UniqueUsers) != 1 {
		t.Errorf("expected 1 unique user, got %d", len(inc.UniqueUsers))
	}
	if len(inc.Techniques) == 0 {
		t.Error("expected techniques to be mapped")
	}
}

func TestCalculateSeverity(t *testing.T) {
	tests := []struct {
		name       string
		count      int
		eventTypes []string
		techniques []string
		want       string
	}{
		{"single low event", 1, []string{"system"}, nil, "low"},
		{"brute force", 100, []string{"authentication"}, []string{"T1110"}, "critical"},
		{"lateral movement", 50, []string{"network_activity"}, []string{"T1021"}, "critical"},
		{"mixed activity", 20, []string{"file_activity", "process_activity"}, []string{"T1059"}, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSeverity(tt.count, tt.eventTypes, tt.techniques)
			if got != tt.want {
				t.Errorf("CalculateSeverity() = %s, want %s", got, tt.want)
			}
		})
	}
}
