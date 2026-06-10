# Roadmap: NarratorAI

## Overview

NarratorAI transforms raw SIEM alerts into human-readable attack stories. This roadmap delivers the system in 4 vertical slices: first establishing the data pipeline, then grouping events into incidents with MITRE mapping, generating AI narratives with source tracing, and finally presenting everything in a story card dashboard with analyst feedback. Each phase delivers a complete, verifiable capability that builds on the previous.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Data Foundation** - Pull and store security logs from Graylog
- [ ] **Phase 2: Incident Intelligence** - Group events into incidents with MITRE ATT&CK mapping
- [ ] **Phase 3: Narrative Engine** - Generate AI narratives with source tracing and injection prevention
- [ ] **Phase 4: Story Dashboard** - Display incidents as story cards with analyst feedback

## Phase Details

### Phase 1: Data Foundation
**Goal**: Analyst can pull security logs from Graylog and see them stored as normalized events
**Depends on**: Nothing (first phase)
**Requirements**: ING-01, ING-02, ING-03
**Success Criteria** (what must be TRUE):
  1. System connects to Graylog and pulls security logs via REST API
  2. Raw events are persisted in SQLite with full metadata (timestamp, source IP, destination, user, action, raw message)
  3. Event fields are parsed and normalized into a consistent schema
**Plans**: TBD

Plans:
- [ ] 01-1: Graylog REST API Integration — [details](phases/01-data-foundation/01-1-graylog-api.md)
- [ ] 01-2: SQLite Schema Design — [details](phases/01-data-foundation/01-2-sqlite-schema.md)
- [ ] 01-3: Event Field Normalization — [details](phases/01-data-foundation/01-3-normalization.md)

### Phase 2: Incident Intelligence
**Goal**: Analyst sees related events grouped into incidents with MITRE ATT&CK technique labels
**Depends on**: Phase 1
**Requirements**: PROC-01, PROC-02, PROC-03, MITR-01, MITR-02, MITR-03
**Success Criteria** (what must be TRUE):
  1. Events are automatically grouped into incidents by source IP within 15-minute time windows
  2. Each incident tracks start/end times, unique users, unique IPs, and event count
  3. MITRE ATT&CK technique IDs (T1110, T1021, etc.) appear in incident metadata via STIX JSON lookup
  4. At least 10 common attack patterns are mapped (brute force, lateral movement, exfiltration, privilege escalation)
**Plans**: 3 plans

Plans:
- [ ] 02-01: Database Schema + ATT&CK Mapping Engine — [details](phases/02-incident-intelligence/02-01-PLAN.md)
- [ ] 02-02: Incident Grouping Engine + Severity — [details](phases/02-incident-intelligence/02-02-PLAN.md)
- [ ] 02-03: API Layer + End-to-End Verification — [details](phases/02-incident-intelligence/02-03-PLAN.md)

### Phase 3: Narrative Engine
**Goal**: Analyst receives AI-generated attack narratives with confidence scores, source tracing, and injection protection
**Depends on**: Phase 2
**Requirements**: NARR-01, NARR-02, NARR-03, NARR-04, SECU-01, SECU-02, SECU-03
**Success Criteria** (what must be TRUE):
  1. System generates chronological attack narratives via Claude API with low temperature (0.1-0.3)
  2. Every narrative sentence links to its source event IDs — hover reveals raw evidence
  3. Each narrative displays a confidence score (0.0-1.0) based on event coverage
  4. User-controlled fields are sanitized before entering LLM prompts (XML wrapping + pattern detection)
  5. Narrative output is validated: all cited events must exist in the database
**Plans**: 3 plans

Plans:
- [ ] 03-01: Database Schema + LLM Client Foundation — [details](phases/03-narrative-engine/03-01-PLAN.md)
- [ ] 03-02: Security Pipeline + Narrative Generator — [details](phases/03-narrative-engine/03-02-PLAN.md)
- [ ] 03-03: API Layer + Integration — [details](phases/03-narrative-engine/03-03-PLAN.md)

### Phase 4: Story Dashboard
**Goal**: Analyst browses incidents, reads story cards, traces sources, and provides feedback
**Depends on**: Phase 3
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04
**Success Criteria** (what must be TRUE):
  1. Incident list view displays all incidents with severity indicators
  2. Story cards show narrative sentences with confidence scores and technique labels
  3. Raw event viewer provides source tracing — hovering a sentence reveals the raw source event
  4. Analyst can submit feedback (thumbs up/down + notes) on narrative quality
**Plans**: TBD

Plans:
- [ ] 04-01: TBD
- [ ] 04-02: TBD
- [ ] 04-03: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Data Foundation | 0/3 | Not started | - |
| 2. Incident Intelligence | 0/3 | Planned | - |
| 3. Narrative Engine | 0/3 | Not started | - |
| 4. Story Dashboard | 0/3 | Not started | - |
