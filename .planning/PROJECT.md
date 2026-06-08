# NarratorAI

## What This Is

A security incident narrative generator that transforms raw alerts from SIEM systems into chronological attack stories. Built for SOC analysts who spend 60-70% of their time correlating disconnected alerts into timelines. The app ingests security events, uses AI to produce plain-English narratives with MITRE ATT&CK mapping, and displays them as story cards with source tracing back to raw events.

## Core Value

Transform hours of manual alert correlation into seconds of AI-generated narratives — with every claim traceable to its source event.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Security log ingestion from Graylog via REST API
- [ ] Event aggregation: group related events into incidents by IP/time window
- [ ] MITRE ATT&CK mapping via STIX JSON lookup
- [ ] LLM narrative generation with hallucination prevention
- [ ] Prompt injection sanitization for user-controlled fields
- [ ] Source tracing: each narrative sentence links to raw source events
- [ ] Confidence scoring based on event coverage
- [ ] Story card dashboard with incident list view
- [ ] Analyst annotation (thumbs up/down + notes)
- [ ] Go backend API (Gin framework)
- [ ] Next.js frontend with TypeScript + Tailwind

### Out of Scope

- **Splunk integration** — Using Graylog instead (free, open source)
- **Real-time alert streaming** — Batch processing for v1
- **Multi-SIEM support** — Graylog only for v1, architecture supports expansion
- **User authentication** — Single analyst view for demo
- **Mobile app** — Web-only for v1

## Context

**Problem Space:**
- SOC analysts face hundreds of disconnected alerts daily
- 60-70% of time spent correlating events, not responding
- Existing SIEM tools show raw alert tables, not narratives
- Splunk's built-in AI doesn't provide source tracing or analyst feedback loops

**Hackathon Constraints:**
- 3-week timeline (solo full-time)
- Demo with real attack data (botsv3 or Graylog sample datasets)
- Public GitHub repo required
- Demo video required (under 3 minutes)

**Tech Decisions:**
- Go backend chosen over Python for performance (goroutines for parallel alert processing)
- Graylog over Splunk for zero cost, open source, simpler REST API
- Next.js + Tailwind for rapid frontend development
- Claude/OpenAI API for narrative generation
- Prompt injection sanitization as core security feature

## Constraints

- **Timeline:** 3 weeks solo full-time — hackathon deadline
- **Budget:** Zero — all tools must be free/open source
- **Demo Data:** Must use real attack datasets, not made-up JSON
- **Security:** Must handle prompt injection and hallucination prevention
- **Open Source:** MIT license for submission

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go + Gin backend | Goroutines for parallel alert processing, single binary deployment | — Pending |
| Graylog over Splunk | Free, open source, simpler REST API, zero cost | — Pending |
| Claude/OpenAI API | Best narrative quality for security context | — Pending |
| Source tracing as differentiator | Splunk's AI doesn't link claims to evidence | — Pending |
| Prompt injection sanitization | Highest-ROI security fix for judge score | — Pending |

---

*Last updated: 2026-06-08 after initialization*
