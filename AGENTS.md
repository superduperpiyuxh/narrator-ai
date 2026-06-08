<!-- GSD:project-start source:PROJECT.md -->

## Project

**NarratorAI**

A security incident narrative generator that transforms raw alerts from SIEM systems into chronological attack stories. Built for SOC analysts who spend 60-70% of their time correlating disconnected alerts into timelines. The app ingests security events, uses AI to produce plain-English narratives with MITRE ATT&CK mapping, and displays them as story cards with source tracing back to raw events.

**Core Value:** Transform hours of manual alert correlation into seconds of AI-generated narratives — with every claim traceable to its source event.

### Constraints

- **Timeline:** 3 weeks solo full-time — hackathon deadline
- **Budget:** Zero — all tools must be free/open source
- **Demo Data:** Must use real attack datasets, not made-up JSON
- **Security:** Must handle prompt injection and hallucination prevention
- **Open Source:** MIT license for submission

<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->

## Technology Stack

## Executive Summary

- Go backend for high-performance event processing and AI integration
- Next.js frontend with TypeScript for the analyst dashboard
- SQLite database for local development and demo
- Claude API integration with prompt injection prevention
- MITRE ATT&CK STIX data parsing and mapping

## Core Technology Stack

### 1. Backend Framework

| Component | Version | Purpose | Confidence |
|-----------|---------|---------|------------|
| **Go** | 1.22+ | Primary language | HIGH |
| **Gin** | v1.12.0 | HTTP framework | HIGH |
| **SQLite** | 3.x | Local database | HIGH |
| **sqlc** | v1.30.0 | Type-safe SQL queries | HIGH |
| **golang-migrate** | v4 | Database migrations | HIGH |

- Go provides excellent performance for concurrent alert processing via goroutines
- Gin offers high-performance HTTP routing with minimal overhead (up to 40x faster than Martini)
- SQLite eliminates external database dependencies for demo/hackathon
- sqlc generates type-safe code from SQL, eliminating runtime type errors
- golang-migrate handles schema versioning cleanly

### 2. Frontend Stack

| Component | Version | Purpose | Confidence |
|-----------|---------|---------|------------|
| **Next.js** | 16.2.x | React framework | HIGH |
| **React** | 19.x | UI library | HIGH |
| **TypeScript** | 5.x | Type safety | HIGH |
| **Tailwind CSS** | 4.3.x | Styling | HIGH |

- Next.js 16 with App Router provides excellent developer experience and API routes
- React 19 offers server components and improved performance
- TypeScript catches errors at compile time, critical for complex data structures
- Tailwind CSS 4 provides rapid UI development with utility classes

### 3. Database & ORM

| Component | Version | Purpose | Confidence |
|-----------|---------|---------|------------|
| **SQLite** | 3.x | Embedded database | HIGH |
| **sqlc** | v1.30.0 | Query code generation | HIGH |
| **mattn/go-sqlite3** | v1.14.45 | SQLite driver (CGO) | HIGH |

- SQLite is perfect for hackathon: zero config, single file, embedded
- sqlc generates type-safe Go code from SQL queries (better than ORMs for this use case)
- mattn/go-sqlite3 is the de-facto standard SQLite driver with excellent performance
- **GORM**: Rejected because it generates suboptimal SQL and has higher memory overhead
- **pgx/PostgreSQL**: Overkill for hackathon demo (adds complexity)

### 4. AI/LLM Integration

| Component | Version | Purpose | Confidence |
|-----------|---------|---------|------------|
| **Anthropic Go SDK** | v1.35.0 | Claude API client | HIGH |
| **Claude Opus 4.8** | - | Primary model for narratives | HIGH |
| **Claude Haiku 4.5** | - | Injection classifier | HIGH |

- Anthropic's official Go SDK provides excellent support with retries and rate limiting
- Claude Opus 4.8 offers best narrative quality for security context
- Claude Haiku 4.5 is fast and cost-effective for injection detection

### 5. Security Incident Data

