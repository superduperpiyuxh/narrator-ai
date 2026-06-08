# Architecture Patterns

**Domain:** Security Incident Narrative Generator
**Research Date:** 2026-06-08
**Research Mode:** Ecosystem

## Executive Summary

This document recommends the architecture for NarratorAI, a security incident narrative generator. The architecture follows an event-driven pipeline pattern with clear separation between data ingestion, processing, AI integration, and presentation layers. Key design decisions prioritize security (prompt injection prevention), traceability (source linking), and hackathon constraints (3-week timeline).

## Recommended Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Next.js)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Story Cards  │  │  Incident    │  │   Raw Event  │         │
│  │  Dashboard   │  │    List      │  │    Viewer    │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      REST API (Gin)                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   /api/v1    │  │  /api/v1     │  │  /api/v1     │         │
│  │  incidents   │  │  narratives  │  │   feedback   │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Processing Pipeline                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Ingestion  │  │  Aggregation │  │  Narrative   │         │
│  │   Service    │  │   Service    │  │   Service    │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    External Services                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Graylog    │  │  Claude API  │  │  MITRE ATT&CK│         │
│  │   REST API   │  │  (Anthropic) │  │    STIX      │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

## Component Boundaries

### 1. Data Ingestion Layer

**Responsibility:** Fetch alerts from Graylog REST API

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| GraylogClient | HTTP client for Graylog API | Graylog Server |
| EventParser | Parse raw alerts into structured format | EventRepository |
| EventRepository | Store raw events in SQLite | SQLite Database |

**Key Patterns:**
- Use standard `net/http` client with proper authentication headers
- Implement retry logic with exponential backoff
- Parse Graylog's JSON response format into normalized event structure
- Store raw events with timestamps for source tracing

### 2. Event Aggregation Layer

**Responsibility:** Group related events into incidents

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| EventAggregator | Group events by IP/time window | EventRepository |
| IncidentClustering | Create incident boundaries | IncidentRepository |
| MitreMapper | Map events to ATT&CK techniques | MitreRepository |

**Key Patterns:**
- Time window clustering (e.g., events within 5 minutes of each other)
- IP-based grouping (same source IP indicates related activity)
- STIX data parsing for ATT&CK technique mapping
- Confidence scoring based on event coverage

### 3. Narrative Generation Layer

**Responsibility:** Generate chronological attack stories

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| NarrativeService | Orchestrate LLM calls | Claude API |
| PromptBuilder | Construct secure prompts | PromptTemplates |
| SourceTracer | Link sentences to events | EventRepository |
| InjectionPreventer | Sanitize inputs | Claude API |

**Key Patterns:**
- XML wrapping for untrusted content
- Input validation with regex patterns
- Output validation with JSON schema
- Haiku classifier for secondary detection

### 4. Presentation Layer

**Responsibility:** Display narratives to analysts

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| StoryCard | Display incident narrative | REST API |
| IncidentList | Browse all incidents | REST API |
| RawEventViewer | Show source events | REST API |
| FeedbackForm | Capture analyst annotations | REST API |

**Key Patterns:**
- Server-side rendering for initial load
- Client-side updates for real-time feedback
- Responsive design for various screen sizes
- Accessible UI components

## Data Flow

### 1. Ingestion Flow

```
Graylog REST API
       │
       ▼
  GraylogClient.FetchAlerts()
       │
       ▼
  EventParser.ParseResponse()
       │
       ▼
  EventRepository.SaveEvents()
       │
       ▼
  SQLite Database
```

### 2. Processing Flow

```
EventRepository.GetUnprocessed()
       │
       ▼
  EventAggregator.GroupByIP()
       │
       ▼
  IncidentClustering.CreateIncidents()
       │
       ▼
  MitreMapper.MapToTechniques()
       │
       ▼
  IncidentRepository.SaveIncidents()
```

### 3. Narrative Generation Flow

```
IncidentRepository.GetIncident()
       │
       ▼
  PromptBuilder.BuildSecurePrompt()
       │
       ▼
  InjectionPreventer.ValidateInput()
       │
       ▼
  Claude API (Opus 4.8)
       │
       ▼
  SourceTracer.LinkToEvents()
       │
       ▼
  NarrativeRepository.SaveNarrative()
```

