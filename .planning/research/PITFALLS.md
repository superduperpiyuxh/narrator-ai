# Domain Pitfalls

**Domain:** Security Incident Narrative Generator
**Research Date:** 2026-06-08
**Research Mode:** Ecosystem

## Executive Summary

This document catalogs critical, moderate, and minor pitfalls for building a security incident narrative generator. The most dangerous pitfalls involve prompt injection vulnerabilities, hallucination in LLM outputs, and source tracing failures. Each pitfall includes prevention strategies and detection methods.

## Critical Pitfalls

Mistakes that cause rewrites or major issues.

### Pitfall 1: Prompt Injection Vulnerability

**What goes wrong:** Attackers embed malicious instructions in security logs that hijack the LLM, causing it to ignore safety guidelines or generate harmful content.

**Why it happens:** Security logs often contain user-controlled fields (IP addresses, hostnames, usernames) that can contain injection payloads. The LLM cannot distinguish between legitimate log data and adversarial input.

**Consequences:**
- LLM generates false narratives about incidents
- Sensitive data leakage through manipulated outputs
- Loss of trust in the system
- Potential legal liability if used for compliance

**Prevention:**
1. **Multi-layer defense** (OWASP #1 LLM vulnerability):
   - Layer 1: Input validation with regex patterns
   - Layer 2: XML wrapping for untrusted content
   - Layer 3: Claude Haiku classifier for secondary detection
   - Layer 4: Output validation with JSON schema
2. **Never trust user-controlled fields** - treat all log data as untrusted
3. **Implement rate limiting** to prevent abuse

**Detection:**
- Monitor for unusual LLM output patterns
- Log all injection detection attempts
- Regular red-team testing with known attack vectors

### Pitfall 2: LLM Hallucination in Narratives

**What goes wrong:** Claude generates plausible-sounding but factually incorrect narratives about security incidents, including events that never happened.

**Why it happens:** LLMs are designed to generate coherent text, not to verify facts against source data. Without proper grounding, the model will fill in gaps with plausible but false information.

**Consequences:**
- Analysts make wrong decisions based on false narratives
- Loss of trust in the system
- Potential security incidents if false positives are acted upon

**Prevention:**
1. **Source tracing** - every sentence must link to specific event IDs
2. **Confidence scoring** - calculate based on event coverage percentage
3. **Structured output** - use JSON schema to constrain responses
4. **Temperature setting** - use low temperature (0.1-0.3) for factual accuracy

**Detection:**
- Analyst feedback system (thumbs up/down)
- Automated consistency checks against source events
- Regular audits of generated narratives

### Pitfall 3: Source Tracing Failure

**What goes wrong:** Narrative sentences cannot be traced back to specific raw events, making the system useless for audit or compliance purposes.

**Why it happens:** The LLM generates text without explicit links to source data, or the linking mechanism is not implemented correctly.

**Consequences:**
- System loses its key differentiator from Splunk AI
- Cannot verify narrative accuracy
- Fails compliance requirements

**Prevention:**
1. **Explicit source tracking** - require LLM to cite event IDs in output
2. **Database schema design** - store source_event_ids with each sentence
3. **Validation layer** - verify all cited events exist in database

**Detection:**
- Check that every sentence has at least one source event
- Monitor for sentences with missing or invalid source references

## Moderate Pitfalls

Issues that cause delays or require workarounds.

### Pitfall 1: Graylog API Compatibility

**What goes wrong:** Graylog REST API client libraries are outdated (v2.4), causing integration issues with newer Graylog versions.

**Why it happens:** Open-source libraries often lag behind commercial products. Graylog's API may have changed significantly since the libraries were last updated.

**Consequences:**
- Integration delays while debugging API differences
- May need to implement custom HTTP client
- Missing features from newer API versions

**Prevention:**
1. **Test early** - verify API compatibility with actual Graylog instance
2. **Use standard HTTP client** - avoid outdated libraries
3. **Document API differences** - keep notes on what works/doesn't

**Detection:**
- Integration tests against real Graylog instance
- Monitor for API errors in logs

### Pitfall 2: SQLite Performance at Scale

**What goes wrong:** SQLite becomes slow with large numbers of events or concurrent access patterns.

**Why it happens:** SQLite is designed for single-writer scenarios and can struggle with concurrent reads/writes.

**Consequences:**
- Slow narrative generation for large incidents
- UI becomes unresponsive
- Demo may fail under load

**Prevention:**
1. **WAL mode** - enable Write-Ahead Logging for better concurrency
2. **Busy timeout** - configure proper timeout for concurrent access
3. **Batch operations** - minimize individual database calls
4. **Index optimization** - add indexes for common query patterns

**Detection:**
- Monitor query execution times
- Test with realistic data volumes

### Pitfall 3: Event Aggregation Logic

**What goes wrong:** Events are incorrectly grouped into incidents, causing unrelated events to be mixed or related events to be separated.

**Why it happens:** Simple time-window/IP-based grouping may not capture complex attack patterns.

**Consequences:**
- Narrative includes irrelevant events
- Important events are missed from the narrative
- MITRE ATT&CK mapping becomes inaccurate

**Prevention:**
1. **Multiple grouping criteria** - combine time, IP, event type
2. **Configurable thresholds** - allow tuning of grouping parameters
3. **Manual review option** - let analysts adjust groupings

**Detection:**
- Analyst feedback on grouping accuracy
- Automated metrics on grouping quality

## Minor Pitfalls

Issues that cause frustration but not project failure.

### Pitfall 1: Claude API Rate Limits

**What goes wrong:** Claude API returns rate limit errors during heavy processing.

**Why it happens:** API has usage limits that can be exceeded during batch processing.

**Consequences:**
- Delayed narrative generation
- Need to implement retry logic
- May need to optimize API usage

**Prevention:**
1. **Implement exponential backoff** - retry with increasing delays
2. **Batch processing** - process incidents sequentially, not all at once
3. **Monitor usage** - track API calls and costs

**Detection:**
- Log API error responses
- Monitor processing queue length

### Pitfall 2: STIX Data Freshness

**What goes wrong:** MITRE ATT&CK data becomes outdated, causing incorrect technique mappings.

**Why it happens:** ATT&CK framework is updated regularly, but local data may not be refreshed.

**Consequences:**
- Incorrect technique mappings
- Missing new attack patterns
- Outdated threat intelligence

**Prevention:**
1. **Regular updates** - fetch ATT&CK data periodically
2. **Version tracking** - store ATT&CK version used
3. **Update mechanism** - implement refresh capability

**Detection:**
- Compare local data version with latest available
- Monitor for missing techniques

### Pitfall 3: Frontend State Management

**What goes wrong:** React state becomes inconsistent, causing UI bugs or data loss.

**Why it happens:** Complex state management with multiple data sources (API, local state, user input).

**Consequences:**
- UI shows stale data
- User actions are lost
- Poor user experience

**Prevention:**
1. **Use React Query** - handle server state properly
2. **Optimistic updates** - update UI immediately, sync with server
3. **Error boundaries** - handle and display errors gracefully

**Detection:**
- React DevTools state inspection
- User-reported issues

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Data Ingestion | Graylog API incompatibility | Test early, use HTTP client |
| Event Aggregation | Incorrect grouping logic | Multiple criteria, manual review |
| Narrative Generation | LLM hallucination | Source tracing, confidence scoring |
| Security | Prompt injection | Multi-layer defense |
| Frontend | State management issues | React Query, error boundaries |
| Demo | Performance under load | SQLite optimization, testing |

## Security-Specific Pitfalls

### Pitfall: Credential Exposure

**What goes wrong:** API keys or database credentials are exposed in code or logs.

**Why it happens:** Hardcoded credentials, logging sensitive data, or improper environment variable handling.

**Consequences:**
- Unauthorized access to Claude API
- Database compromise
- Security incident

**Prevention:**
1. **Environment variables only** - never hardcode credentials
2. **Log sanitization** - filter sensitive data from logs
3. **Git ignore** - ensure .env files are not committed

**Detection:**
- Code scanning for credentials
- Log review for sensitive data

### Pitfall: SQL Injection

**What goes wrong:** User input is used in SQL queries without proper parameterization.

**Why it happens:** Using string concatenation instead of parameterized queries.

**Consequences:**
- Database compromise
- Data leakage
- System takeover

**Prevention:**
1. **Use sqlc** - generates parameterized queries automatically
2. **Never concatenate** - always use query parameters
3. **Input validation** - validate all user input

**Detection:**
- SQL injection testing tools
- Code review for raw SQL

## Risk Assessment Summary

| Risk Level | Count | Examples |
|------------|-------|----------|
| Critical | 3 | Prompt injection, hallucination, source tracing |
| Moderate | 3 | API compatibility, SQLite performance, aggregation |
| Minor | 3 | Rate limits, data freshness, state management |
| Security | 2 | Credential exposure, SQL injection |

## Recommendations

### Immediate Actions (Week 1)
1. Implement prompt injection prevention layers
2. Set up source tracing mechanism
3. Test Graylog API integration
4. Configure SQLite for performance

### Short-term Actions (Week 2)
1. Implement confidence scoring
2. Set up analyst feedback system
3. Add input validation
4. Create red-team test suite

### Ongoing Actions (Week 3)
1. Monitor for hallucination patterns
2. Update MITRE ATT&CK data
3. Performance testing
4. Security audit

---

*Last Updated: 2026-06-08*
*Research Mode: Ecosystem*
*Confidence: HIGH*
