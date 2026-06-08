# Project Research Summary

**Project:** NarratorAI - Security Incident Narrative Generator
**Domain:** Security / SIEM / AI-Powered Analysis
**Researched:** 2026-06-08
**Confidence:** HIGH

## Executive Summary

NarratorAI is a security incident narrative generator that transforms raw SIEM alerts (from Graylog) into human-readable attack stories using Claude AI. The product targets SOC analysts who currently spend hours correlating raw alerts into narratives. The recommended approach is a Go backend (Gin framework) handling event ingestion, aggregation, and LLM orchestration, paired with a Next.js 16 frontend displaying "story cards" — each narrative sentence linked back to source events via a unique source-tracing mechanism.

The architecture follows an event-driven pipeline: Graylog alerts are ingested via REST API, grouped into incidents using time-window/IP clustering, mapped to MITRE ATT&CK techniques via STIX data, then fed through Claude (Opus 4.8) with multi-layer prompt injection prevention. The critical differentiator over existing tools (Splunk AI, etc.) is source tracing — every narrative claim links to specific raw event IDs, enabling analyst verification and compliance audit trails.

The top risks are prompt injection (OWASP #1 LLM vulnerability — security logs contain attacker-controlled fields that could hijack the LLM), LLM hallucination (Claude generating plausible but false narratives), and source tracing failure (narrative sentences losing their links to raw evidence). All three are mitigated through layered defenses: input validation, XML wrapping, Haiku classifier for injection detection, low-temperature generation for factual accuracy, and explicit source-event linking in the database schema. The 3-week hackathon timeline is feasible given SQLite for zero-config storage and batch processing (no real-time requirements for v1).

## Key Findings

### Recommended Stack

The backend uses Go 1.22+ with Gin v1.12.0 for high-performance concurrent event processing (goroutines enable 10-100x parallelism). SQLite with sqlc provides type-safe, zero-config storage ideal for hackathon demos. The frontend is Next.js 16.2.x with React 19, TypeScript 5, and Tailwind CSS 4.3.x for rapid dashboard development. Claude API (Anthropic Go SDK v1.35.0) provides narrative generation, with Opus 4.8 for quality and Haiku 4.5 for fast injection classification. MITRE ATT&CK STIX data is parsed via panther-labs/stix2.

**Core technologies:**
- **Go + Gin**: High-performance backend with goroutine-based concurrency for event processing
- **SQLite + sqlc**: Zero-config database with type-safe query generation (rejected GORM for lower overhead)
- **Next.js 16 + React 19**: Frontend with SSR for fast initial load, App Router for clean API routes
- **Claude Opus 4.8**: Primary narrative generation model (better security context quality than OpenAI)
- **Claude Haiku 4.5**: Fast, cost-effective injection classifier (secondary security layer)
- **MITRE ATT&CK STIX**: Industry-standard framework data via panther-labs/stix2 parser

### Expected Features

**Must have (table stakes):**
- Chronological timeline — core value proposition transforming alerts into stories
- MITRE ATT&CK mapping — industry standard analysts expect
- Source tracing — every narrative claim links to raw event IDs (key differentiator)
- Confidence scoring — analysts need to trust the narrative
- Raw event viewer — ability to drill down into details
- Incident list view — dashboard to browse all incidents
- Analyst feedback — thumbs up/down for training data
- Alert ingestion — pull from Graylog REST API

**Should have (competitive):**
- Prompt injection prevention — security-focused, builds trust
- Real-time narrative streaming — better UX for long narratives (Claude API streaming)
- Event correlation engine — groups related events by IP/time window
- Narrative templates — customizable output format
- Export capabilities — share narratives as PDF/Markdown

**Defer (v2+):**
- Multi-SIEM support — architecture supports future, demo uses Graylog only
- User authentication — single analyst view for demo
- Mobile app — responsive web-only
- Custom ML models — leverage Claude API instead

### Architecture Approach

The system follows a 4-layer event-driven pipeline: Data Ingestion (GraylogClient → EventParser → EventRepository), Event Aggregation (EventAggregator → IncidentClustering → MitreMapper), Narrative Generation (NarrativeService → PromptBuilder → InjectionPreventer → SourceTracer), and Presentation (StoryCard, IncidentList, RawEventViewer, FeedbackForm in Next.js). Key patterns include event-driven pipeline processing, source tracing (every NarrativeSentence stores source_event_ids), and layered security (4-layer injection prevention). Anti-patterns to avoid: monolithic handlers, string concatenation for prompts, trusting LLM output without validation, and N+1 query problems.

**Major components:**
1. **Data Ingestion Layer** — GraylogClient, EventParser, EventRepository (fetch and store raw alerts)
2. **Event Aggregation Layer** — EventAggregator, IncidentClustering, MitreMapper (group events, map to ATT&CK)
3. **Narrative Generation Layer** — NarrativeService, PromptBuilder, SourceTracer, InjectionPreventer (LLM orchestration with security)
4. **Presentation Layer** — StoryCard, IncidentList, RawEventViewer, FeedbackForm (Next.js dashboard)

### Critical Pitfalls

1. **Prompt Injection Vulnerability** — Security logs contain attacker-controlled fields (IPs, hostnames, usernames) that could hijack Claude. Prevention: 4-layer defense (regex validation → XML wrapping → Haiku classifier → output schema enforcement). CRITICAL priority.
2. **LLM Hallucination** — Claude may generate plausible but false narratives. Prevention: source tracing (every sentence links to event IDs), low temperature (0.1-0.3), structured JSON output. CRITICAL priority.
3. **Source Tracing Failure** — Narrative sentences lose links to raw events, destroying audit capability. Prevention: explicit source tracking in DB schema, require LLM to cite event IDs, validate cited events exist. CRITICAL priority.
4. **Graylog API Compatibility** — Client libraries are outdated (v2.4). Prevention: test early with actual instance, use standard HTTP client instead of outdated libraries. MODERATE priority.
5. **SQLite Performance at Scale** — Can struggle with concurrent access. Prevention: WAL mode, busy timeout, batch operations, index optimization. MODERATE priority.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Foundation & Data Pipeline
**Rationale:** Must establish data layer before any AI integration. Graylog API compatibility is a known risk — test early. SQLite schema defines all entities for downstream features.
**Delivers:** Go project structure, SQLite database with migrations, Graylog REST API client, raw event ingestion and storage.
**Addresses:** Alert ingestion, raw event storage, database schema (RawEvent, Incident, NarrativeSentence, MitreMapping, AnalystFeedback entities).
**Avoids:** Graylog API compatibility pitfall — test early with real instance, use standard HTTP client.
**Stack:** Go + Gin, SQLite + sqlc, golang-migrate.
**Research needed:** Graylog 5.x API specifics (research-phase recommended).

### Phase 2: Event Aggregation & MITRE Mapping
**Rationale:** Events must be grouped into incidents before narrative generation. STIX data loading and ATT&CK mapping are prerequisites for the narrative service.
**Delivers:** Event aggregation engine (time-window/IP clustering), incident creation, MITRE ATT&CK technique mapping via STIX data.
**Addresses:** Event correlation engine, MITRE ATT&CK mapping, incident list view.
**Avoids:** Event aggregation logic pitfalls — use multiple grouping criteria, configurable thresholds.
**Stack:** panther-labs/stix2, goattck.
**Research needed:** STIX data parsing edge cases (research-phase recommended for STIX integration details).

### Phase 3: Narrative Generation & Security
**Rationale:** Core differentiator. Depends on Phase 1 (events in DB) and Phase 2 (incidents with MITRE mappings). Must implement prompt injection prevention from day one — not bolted on later.
**Delivers:** Claude API integration, prompt injection prevention (4-layer defense), source tracing mechanism, narrative generation with confidence scoring.
**Addresses:** Source tracing (key differentiator), prompt injection prevention, confidence scoring, chronological timeline, real-time narrative streaming.
**Avoids:** Prompt injection vulnerability, LLM hallucination, source tracing failure — all three critical pitfalls addressed in this phase.
**Stack:** Anthropic Go SDK, Claude Opus 4.8, Claude Haiku 4.5.
**Research needed:** None — well-documented Claude API patterns, OWASP guidelines for injection prevention.

### Phase 4: Frontend Dashboard
**Rationale:** Presentation layer depends on all backend services being functional. Build after API contracts are stable.
**Delivers:** Story card components, incident list view, raw event viewer with source links, analyst feedback UI, responsive dashboard.
**Addresses:** Story card dashboard, raw event viewer, analyst feedback system, export capabilities (low priority).
**Avoids:** Frontend state management pitfalls — use React Query, error boundaries.
**Stack:** Next.js 16, React 19, TypeScript, Tailwind CSS 4.
**Research needed:** Standard patterns — Next.js SSR + React Query.

### Phase 5: Polish & Demo Preparation
**Rationale:** Final phase for performance optimization, red-team testing, demo data preparation.
**Delivers:** Performance tuning (SQLite WAL, API rate limiting), red-team test suite for injection, demo data from botsv3/Graylog datasets, export capabilities.
**Addresses:** Performance under load, security audit, narrative templates, export.
**Avoids:** Claude API rate limits — implement exponential backoff, batch processing.
**Research needed:** None — standard optimization patterns.

### Phase Ordering Rationale

- **Data first (Phase 1):** All downstream features depend on having events in the database. Graylog API is a known risk — validate early.
- **Aggregation before AI (Phase 2):** Narrative generation needs structured incidents, not raw events. MITRE mappings enrich the context for Claude.
- **Security built-in (Phase 3):** Prompt injection prevention cannot be bolted on — it must be part of the narrative generation pipeline from the start.
- **Frontend last (Phase 4):** API contracts must be stable before building UI. Server-side rendering means frontend can be built after backend is functional.
- **Polish final (Phase 5):** Demo prep and performance tuning are natural final phases.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 1:** Graylog 5.x REST API specifics — client libraries are outdated, need to verify API compatibility
- **Phase 2:** STIX 2.x parsing edge cases and goattck library API — less battle-tested than core stack

Phases with standard patterns (skip research-phase):
- **Phase 3:** Claude API integration follows well-documented patterns (Anthropic SDK, OWASP injection prevention)
- **Phase 4:** Next.js 16 + React 19 dashboard is standard web development
- **Phase 5:** Performance optimization and demo prep are well-understood

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified, production-tested libraries, alternatives documented |
| Features | HIGH | Based on SOC analyst workflows, existing SIEM tool analysis, clear table stakes vs differentiators |
| Architecture | HIGH | Event-driven pipeline is proven pattern, component boundaries well-defined, anti-patterns documented |
| Pitfalls | HIGH | Based on OWASP guidelines, LLM security best practices, SQLite performance documentation |

**Overall confidence:** HIGH

### Gaps to Address

- **Graylog API version compatibility:** REST API client libraries are outdated (v2.4). Need to verify if Graylog 5.x API changes affect implementation. *Handle:* Test early in Phase 1 with actual Graylog instance, use standard HTTP client.
- **STIX data freshness:** How often should ATT&CK updates be pulled? *Handle:* Implement version tracking and manual refresh mechanism, update quarterly.
- **Claude model selection:** Opus 4.8 vs Sonnet for cost/quality tradeoff. *Handle:* Start with Opus for demo quality, profile cost, switch to Sonnet if needed.
- **SQLite scalability:** Acceptable for hackathon, but production would need PostgreSQL. *Handle:* Note in architecture docs, design schema to be PostgreSQL-compatible.
- **Source tracing precision:** How precisely should LLM cite event IDs? *Handle:* Require JSON output with source_event_ids array, validate all cited events exist in DB.

## Sources

### Primary (HIGH confidence)
- Anthropic Go SDK documentation — Claude API integration patterns
- OWASP Top 10 for LLM Applications — prompt injection prevention guidelines
- MITRE ATT&CK framework — STIX data format and technique taxonomy
- Go Gin framework documentation — HTTP routing patterns
- Next.js 16 documentation — App Router, SSR patterns

### Secondary (MEDIUM confidence)
- panther-labs/stix2 library — STIX 2.x parsing (less battle-tested)
- MSAdministrator/goattck — ATT&CK data loading (community library)
- Graylog REST API documentation — endpoint specifications

### Tertiary (LOW confidence)
- botsv3 dataset format — demo data compatibility (needs validation)
- Claude Opus 4.8 specific capabilities — model version details

---
*Research completed: 2026-06-08*
*Ready for roadmap: yes*