## Patterns to Follow

### Pattern 1: Event-Driven Pipeline

**What:** Process events through a series of transformations
**When:** When you have sequential processing steps with clear inputs/outputs
**Example:**
```go
type Pipeline struct {
    steps []Step
}

type Step interface {
    Process(ctx context.Context, event *Event) (*Event, error)
}

func (p *Pipeline) Execute(ctx context.Context, event *Event) error {
    for _, step := range p.steps {
        var err error
        event, err = step.Process(ctx, event)
        if err != nil {
            return fmt.Errorf("step %T failed: %w", step, err)
        }
    }
    return nil
}
```

### Pattern 2: Source Tracing

**What:** Link every narrative sentence to its source events
**When:** When you need to prove claims with evidence
**Example:**
```go
type NarrativeSentence struct {
    ID              string    `json:"id"`
    Text            string    `json:"text"`
    Confidence      float64   `json:"confidence"`
    SourceEventIDs  []string  `json:"source_event_ids"`
    CreatedAt       time.Time `json:"created_at"`
}
```

### Pattern 3: Layered Security

**What:** Multiple defense layers against prompt injection
**When:** When processing untrusted content through LLMs
**Example:**
```go
func (s *SecurityService) SanitizeInput(input string) (string, error) {
    // Layer 1: Regex pattern matching
    if s.detectPatterns(input) {
        return "", ErrInjectionDetected
    }
    
    // Layer 2: XML wrapping
    wrapped := s.wrapInXML(input)
    
    // Layer 3: Claude Haiku classifier
    if s.classifyWithHaiku(wrapped) {
        return "", ErrInjectionDetected
    }
    
    return wrapped, nil
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Monolithic Handler

**What:** Putting all logic in a single HTTP handler
**Why bad:** Hard to test, maintain, and scale
**Instead:** Separate concerns into services (Ingestion, Aggregation, Narrative)

### Anti-Pattern 2: String Concatenation for Prompts

**What:** Building prompts by concatenating strings
**Why bad:** Vulnerable to injection, hard to maintain
**Instead:** Use prompt templates with XML wrapping

### Anti-Pattern 3: Trusting LLM Output

**What:** Using LLM output directly without validation
**Why bad:** LLMs can hallucinate or be manipulated
**Instead:** Validate against JSON schema, check for injection patterns

### Anti-Pattern 4: N+1 Query Problem

**What:** Making separate database calls for related data
**Why bad:** Performance degradation at scale
**Instead:** Use JOINs and batch queries

## Scalability Considerations

| Concern | At 100 incidents | At 10K incidents | At 1M incidents |
|---------|------------------|------------------|-----------------|
| Event storage | SQLite sufficient | PostgreSQL needed | Distributed DB |
| Processing | Single goroutine | Worker pool | Distributed queue |
| LLM API | Sequential calls | Concurrent calls | Rate limiting |
| Frontend | Server rendering | SSR + caching | CDN + edge |

## Security Architecture

### Prompt Injection Prevention

```
┌─────────────────────────────────────────┐
│            User Input                   │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│     Layer 1: Input Validation           │
│  - Regex pattern matching              │
│  - Known injection patterns            │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│     Layer 2: XML Wrapping               │
│  - Wrap untrusted content               │
│  - Clear data boundaries                │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│     Layer 3: Haiku Classifier           │
│  - Secondary detection                  │
│  - Fast response time                   │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│     Layer 4: Output Validation          │
│  - JSON schema enforcement              │
│  - Anomaly detection                    │
└─────────────────────────────────────────┘
```

### Data Flow Security

- All API keys stored in environment variables
- SQLite database encrypted at rest (if needed)
- HTTPS for all external API calls
- Input sanitization at API boundary
- Output validation before display

## Testing Strategy

| Layer | Test Type | Tools |
|-------|-----------|-------|
| Unit | Service logic | Go testing, testify |
| Integration | API endpoints | httptest, testcontainers |
| Security | Injection prevention | Custom red-team tests |
| E2E | User workflows | Playwright (if time permits) |

---

*Last Updated: 2026-06-08*
*Research Mode: Ecosystem*
*Confidence: HIGH*
