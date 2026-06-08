# Requirements: NarratorAI

**Defined:** 2026-06-08
**Core Value:** Transform hours of manual alert correlation into seconds of AI-generated narratives — with every claim traceable to its source event.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Data Ingestion

- [ ] **ING-01**: System can pull security logs from Graylog via REST API
- [ ] **ING-02**: System stores raw events in SQLite database with full metadata
- [ ] **ING-03**: System parses and normalizes event fields (timestamp, source IP, destination, user, action, raw message)

### Event Processing

- [ ] **PROC-01**: System groups related events into incidents by source IP within 15-minute time window
- [ ] **PROC-02**: System creates incident records with start/end times and unique users/IPs
- [ ] **PROC-03**: System tracks event count and severity indicators per incident

### MITRE ATT&CK

- [ ] **MITR-01**: System maps Windows EventCodes to ATT&CK techniques via STIX JSON lookup
- [ ] **MITR-02**: System displays technique IDs (T1110, T1021, etc.) in incident metadata
- [ ] **MITR-03**: System supports at least 10 common attack patterns (brute force, lateral movement, exfiltration, privilege escalation)

### Narrative Generation

- [ ] **NARR-01**: System generates chronological attack narratives via Claude API
- [ ] **NARR-02**: System includes confidence score (0.0-1.0) per narrative based on event coverage
- [ ] **NARR-03**: System links every narrative claim to source event IDs via source tracing
- [ ] **NARR-04**: System uses low temperature (0.1-0.3) for factual accuracy

### Security

- [ ] **SECU-01**: System sanitizes all user-controlled fields before entering LLM prompt
- [ ] **SECU-02**: System prevents prompt injection via XML wrapping and pattern detection
- [ ] **SECU-03**: System validates narrative output against source data (cited events must exist)

### Dashboard

- [ ] **DASH-01**: System displays incident list view with severity indicators
- [ ] **DASH-02**: System shows story cards with narrative sentences and confidence scores
- [ ] **DASH-03**: System provides raw event viewer with source tracing (hover sentence → see raw event)
- [ ] **DASH-04**: System captures analyst feedback (thumbs up/down + notes)

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Real-time Features

- **REAL-01**: System streams narratives in real-time via Claude API streaming
- **REAL-02**: System updates dashboard as new events arrive

### Export

- **EXPO-01**: System exports narratives as Markdown files
- **EXPO-02**: System exports narratives as PDF reports

### Multi-SIEM Support

- **MULT-01**: System supports Splunk as additional data source
- **MULT-02**: System supports ElasticSearch as additional data source

### Authentication

- **AUTH-01**: System requires user login
- **AUTH-02**: System tracks analyst feedback per user

## Out of Scope

| Feature | Reason |
|---------|--------|
| Real-time alert streaming | Batch processing sufficient for v1 demo |
| Multi-SIEM support | Architecture supports expansion, demo uses Graylog only |
| User authentication | Single analyst view for demo |
| Mobile app | Web-only for v1 |
| Custom ML models | Leverage Claude API instead |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| ING-01 | Phase 1 | Pending |
| ING-02 | Phase 1 | Pending |
| ING-03 | Phase 1 | Pending |
| PROC-01 | Phase 2 | Pending |
| PROC-02 | Phase 2 | Pending |
| PROC-03 | Phase 2 | Pending |
| MITR-01 | Phase 2 | Pending |
| MITR-02 | Phase 2 | Pending |
| MITR-03 | Phase 2 | Pending |
| NARR-01 | Phase 3 | Pending |
| NARR-02 | Phase 3 | Pending |
| NARR-03 | Phase 3 | Pending |
| NARR-04 | Phase 3 | Pending |
| SECU-01 | Phase 3 | Pending |
| SECU-02 | Phase 3 | Pending |
| SECU-03 | Phase 3 | Pending |
| DASH-01 | Phase 4 | Pending |
| DASH-02 | Phase 4 | Pending |
| DASH-03 | Phase 4 | Pending |
| DASH-04 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0 ✓

---
*Requirements defined: 2026-06-08*
*Last updated: 2026-06-08 after initial definition*
