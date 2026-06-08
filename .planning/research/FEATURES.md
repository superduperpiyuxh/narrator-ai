# Feature Landscape

**Domain:** Security Incident Narrative Generator
**Research Date:** 2026-06-08
**Research Mode:** Ecosystem

## Executive Summary

This document maps the complete feature landscape for a security incident narrative generator. Features are categorized as table stakes (must-have for MVP), differentiators (competitive advantages), and anti-features (explicitly NOT to build). All recommendations are based on analysis of existing SIEM tools, SOC analyst workflows, and hackathon constraints.

## Table Stakes

Features users expect. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Chronological timeline | Core value proposition - transforms alerts into stories | Medium | Group by incident, sort by time |
| MITRE ATT&CK mapping | Industry standard framework analysts know | Medium | STIX data parsing required |
| Source tracing | Every claim links to raw evidence | High | Differentiator from Splunk AI |
| Confidence scoring | Analysts need to trust the narrative | Low | Based on event coverage percentage |
| Raw event viewer | Ability to drill down into details | Low | Standard table view |
| Incident list view | Dashboard to browse all incidents | Low | Filterable by severity/date |
| Analyst feedback | Thumbs up/down + notes for training | Low | Simple UI component |
| Alert ingestion | Pull from Graylog REST API | Medium | Batch processing for v1 |

## Differentiators

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Prompt injection prevention | Security-focused, builds trust | Medium | Multi-layer defense |
| Real-time narrative streaming | Better UX for long narratives | Low | Claude API streaming |
| Event correlation engine | Groups related events automatically | High | IP/time window clustering |
| Narrative templates | Customizable output format | Low | YAML/JSON config |
| Export capabilities | Share narratives with team | Low | PDF/Markdown export |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Real-time streaming | Over-engineering for v1 | Batch processing sufficient |
| Multi-SIEM support | Graylog only for demo | Architecture supports future |
| User authentication | Single analyst view for demo | Environment variables for config |
| Mobile app | Web-only for hackathon | Responsive design only |
| Custom ML models | Too complex, use Claude API | Leverage existing AI models |
| Complex role-based access | Not needed for demo | Simple admin/user views |
| Custom alert rules | Out of scope | Use Graylog's built-in rules |

## Feature Dependencies

```
Alert Ingestion → Event Aggregation → Incident Clustering → Narrative Generation → Source Tracing
                            ↓
                    MITRE ATT&CK Mapping → Confidence Scoring
                            ↓
                    Story Card Dashboard → Analyst Feedback
```

## MVP Recommendation

**Prioritize (Week 1-2):**
1. Alert ingestion from Graylog REST API
2. Event aggregation and incident clustering
3. MITRE ATT&CK mapping via STIX data
4. Claude API integration with prompt injection prevention
5. Source tracing implementation

**Defer (Week 3):**
- Analyst feedback system (nice-to-have)
- Export capabilities (can be added later)
- Narrative templates (can be hardcoded initially)

**Never Build:**
- Real-time streaming (batch sufficient)
- Multi-SIEM support (Graylog only)
- User authentication (single view)
- Mobile app (web-only)

## Data Model Requirements

### Core Entities

| Entity | Purpose | Key Fields |
|--------|---------|------------|
| RawEvent | Individual alert from Graylog | id, timestamp, source_ip, event_type, raw_message |
| Incident | Grouped related events | id, title, severity, start_time, end_time, narrative |
| NarrativeSentence | Individual claim with source | id, incident_id, text, confidence, source_event_ids |
| MitreMapping | ATT&CK technique mapping | id, incident_id, technique_id, technique_name |
| AnalystFeedback | User annotations | id, incident_id, rating, notes, created_at |

### Relationships

```
Incident 1:N RawEvent (incident_events)
Incident 1:N NarrativeSentence (incident_sentences)
Incident 1:N MitreMapping (incident_mitre)
Incident 1:N AnalystFeedback (incident_feedback)
```

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Narrative accuracy | >90% | Analyst feedback ratings |
| Source traceability | 100% | Every sentence links to events |
| Processing time | <30s per incident | API response time |
| Prompt injection resistance | 100% | Red-team testing |
| MITRE ATT&CK coverage | >80% | Techniques mapped correctly |

---

*Last Updated: 2026-06-08*
*Research Mode: Ecosystem*
*Confidence: HIGH*
