# Phase 3: Narrative Engine — Execution Plan Summary

**Created:** 2026-06-10
**Status:** Ready to execute
**Plans:** 3 plans in 3 waves

---

## Overview

Phase 3 transforms structured incidents with MITRE ATT&CK mappings into human-readable attack narratives with source tracing and confidence scores. The architecture uses OpenRouter API (Claude-compatible) for LLM calls with a layered security approach.

**Core Architecture:**
```
Incident Data → Input Validation → XML Wrapping → LLM Generation → Output Validation → Database Storage
     ↓              ↓                  ↓                ↓                ↓                ↓
   Events      Pattern Detection   Untrusted     Structured JSON    Verify Cited    Narratives Table
  (Phase 2)    (7 patterns)        Content       Output with        Events Exist    with JSON Sentences
                                    Separated     Source Tracing     in Database
```

---

## Wave Structure

| Wave | Plan | Objective | Tasks | Dependencies |
|------|------|-----------|-------|--------------|
| 1 | 03-01 | Database Schema + LLM Client Foundation | 2 | None |
| 2 | 03-02 | Security Pipeline + Narrative Generator | 2 | 03-01 |
| 3 | 03-03 | API Layer + Integration | 2 | 03-01, 03-02 |

---

## Plan Details

### Plan 03-01: Database Schema + LLM Client Foundation
**Wave:** 1 | **Tasks:** 2 | **Requirements:** NARR-01, NARR-04

**Objective:** Establish the database schema for narratives and create the OpenRouter LLM client foundation.

**Tasks:**
1. **Database Migration — Narrative Schema**
   - Add `narratives` table to sqlite.go
   - JSON sentences column for flexible sentence structure
   - Foreign key to incidents table with CASCADE delete
   - Indexes for incident_id and confidence

2. **Config + LLM Client — OpenRouter Integration**
   - Add `OPENROUTER_API_KEY` to config
   - Install `hra42/openrouter-go v1.7.0`
   - Create LLM client wrapper with temperature 0.2 (NARR-04 compliant)
   - Handle retries and error types

**Files Modified:**
- `backend/internal/database/sqlite.go`
- `backend/internal/config/config.go`
- `backend/go.mod`
- `backend/internal/llm/client.go`
- `backend/internal/llm/client_test.go`

---

### Plan 03-02: Security Pipeline + Narrative Generator
**Wave:** 2 | **Tasks:** 2 | **Requirements:** SECU-01, SECU-02, SECU-03, NARR-02, NARR-03

**Objective:** Build the 4-layer security pipeline and narrative generator with source tracing.

**Tasks:**
1. **Security Pipeline — Injection Detection + XML Wrapping**
   - Create `security/injection.go` with 7 injection patterns
   - Create `narrative/sanitizer.go` with XML wrapping
   - Escape XML special characters in event data
   - Validate no injection patterns in event data

2. **Narrative Generator + Source Tracing + Confidence Scoring**
   - Create `narrative/generator.go` with structured JSON output
   - Create `narrative/validator.go` for output validation
   - Create `narrative/confidence.go` with 4-factor weighted algorithm
   - Link every sentence to source event IDs (NARR-03)

**Files Modified:**
- `backend/internal/security/injection.go`
- `backend/internal/security/injection_test.go`
- `backend/internal/narrative/sanitizer.go`
- `backend/internal/narrative/sanitizer_test.go`
- `backend/internal/narrative/generator.go`
- `backend/internal/narrative/generator_test.go`
- `backend/internal/narrative/validator.go`
- `backend/internal/narrative/validator_test.go`
- `backend/internal/narrative/confidence.go`
- `backend/internal/narrative/confidence_test.go`

---

### Plan 03-03: API Layer + Integration
**Wave:** 3 | **Tasks:** 2 | **Requirements:** NARR-01, NARR-02, NARR-03, NARR-04, SECU-01, SECU-02, SECU-03

**Objective:** Add REST API endpoints for narrative generation and integrate all components.

**Tasks:**
1. **Database Operations + Narrative Handler**
   - Create `database/narratives.go` with Save/Get operations
   - Create `handler/narrative_handler.go` with 3 endpoints
   - Validate incident exists before generation
   - Validate output before storage

2. **Route Registration + Integration Tests**
   - Register routes in main.go
   - Create handler tests
   - Create integration test script

**Files Modified:**
- `backend/internal/database/narratives.go`
- `backend/internal/handler/narrative_handler.go`
- `backend/internal/handler/narrative_handler_test.go`
- `backend/main.go`

---

## Database Schema

### Narratives Table
```sql
CREATE TABLE IF NOT EXISTS narratives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    incident_id INTEGER NOT NULL,
    summary TEXT NOT NULL,
    confidence REAL NOT NULL DEFAULT 0.0,
    sentences JSON NOT NULL,
    model_used TEXT NOT NULL,
    temperature REAL NOT NULL,
    tokens_used INTEGER DEFAULT 0,
    generation_time_ms INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE
);
```

### Narrative Sentences (JSON Structure)
```json
{
    "sentences": [
        {
            "text": "At 14:23:45, the attacker initiated a brute force attack from IP 10.1.50.76.",
            "timestamp": "2025-12-21T14:23:45Z",
            "source_event_ids": [12345, 12346, 12347],
            "confidence": 0.95,
            "technique": "T1110"
        }
    ]
}
```

---

## LLM Integration Approach