| Component | Version | Purpose | Confidence |
|-----------|---------|---------|------------|
| **MITRE ATT&CK STIX** | - | Framework data | HIGH |
| **panther-labs/stix2** | v0.1.1 | STIX parsing | MEDIUM |
| **MSAdministrator/goattck** | - | ATT&CK data loading | MEDIUM |

- panther-labs/stix2 provides clean STIX 2.x parsing with FromJSON helper
- goattck offers convenient ATT&CK data models (Technique, Tactic, Actor, etc.)
- Both are pure Go with no CGO dependencies

### 6. Prompt Injection Prevention

| Component | Purpose | Confidence |
|-----------|---------|------------|
| **XML Wrapping** | Separate trusted/untrusted content | HIGH |
| **Input Validation** | Regex + pattern matching | HIGH |
| **Claude Haiku Classifier** | Secondary injection detection | HIGH |
| **Output Validation** | JSON schema enforcement | HIGH |

- Layered defense is essential (OWASP #1 LLM vulnerability)
- XML wrapping leverages Claude's training to treat tagged content as data
- Haiku classifier adds security without significant latency/cost

## Complete Dependency List

### Go Backend

### Frontend (Next.js)

## Alternatives Considered & Rejected

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Backend | Go + Gin | Python + FastAPI | Go chosen for goroutine performance |
| Database | SQLite | PostgreSQL | SQLite simpler for demo, zero config |
| ORM | sqlc | GORM | sqlc is faster, generates less boilerplate |
| Frontend | Next.js 16 | Vite + React | Next.js offers better DX and API routes |
| Styling | Tailwind 4 | CSS Modules | Tailwind faster for rapid prototyping |
| AI | Claude | OpenAI | Claude better for security narrative quality |
| STIX Parser | panther-labs/stix2 | Custom | Standard library exists, no need to reinvent |

## Key Architecture Decisions

### 1. Event Processing Pipeline

### 2. Source Tracing Implementation

- Each narrative sentence links to specific raw event IDs
- Event IDs stored in database with timestamps
- Frontend shows clickable links from narrative to raw events
- Confidence scores based on event coverage

### 3. Prompt Injection Prevention Strategy

## Performance Considerations

| Concern | Solution | Impact |
|---------|----------|--------|
| Concurrent alert processing | Go goroutines | 10-100x parallelism |
| SQLite write performance | WAL mode + busy timeout | 10x improvement |
| LLM API latency | Streaming responses | Better UX |
| Frontend rendering | Next.js SSR/SSG | Faster initial load |

## Security Considerations

| Concern | Mitigation | Priority |
|---------|------------|----------|
| Prompt injection | Multi-layer defense | CRITICAL |
| API key exposure | Environment variables only | HIGH |
| SQL injection | sqlc parameterized queries | HIGH |
| XSS attacks | React auto-escaping + CSP | MEDIUM |
| Rate limiting | API middleware | MEDIUM |

## Development Environment Setup

### Prerequisites

# Go 1.22+

# Node.js 20+ (for Next.js 16)

# SQLite (usually pre-installed)

# sqlc (for code generation)

### Quick Start Commands

# Backend

# Frontend

## Verification Checklist

- [x] All versions verified against official documentation
- [x] Alternative libraries researched and documented
- [x] Performance implications considered
- [x] Security requirements addressed
- [x] Hackathon constraints (3 weeks, zero budget) respected
- [x] Demo data compatibility verified (botsv3/Graylog datasets)

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Go Backend | HIGH | Gin is production-tested, excellent documentation |
| Next.js Frontend | HIGH | Latest version with stable App Router |
| SQLite Database | HIGH | Perfect for hackathon, zero config |
| Claude API | HIGH | Official SDK, well-documented |
| STIX Parsing | MEDIUM | Libraries exist but less battle-tested |
| Graylog Integration | MEDIUM | REST API works but client libraries outdated |
| Prompt Injection Defense | HIGH | Layered approach based on OWASP best practices |

## Open Questions

## Next Steps

<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->

## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->

## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->

## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->

## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:

- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->

## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
