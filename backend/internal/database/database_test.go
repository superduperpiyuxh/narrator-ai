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

func TestGetIncidents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{Title: "A", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{Title: "B", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:06:00Z", EndTime: "2025-12-21T10:10:00Z", Severity: "low", Status: "closed"}, nil)

	incs, total, err := db.GetIncidents(10, 0, "", "", "")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(incs) != 2 {
		t.Errorf("expected 2 incidents, got %d", len(incs))
	}
}

func TestGetIncidents_FilterSeverity(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{Title: "High", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{Title: "Low", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "low", Status: "new"}, nil)

	incs, total, err := db.GetIncidents(10, 0, "high", "", "")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(incs) > 0 && incs[0].Severity != "high" {
		t.Errorf("expected high severity, got %s", incs[0].Severity)
	}
}

func TestGetIncidents_FilterStatus(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{Title: "New", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{Title: "Closed", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "closed"}, nil)

	_, total, err := db.GetIncidents(10, 0, "", "closed", "")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 closed incident, got %d", total)
	}
}

func TestGetIncidents_FilterSourceIP(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{Title: "A", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{Title: "B", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)

	_, total, err := db.GetIncidents(10, 0, "", "", "10.0.0.1")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 incident from 10.0.0.1, got %d", total)
	}
}

func TestGetIncidents_Pagination(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	for i := 0; i < 5; i++ {
		db.CreateIncident(&Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}, nil)
	}

	incs, total, _ := db.GetIncidents(2, 0, "", "", "")
	if total != 5 {
		t.Errorf("expected 5 total, got %d", total)
	}
	if len(incs) != 2 {
		t.Errorf("expected 2 incidents on page 1, got %d", len(incs))
	}

	incs2, _, _ := db.GetIncidents(2, 2, "", "", "")
	if len(incs2) != 2 {
		t.Errorf("expected 2 incidents on page 2, got %d", len(incs2))
	}
}

func TestGetIncidentsByUserID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{UserID: "user-1", Title: "A", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{UserID: "user-2", Title: "B", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "low", Status: "new"}, nil)

	incs, total, err := db.GetIncidentsByUserID("user-1", 10, 0, "", "", "")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(incs) > 0 && incs[0].UserID != "user-1" {
		t.Errorf("expected user-1, got %s", incs[0].UserID)
	}
}

func TestGetIncidentsByUserID_FallbackToAll(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{UserID: "user-1", Title: "A", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)

	// Query for nonexistent user — should fall back to all
	incs, total, err := db.GetIncidentsByUserID("nonexistent", 10, 0, "", "", "")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 (fallback), got %d", total)
	}
	if len(incs) != 1 {
		t.Errorf("expected 1 incident in fallback, got %d", len(incs))
	}
}

func TestGetIncidentsByUserID_FilterSeverity(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{UserID: "user-1", Title: "High", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{UserID: "user-1", Title: "Low", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "low", Status: "new"}, nil)

	_, total, err := db.GetIncidentsByUserID("user-1", 10, 0, "low", "", "")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 low incident, got %d", total)
	}
}

func TestGetIncidentByID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	inc := &Incident{Title: "TestIncident", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new", EventCount: 42}
	db.CreateIncident(inc, nil)

	got, err := db.GetIncidentByID(inc.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected incident, got nil")
	}
	if got.Title != "TestIncident" {
		t.Errorf("expected 'TestIncident', got '%s'", got.Title)
	}
	if got.EventCount != 42 {
		t.Errorf("expected 42 events, got %d", got.EventCount)
	}
}

func TestGetIncidentByID_NotFound(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	got, err := db.GetIncidentByID(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent incident")
	}
}

func TestGetIncidentStats_NonEmpty(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{Title: "High", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 10, Severity: "high", Status: "new"}, nil)
	db.CreateIncident(&Incident{Title: "Low", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 5, Severity: "low", Status: "new"}, nil)

	stats, err := db.GetIncidentStats()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if stats["total_incidents"] != 2 {
		t.Errorf("expected 2, got %v", stats["total_incidents"])
	}
	bySev, ok := stats["by_severity"].(map[string]int)
	if !ok {
		t.Fatal("expected map[string]int for by_severity")
	}
	if bySev["high"] != 1 {
		t.Errorf("expected 1 high, got %d", bySev["high"])
	}
	if bySev["low"] != 1 {
		t.Errorf("expected 1 low, got %d", bySev["low"])
	}
}

func TestGetIncidentStatsByUserID_Empty(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	stats, err := db.GetIncidentStatsByUserID("")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if stats["total_incidents"] != 0 {
		t.Errorf("expected 0, got %v", stats["total_incidents"])
	}
}