### Client Setup
- **SDK:** `hra42/openrouter-go v1.7.0`
- **Primary Model:** `anthropic/claude-3-opus`
- **Temperature:** 0.2 (within NARR-04 range of 0.1-0.3)
- **Max Tokens:** 4096
- **Structured Output:** JSON schema enforcement

### Prompt Template
- **System Prompt:** Rules for secure narrative generation
- **User Prompt:** Incident data wrapped in XML tags
- **Output Format:** Structured JSON with sentences linked to event IDs

---

## Security Implementation

### 4-Layer Defense Pipeline

| Layer | Component | Purpose |
|-------|-----------|---------|
| 1 | Pattern Detection | Block 7 known injection patterns |
| 2 | XML Wrapping | Separate trusted/untrusted content |
| 3 | Output Validation | Verify cited events exist in database |
| 4 | Confidence Scoring | 4-factor weighted algorithm |

### Injection Patterns Detected
1. `ignore previous instructions`
2. `disregard.*system prompt`
3. `you are now in (admin|developer|DAN) mode`
4. `` ```system ``
5. `[INST]`
6. `### Instruction:`
7. `base64:[A-Za-z0-9+/=]{20,}`

### XML Wrapping
```xml
<event_data>
<event id="12345" timestamp="2025-12-21T14:23:45Z" source_ip="10.1.50.76" event_type="login_failed" details="{...}"/>
</event_data>
```

---

## API Endpoints

### POST /api/incidents/:id/narrative
**Purpose:** Generate narrative for an incident
**Request:** `{ }`
**Response:**
```json
{
    "narrative": {
        "id": 1,
        "incident_id": 123,
        "summary": "Brute force attack from 10.1.50.76...",
        "confidence": 0.87,
        "sentences": [...],
        "model_used": "anthropic/claude-3-opus",
        "temperature": 0.2,
        "tokens_used": 1234,
        "generation_time_ms": 2345
    },
    "cached": false
}
```

### GET /api/incidents/:id/narrative
**Purpose:** Get existing narrative for an incident
**Response:** `{ "narrative": {...} }` or 404

### GET /api/narratives/:id
**Purpose:** Get narrative detail with source events
**Response:**
```json
{
    "narrative": {...},
    "source_events": {
        "12345": {...},
        "12346": {...}
    }
}
```

---

## Implementation Order

1. **Plan 03-01** (Wave 1): Database + LLM Client
   - Narrative table schema
   - Config updates
   - OpenRouter SDK installation
   - LLM client wrapper

2. **Plan 03-02** (Wave 2): Security + Generator
   - Injection detection
   - XML wrapping
   - Narrative generator
   - Output validation
   - Confidence scoring

3. **Plan 03-03** (Wave 3): API + Integration
   - Database operations
   - Narrative handler
   - Route registration
   - Integration tests

---

## Verification Steps

### Per Plan
1. `go build ./...` — Backend compiles
2. `go test ./internal/... -v` — Unit tests pass
3. `go test ./... -v` — Full suite green

### End-to-End
1. Start server: `./start.sh`
2. Generate narrative: `curl -X POST http://localhost:8080/api/incidents/1/narrative`
3. Verify response contains:
   - `summary` field
   - `confidence` between 0.0-1.0
   - `sentences` array with `source_event_ids`
4. Retrieve narrative: `curl http://localhost:8080/api/incidents/1/narrative`
5. Verify source events: `curl http://localhost:8080/api/narratives/1`

---

## Success Criteria

### Phase 3 Requirements Coverage

| Requirement | Plan | Status |
|-------------|------|--------|
| NARR-01 | 03-01, 03-03 | ✅ Covered |
| NARR-02 | 03-02, 03-03 | ✅ Covered |
| NARR-03 | 03-02, 03-03 | ✅ Covered |
| NARR-04 | 03-01, 03-03 | ✅ Covered |
| SECU-01 | 03-02 | ✅ Covered |
| SECU-02 | 03-02 | ✅ Covered |
| SECU-03 | 03-02, 03-03 | ✅ Covered |

### Must-Have Truths

1. ✅ Narrative table exists with JSON sentences column
2. ✅ OpenRouter API key loads from environment variable
3. ✅ LLM client wrapper successfully connects to OpenRouter API
4. ✅ Temperature parameter set to 0.1-0.3 for factual accuracy
5. ✅ 7 injection patterns detected in security package
6. ✅ XML wrapping properly escapes event data
7. ✅ Every narrative sentence links to source event IDs
8. ✅ Confidence scoring calculates 0.0-1.0 using 4-factor weighted algorithm
9. ✅ Output validation verifies cited events exist in database
10. ✅ POST /api/incidents/:id/narrative generates AI narrative
11. ✅ GET /api/incidents/:id/narrative retrieves existing narrative
12. ✅ GET /api/narratives/:id returns narrative with source events

---

## Environment Setup

### Required Environment Variables
```bash
export OPENROUTER_API_KEY="sk-or-v1-your-key-here"
```

### Dependencies to Install
```bash
cd backend
go get github.com/hra42/openrouter-go@v1.7.0
```

---

## Next Steps

1. **Execute Plan 03-01:** Database + LLM Client (Wave 1)
2. **Execute Plan 03-02:** Security + Generator (Wave 2)
3. **Execute Plan 03-03:** API + Integration (Wave 3)
4. **Verify:** Run integration tests
5. **Proceed to Phase 4:** Story Dashboard

---

*Plan created: 2026-06-10*
*Ready to execute with: `/gsd-execute-phase 03`*
