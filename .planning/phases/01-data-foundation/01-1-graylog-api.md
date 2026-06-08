# Plan 01-1: Graylog REST API Integration

## Goal
Go backend can query Graylog REST API and retrieve security events with pagination support.

## Depends On
Nothing (first plan)

## Requirements
ING-01: Pull security logs from Graylog via REST API

## Technical Details

### Graylog REST API
- Base URL: `http://localhost:9000`
- Auth: Basic (admin/admin)
- Query endpoint: `GET /api/search/universal/relative`
- Params: `query=*`, `range=86400` (last 24h), `limit=500`, `fields=timestamp,source,event_type,...`
- Pagination: Use `search_after` token from response for next page

### Implementation
1. Create `backend/internal/graylog/client.go` — HTTP client for Graylog API
2. Create `backend/internal/graylog/types.go` — response/request types
3. Implement `FetchEvents(query string, limit int) ([]Event, error)`
4. Implement pagination with `search_after` token
5. Add `GET /api/events` endpoint in `backend/cmd/server/main.go`
6. Return events as JSON to frontend

### Event Mapping
Map Graylog fields to Go struct:
```
Graylog field → Go field
timestamp → Timestamp (time.Time)
source → Hostname (string)
_event_type → EventType (string)
_user → User (string)
_source_ip → SourceIP (string)
_command_line → CommandLine (string)
```

## Verification
- `curl http://localhost:8080/api/events` returns JSON array of events
- Response includes at least 10 events with all fields populated
- Pagination works: first page returns 500, second page returns next 500
