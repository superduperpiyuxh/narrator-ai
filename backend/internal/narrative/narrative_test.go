package narrative

import (
	"testing"

	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func TestCalculateConfidence_FullCoverage(t *testing.T) {
	incident := database.Incident{
		Techniques: []database.TechniqueRef{
			{TechniqueID: "T1110", Name: "Brute Force"},
			{TechniqueID: "T1078", Name: "Valid Accounts"},
		},
	}
	events := []database.Event{
		{ID: 1}, {ID: 2}, {ID: 3},
	}
	narrative := &Narrative{
		Sentences: []Sentence{
			{SourceEventIDs: []int64{1, 2, 3}, Confidence: 1.0, Technique: "T1110"},
			{SourceEventIDs: []int64{1}, Confidence: 1.0, Technique: "T1078"},
		},
	}

	score := CalculateConfidence(incident, events, narrative)
	if score < 0.9 {
		t.Errorf("expected confidence >= 0.9 for full coverage, got %f", score)
	}
}

func TestCalculateConfidence_NoSentences(t *testing.T) {
	incident := database.Incident{}
	events := []database.Event{{ID: 1}}
	narrative := &Narrative{Sentences: []Sentence{}}

	score := CalculateConfidence(incident, events, narrative)
	if score != 0 {
		t.Errorf("expected 0 for no sentences, got %f", score)
	}
}

func TestCalculateConfidence_NoTechniques(t *testing.T) {
	incident := database.Incident{Techniques: nil}
	events := []database.Event{{ID: 1}}
	narrative := &Narrative{
		Sentences: []Sentence{
			{SourceEventIDs: []int64{1}, Confidence: 0.8},
		},
	}

	score := CalculateConfidence(incident, events, narrative)
	if score < 0 || score > 1 {
		t.Errorf("expected 0-1, got %f", score)
	}
}

func TestCalculateConfidence_LowCoverage(t *testing.T) {
	incident := database.Incident{
		Techniques: []database.TechniqueRef{
			{TechniqueID: "T1110"},
			{TechniqueID: "T1078"},
			{TechniqueID: "T1059"},
			{TechniqueID: "T1021"},
		},
	}
	events := []database.Event{
		{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5},
	}
	narrative := &Narrative{
		Sentences: []Sentence{
			{SourceEventIDs: []int64{1}, Confidence: 0.5},
		},
	}

	score := CalculateConfidence(incident, events, narrative)
	if score > 0.5 {
		t.Errorf("expected low confidence for low coverage, got %f", score)
	}
}

func TestParseResponse_ValidJSON(t *testing.T) {
	g := &Generator{}
	response := `{"summary":"Test summary","sentences":[{"text":"Sentence 1","timestamp":"2025-12-21T10:00:00Z","source_event_ids":[1,2],"confidence":0.9,"technique":"T1110"}]}`
	events := []database.Event{{ID: 1}, {ID: 2}}

	narr, err := g.parseResponse(response, events)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if narr.Summary != "Test summary" {
		t.Errorf("expected 'Test summary', got '%s'", narr.Summary)
	}
	if len(narr.Sentences) != 1 {
		t.Errorf("expected 1 sentence, got %d", len(narr.Sentences))
	}
	if len(narr.Sentences[0].SourceEventIDs) != 2 {
		t.Errorf("expected 2 source event IDs, got %d", len(narr.Sentences[0].SourceEventIDs))
	}
}

func TestParseResponse_WithMarkdown(t *testing.T) {
	g := &Generator{}
	response := "```json\n{\"summary\":\"Test\",\"sentences\":[]}\n```"
	events := []database.Event{}

	narr, err := g.parseResponse(response, events)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if narr.Summary != "Test" {
		t.Errorf("expected 'Test', got '%s'", narr.Summary)
	}
}

func TestParseResponse_InvalidJSON(t *testing.T) {
	g := &Generator{}
	events := []database.Event{}

	_, err := g.parseResponse("not json at all", events)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseResponse_FiltersInvalidEventIDs(t *testing.T) {
	g := &Generator{}
	response := `{"summary":"Test","sentences":[{"text":"Sentence","timestamp":"2025-12-21T10:00:00Z","source_event_ids":[1,999],"confidence":0.9}]}`
	events := []database.Event{{ID: 1}} // Only ID 1 exists

	narr, err := g.parseResponse(response, events)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(narr.Sentences[0].SourceEventIDs) != 1 {
		t.Errorf("expected 1 valid event ID (999 filtered), got %d", len(narr.Sentences[0].SourceEventIDs))
	}
}

func TestFormatEvents(t *testing.T) {
	g := &Generator{}
	events := []database.Event{
		{ID: 1, Timestamp: "2025-12-21T10:00:00Z", EventType: "auth", SourceIP: "10.0.0.1", UserName: "admin", ProcessName: "lsass.exe", CommandLine: "lsass"},
		{ID: 2, Timestamp: "2025-12-21T10:01:00Z", EventType: "network", SourceIP: "10.0.0.1", UserName: "admin", ProcessName: "chrome", CommandLine: "chrome"},
	}

	result := g.formatEvents(events)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	if len(result) < 50 {
		t.Errorf("expected substantial output, got %d chars", len(result))
	}
}

func TestBuildPrompt(t *testing.T) {
	g := &Generator{}
	incident := database.Incident{
		Title:       "Brute Force Attack",
		SourceIP:    "10.0.0.1",
		StartTime:   "2025-12-21T10:00:00Z",
		EndTime:     "2025-12-21T10:05:00Z",
		EventCount:  10,
		Severity:    "high",
		MitreAttackIDs: []string{"T1110"},
	}
	events := []database.Event{
		{ID: 1, Timestamp: "2025-12-21T10:00:00Z", EventType: "auth", SourceIP: "10.0.0.1"},
	}

	prompt := g.buildPrompt(incident, events)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if len(prompt) < 100 {
		t.Errorf("expected substantial prompt, got %d chars", len(prompt))
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	g := &Generator{}
	prompt := g.buildSystemPrompt()
	if prompt == "" {
		t.Fatal("expected non-empty system prompt")
	}
}

func TestNewGenerator(t *testing.T) {
	g := NewGenerator(nil)
	if g == nil {
		t.Fatal("expected non-nil generator")
	}
}