func TestGetIncidentStatsByUserID_WithSeverity(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.CreateIncident(&Incident{UserID: "user-1", Title: "A", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 10, Severity: "critical", Status: "new"}, nil)
	db.CreateIncident(&Incident{UserID: "user-1", Title: "B", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", EventCount: 5, Severity: "critical", Status: "new"}, nil)

	stats, err := db.GetIncidentStatsByUserID("user-1")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if stats["total_incidents"] != 2 {
		t.Errorf("expected 2, got %v", stats["total_incidents"])
	}
	bySev := stats["by_severity"].(map[string]int)
	if bySev["critical"] != 2 {
		t.Errorf("expected 2 critical, got %d", bySev["critical"])
	}
}

func TestCreateIncident_WithTechniques(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.SeedTechniques([]TechniqueRef{
		{TechniqueID: "T1110", Name: "Brute Force", Tactic: "credential-access"},
		{TechniqueID: "T1059", Name: "Command and Scripting Interpreter", Tactic: "execution"},
	})

	inc := &Incident{
		Title:      "Attack",
		SourceIP:   "10.0.0.1",
		StartTime:  "2025-12-21T10:00:00Z",
		EndTime:    "2025-12-21T10:05:00Z",
		Severity:   "critical",
		Status:     "new",
		Techniques: []TechniqueRef{{TechniqueID: "T1110", EventCount: 5}, {TechniqueID: "T1059", EventCount: 3}},
	}
	err = db.CreateIncident(inc, nil)
	if err != nil {
		t.Fatalf("create incident failed: %v", err)
	}
	if inc.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateIncident_WithEventIDs(t *testing.T) {
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

	inc := &Incident{Title: "Linked", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new"}
	err = db.CreateIncident(inc, []int64{1, 2})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

linkedEvents, err := db.GetIncidentEvents(inc.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(linkedEvents) != 2 {
		t.Errorf("expected 2 linked events, got %d", len(linkedEvents))
	}
}

func TestCreateNarrative(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	narr := &Narrative{IncidentID: inc.ID, UserID: "user-1", Summary: "Test narrative", Confidence: 0.9, Sentences: "[]", ModelUsed: "gpt-4o", Temperature: 0.3, TokensUsed: 150, GenerationTimeMs: 2500}
	err = db.CreateNarrative(narr)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if narr.ID == 0 {
		t.Error("expected non-zero ID")
	}

	got, _ := db.GetNarrativeByID(narr.ID)
	if got == nil {
		t.Fatal("expected narrative, got nil")
	}
	if got.Summary != "Test narrative" {
		t.Errorf("expected 'Test narrative', got '%s'", got.Summary)
	}
}

func TestCreateFeedback(t *testing.T) {
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

	fb := &Feedback{NarrativeID: narr.ID, IncidentID: inc.ID, Rating: 1, Notes: "Excellent"}
	err = db.CreateFeedback(fb)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if fb.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestGetTechniqueCounts(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.SeedTechniques([]TechniqueRef{
		{TechniqueID: "T1110", Name: "Brute Force", Tactic: "credential-access"},
		{TechniqueID: "T1059", Name: "Command and Scripting Interpreter", Tactic: "execution"},
	})

	inc1 := &Incident{Title: "A", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new", Techniques: []TechniqueRef{{TechniqueID: "T1110", EventCount: 5}}}
	inc2 := &Incident{Title: "B", SourceIP: "10.0.0.2", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "high", Status: "new", Techniques: []TechniqueRef{{TechniqueID: "T1110", EventCount: 3}, {TechniqueID: "T1059", EventCount: 2}}}
	db.CreateIncident(inc1, nil)
	db.CreateIncident(inc2, nil)

	counts, err := db.GetTechniqueCounts()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if counts["T1110"] != 2 {
		t.Errorf("expected 2 for T1110, got %d", counts["T1110"])
	}
	if counts["T1059"] != 1 {
		t.Errorf("expected 1 for T1059, got %d", counts["T1059"])
	}
}

func TestGetTechniqueCounts_Empty(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	counts, err := db.GetTechniqueCounts()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map, got %d entries", len(counts))
	}
}

func TestGetNarrativeByIncidentID(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	narr := &Narrative{IncidentID: inc.ID, Summary: "Found", Confidence: 0.8, Sentences: "[]", ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	got, err := db.GetNarrativeByIncidentID(inc.ID)
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

func TestGetNarrativeSourceEvents_WithEventIDs(t *testing.T) {
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

	inc := &Incident{Title: "Test", SourceIP: "10.0.0.1", StartTime: "2025-12-21T10:00:00Z", EndTime: "2025-12-21T10:05:00Z", Severity: "medium", Status: "new"}
	db.CreateIncident(inc, nil)

	sentences := `{"sentences":[{"text":"First","timestamp":"2025-12-21T10:00:00Z","source_event_ids":[1,2],"confidence":0.9},{"text":"Second","timestamp":"2025-12-21T10:01:00Z","source_event_ids":[2,3],"confidence":0.85}]}`
	narr := &Narrative{IncidentID: inc.ID, Summary: "Test", Confidence: 0.8, Sentences: sentences, ModelUsed: "test", Temperature: 0.2}
	db.CreateNarrative(narr)

	got, err := db.GetNarrativeSourceEvents(narr.ID)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 events, got %d", len(got))
	}
}

func TestGetEvents_DefaultParams(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1"},
	}
	db.InsertEvents(events)

	got, total, err := db.GetEvents(10, 0)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(got) != 1 {
		t.Errorf("expected 1, got %d", len(got))
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
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1", UserName: "admin"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "web01", EventType: "network", SourceIP: "10.0.0.2", UserName: "guest"},
	}
	db.InsertEvents(events)

	got, err := db.SearchEvents("admin")
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
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "host1", EventType: "auth", SourceIP: "10.0.0.1"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "host2", EventType: "network", SourceIP: "10.0.0.2"},
	}
	db.InsertEvents(events)

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if stats["total_events"] != 2 {
		t.Errorf("expected 2, got %v", stats["total_events"])
	}
	if stats["unique_hosts"] != 2 {
		t.Errorf("expected 2 hosts, got %v", stats["unique_hosts"])
	}
}


