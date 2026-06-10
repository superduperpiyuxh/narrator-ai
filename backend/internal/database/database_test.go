package database

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	if err := db.HealthCheck(); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestInsertAndGetEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{
			Timestamp: "2025-12-21T10:00:00Z",
			Hostname:  "test-host",
			EventType: "authentication",
			SourceIP:  "10.0.0.1",
			UserName:  "admin",
		},
	}

	err = db.InsertEvents(events)
	if err != nil {
		t.Fatalf("failed to insert events: %v", err)
	}

	got, total, err := db.GetEvents(10, 0)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 event, got %d", total)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 event in slice, got %d", len(got))
	}
	if got[0].Hostname != "test-host" {
		t.Errorf("expected hostname test-host, got %s", got[0].Hostname)
	}
}

func TestSearchEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "authentication", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "web01", EventType: "network_activity", SourceIP: "10.0.0.2"},
	}
	db.InsertEvents(events)

	got, err := db.SearchEvents("dc01")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result, got %d", len(got))
	}
}

func TestGetStats(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "authentication", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "host2", EventType: "network_activity", SourceIP: "10.0.0.2"},
	}
	db.InsertEvents(events)

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats["total_events"] != 2 {
		t.Errorf("expected 2 events, got %v", stats["total_events"])
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
