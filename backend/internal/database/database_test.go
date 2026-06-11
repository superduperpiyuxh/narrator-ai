package database

import (
	"testing"
)

func TestGetEventsByUserID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{UserID: "user-1", Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1"},
		{UserID: "user-1", Timestamp: "2025-12-21T10:01:00Z", Hostname: "host2", EventType: "network", SourceIP: "10.0.0.2"},
		{UserID: "user-2", Timestamp: "2025-12-21T10:02:00Z", Hostname: "host3", EventType: "auth", SourceIP: "10.0.0.3"},
	}
	db.InsertEvents(events)

	got, total, err := db.GetEventsByUserID("user-1", 10, 0)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 events for user-1, got %d", total)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 events in slice, got %d", len(got))
	}
}

func TestGetEventsByUserID_Empty(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	got, total, err := db.GetEventsByUserID("nonexistent", 10, 0)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0, got %d", total)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestSearchEventsByUserID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{UserID: "user-1", Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", UserName: "admin", SourceIP: "10.0.0.1"},
		{UserID: "user-1", Timestamp: "2025-12-21T10:01:00Z", Hostname: "web01", EventType: "network", UserName: "guest", SourceIP: "10.0.0.2"},
		{UserID: "user-2", Timestamp: "2025-12-21T10:02:00Z", Hostname: "dc01", EventType: "auth", UserName: "admin", SourceIP: "10.0.0.3"},
	}
	db.InsertEvents(events)

	got, err := db.SearchEventsByUserID("user-1", "dc01")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result, got %d", len(got))
	}
}

func TestGetStatsByUserID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{UserID: "user-1", Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1"},
		{UserID: "user-1", Timestamp: "2025-12-21T10:01:00Z", Hostname: "host2", EventType: "network", SourceIP: "10.0.0.2"},
	}
	db.InsertEvents(events)

	db.CreateIncident(&Incident{UserID: "user-1", Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)

	stats, err := db.GetStatsByUserID("user-1")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if stats["total_events"] != 2 {
		t.Errorf("expected 2 events, got %v", stats["total_events"])
	}
	if stats["unique_hosts"] != 2 {
		t.Errorf("expected 2 hosts, got %v", stats["unique_hosts"])
	}
	if stats["total_incidents"] != 1 {
		t.Errorf("expected 1 incident, got %v", stats["total_incidents"])
	}
}

func TestGetEventsByHost(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "web01", EventType: "network", SourceIP: "10.0.0.2"},
	}
	db.InsertEvents(events)

	got, err := db.GetEventsByHost("dc01")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 event, got %d", len(got))
	}
}

func TestGetEventsByType(t *testing.T) {
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

	got, err := db.GetEventsByType("authentication")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 event, got %d", len(got))
	}
}

func TestInsertEvents_DuplicateIgnored(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1", EventID: "evt-1"},
	}
	err = db.InsertEvents(events)
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	err = db.InsertEvents(events)
	if err != nil {
		t.Fatalf("second insert failed: %v", err)
	}

	_, total, _ := db.GetEvents(100, 0)
	if total != 1 {
		t.Errorf("expected 1 event (duplicate ignored), got %d", total)
	}
}

func TestGetStats_Empty(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if stats["total_events"] != 0 {
		t.Errorf("expected 0 events, got %v", stats["total_events"])
	}
}

func TestGetIncidentStatsByUserID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{UserID: "user-1", Title: "High", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 10, Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{UserID: "user-1", Title: "Low", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 5, Severity: "low", Status: "new"}, nil)
	db.CreateIncident(&Incident{UserID: "user-2", Title: "Other", SourceIP: "10.0.0.3", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}, nil)

	stats, err := db.GetIncidentStatsByUserID("user-1")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if stats["total_incidents"] != 2 {
		t.Errorf("expected 2 incidents for user-1, got %v", stats["total_incidents"])
	}
}

func TestGetIncidentEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1", UserName: "admin"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "dc01", EventType: "process", SourceIP: "10.0.0.1", UserName: "admin"},
	}
	db.InsertEvents(events)

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, []int64{1, 2})

	got, err := db.GetIncidentEvents(inc.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 events, got %d", len(got))
	}
}

func TestGetUnprocessedEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "dc01", EventType: "process", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, []int64{1})

	unprocessed, err := db.GetUnprocessedEvents()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(unprocessed) != 1 {
		t.Errorf("expected 1 unprocessed event, got %d", len(unprocessed))
	}
}

func TestSeedTechniques(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	techniques := []TechniqueRef{
		{TechniqueID: "T1110", Name: "Brute Force", Tactic: "credential-access"},
		{TechniqueID: "T1078", Name: "Valid Accounts", Tactic: "initial-access"},
	}
	err = db.SeedTechniques(techniques)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	// Seed again — INSERT OR IGNORE should not error
	err = db.SeedTechniques(techniques)
	if err != nil {
		t.Fatalf("re-seed failed: %v", err)
	}
}

func TestGetNarrativeByID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	narr := &Narrative{IncidentID: inc.ID, Summary: "Test narrative", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	got, err := db.GetNarrativeByID(narr.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected narrative, got nil")
	}
	if got.ID != narr.ID {
		t.Errorf("expected ID %d, got %d", narr.ID, got.ID)
	}
}

func TestGetNarrativeByID_NotFound(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	got, err := db.GetNarrativeByID(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for non-existent narrative")
	}
}

func TestGetNarrativeByIncidentID_NotFound(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	got, err := db.GetNarrativeByIncidentID(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for non-existent narrative")
	}
}

func TestGetNarrativeSourceEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "dc01", EventType: "process", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:02:00Z", Hostname: "web01", EventType: "network", SourceIP: "10.0.0.2"},
	}
	db.InsertEvents(events)

	sentences := `{"sentences":[{"text":"First sentence","timestamp":"2025-12-21T10:00:00Z","source_event_ids":[1,2],"confidence":0.9},{"text":"Second sentence","timestamp":"2025-12-21T10:01:00Z","source_event_ids":[2,3],"confidence":0.85}]}`

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	narr := &Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: sentences, ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	got, err := db.GetNarrativeSourceEvents(narr.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	// Event IDs 1, 2, 3 referenced — should get 3 events (deduped)
	if len(got) != 3 {
		t.Errorf("expected 3 source events, got %d", len(got))
	}
}

func TestGetNarrativeSourceEvents_NoEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	narr := &Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: `{"sentences":[{"text":"No refs","timestamp":"2025-12-21T10:00:00Z","source_event_ids":[],"confidence":0.9}]}`, ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	got, err := db.GetNarrativeSourceEvents(narr.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 events, got %d", len(got))
	}
}

func TestGetFeedbackByIncidentID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	narr := &Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	fb1 := &Feedback{NarrativeID: narr.ID, IncidentID: inc.ID, Rating: 1, Notes: "Good"}
	fb2 := &Feedback{NarrativeID: narr.ID, IncidentID: inc.ID, Rating: -1, Notes: "Bad"}
	db.CreateFeedback(fb1)
	db.CreateFeedback(fb2)

	feedbacks, err := db.GetFeedbackByIncidentID(inc.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(feedbacks) != 2 {
		t.Errorf("expected 2 feedbacks, got %d", len(feedbacks))
	}
}

func TestGetFeedbackByIncidentID_Empty(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	feedbacks, err := db.GetFeedbackByIncidentID(999)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(feedbacks) != 0 {
		t.Errorf("expected 0 feedbacks, got %d", len(feedbacks))
	}
}


