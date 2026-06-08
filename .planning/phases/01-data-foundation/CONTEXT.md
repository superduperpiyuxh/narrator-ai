# Context: Phase 1 — Data Foundation

## Current State

**Graylog:**
- Running: Graylog 7.1.2 on localhost:9000 (admin/admin)
- GELF HTTP input: port 12201 (running, no events yet)
- Streams: Default Stream, All events, All system events
- OpenSearch: 2.11.0 (no events indexed yet)
- Mongo: 7.0 (metadata only)

**Sample Data:**
- `data/sample_json_20260301/20251221.json` — 323,182 events (Dec 21, 2025)
- `data/sample_json_20260301/20251222.json` — 330,487 events (Dec 22, 2025)
- Total: ~653K events ready for import
- Import script exists: `scripts/import_to_graylog.py` (GELF HTTP, one event per request)

**Backend:**
- Go 1.26.3 + Gin 1.12.0 initialized at `backend/`
- Only has health endpoint (`GET /health`)
- SQLite driver (`mattn/go-sqlite3`) already in go.mod
- Module: `github.com/superduperpiyuxh/narrator-ai/backend`

**Frontend:**
- Next.js + TypeScript initialized at `frontend/`

## Event Field Schema (from sample data)

Core fields present in every event:

| Field | Type | Example | Notes |
|-------|------|---------|-------|
| `timestamp` | string | `"2025-12-21 09:19:00"` | Format: `YYYY-MM-DD HH:MM:SS` |
| `hostname` | string | `"WS-SAL-0013"` | Source machine |
| `event_type` | string | `"process_start"` | Event category |
| `event_id` | string | `"4688"` | Windows Event ID |
| `user` | string | `"svc_siem_collector"` | Acting user |
| `account` | string | `"svc_siem_collector"` | Often same as user |
| `source_ip` | string | `"10.1.0.197"` | Source IP |
| `destination_ip` | string | `"10.1.50.8"` | Destination IP (optional) |
| `process_name` | string | `"sysmon.exe"` | Process name |
| `command_line` | string | `"Process creation event logged"` | Command/description |
| `parent_process` | string | `"services.exe"` | Parent process |
| `success` | string | `"true"` | Boolean as string |
| `log_type` | string | `"windows_security_event"` | Log source type |
| `session_id` | string | `"svc_siem_collector_2025-12-21"` | Session identifier |
| `department` | string | `"Sales"` | Org context |
| `location` | string | `"NYC_HQ"` | Physical location |
| `device_type` | string | `"workstation"` | Device role |
| `file_path` | string | `"C:\\Windows\\System32\\"` | Optional |
| `port` | string | `"443"` | Optional |
| `protocol` | string | `"HTTPS"` | Optional |
| `error` | string | `"IP not whitelisted"` | Optional |
| `severity` | string | `"low"` | Optional |

Event types observed: `process_start`, `system_event`, `network_connection`, `login_attempt`, `login_success`, `admin_action`, `file_access`, `web_browsing`, `kerberos_auth_success`, `api_key_auth_failure`, `api_key_auth_success`, `certificate_auth_success`, `oauth_token_success`, and more.

## GELF HTTP API (for Go backend)

**Endpoint:** `POST http://localhost:12201/gelf`
**Content-Type:** `application/x-gelf`

GELF message format:
```json
{
  "version": "1.1",
  "host": "WS-SAL-0013",
  "short_message": "process_start: Process creation event logged",
  "timestamp": 1734758340.0,
  "_event_type": "process_start",
  "_user": "svc_siem_collector",
  "_source_ip": "10.1.0.197",
  "_event_id": "4688"
}
```

Fields prefixed with `_` are custom fields. Graylog stores them in OpenSearch.

**Graylog REST API endpoints:**
- `GET /api/system/inputs` — list inputs
- `GET /api/streams` — list streams
- `GET /api/search/universal/relative?query=*&range=86400&limit=10&fields=timestamp,source` — query events
- `POST /api/search/universal/relative` — query events (POST)

## Phase 1 Requirements Coverage

| Requirement | Description | Plan |
|-------------|-------------|------|
| ING-01 | Pull security logs from Graylog via REST API | 01-1 |
| ING-02 | Store raw events in SQLite with full metadata | 01-2 |
| ING-03 | Parse and normalize event fields | 01-3 |

## Open Questions for Discussion

1. **Data Import Strategy:** Should we import the ~653K events into Graylog first, or have the Go backend pull from Graylog REST API directly? The import scripts exist but are slow (one event per request).

2. **SQLite vs Ingesting from Graylog:** The ROADMAP says "pull from Graylog via REST API" — but for demo purposes, should we also support direct JSON import as a fallback?

3. **Event Count Target:** The sample data has 653K events. For demo purposes, should we import a subset (e.g., 10K-50K) for faster queries during the hackathon?

4. **Normalization Scope:** ING-03 says "normalize event fields" — should we:
   - Map `event_type` values to a canonical enum?
   - Normalize timestamp formats?
   - Map `success` string to boolean?
   - All of the above?

5. **Graylog REST API vs GELF:** The import script uses GELF HTTP (port 12201) for ingestion. For pulling events, we need the REST API (port 9000). Both are needed.

6. **Schema Design:** Should we have:
   - A single `events` table with all fields?
   - Separate tables for different event types?
   - Reference tables for hosts, users, IPs?

## Decisions Made

- **Event Count:** Keep 242K events currently in Graylog (no reimport needed)
- **Schema Design:** Hybrid approach — single `events` table with indexed lookup columns for hosts/users/IPs
- **Normalization:** Full normalization — parse timestamps, normalize event_type enum, cast success to boolean
- Graylog REST API for pulling events (ING-01)
- SQLite for local storage (ING-02)
- Go + Gin backend processes events
- Events stored with full metadata for source tracing in Phase 3

## Risks

- **Graylog REST API pagination:** Querying 653K events will require paginating through results. Need to handle `search_after` token.
- **Event field variability:** Not all events have all fields. Schema must handle nullable columns.
- **Import speed:** GELF HTTP one-at-a-time is slow for 653K events. Consider bulk import or direct SQLite seeding for demo.
