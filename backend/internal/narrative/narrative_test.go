package narrative

import (
	"testing"

	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func TestCalculateConfidence(t *testing.T) {
	incident := database.Incident{
		EventCount: 10,
		Techniques: []database.TechniqueRef{
			{TechniqueID: "T1110", Name: "Brute Force"},
			{TechniqueID: "T1021", Name: "Remote Services"},
		},
	}

	events := make([]database.Event, 10)
	for i := range events {
		events[i] = database.Event{ID: int64(i + 1)}
	}

	narrative := &Narrative{
		Sentences: []Sentence{
			{
				Text:          "Attacker performed brute force",
				SourceEventIDs: []int64{1, 2, 3},
				Confidence:    0.9,
				Technique:     "T1110",
			},
			{
				Text:          "Then moved laterally",
				SourceEventIDs: []int64{4, 5},
				Confidence:    0.8,
				Technique:     "T1021",
			},
		},
	}

	conf := CalculateConfidence(incident, events, narrative)
	if conf < 0 || conf > 1 {
		t.Errorf("confidence out of range: %f", conf)
	}
	if conf < 0.3 {
		t.Errorf("confidence too low: %f", conf)
	}
}

func TestCalculateConfidenceEmpty(t *testing.T) {
	incident := database.Incident{}
	events := []database.Event{}
	narrative := &Narrative{Sentences: []Sentence{}}

	conf := CalculateConfidence(incident, events, narrative)
	if conf != 0 {
		t.Errorf("expected 0 for empty narrative, got %f", conf)
	}
}
