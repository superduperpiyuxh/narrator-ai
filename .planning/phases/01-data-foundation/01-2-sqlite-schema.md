# Plan 01-2: SQLite Schema Design

## Goal
Design and create SQLite database schema for storing security events with indexed lookup columns.

## Depends On
01-1 (Graylog API integration)

## Requirements
ING-02: Store raw events in SQLite with full metadata

## Technical Details

### Schema Design (Hybrid Approach)
Single `events` table with all fields + indexed columns for common queries:

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Core fields (always present)
    timestamp TEXT NOT NULL,           -- ISO 8601 format
    hostname TEXT NOT NULL,
    event_type TEXT NOT NULL,
    event_id TEXT,
    user_name TEXT,
    source_ip TEXT,
    dest_ip TEXT,
    -- Process fields
    process_name TEXT,
    command_line TEXT,
    parent_process TEXT,
    -- Metadata
    log_type TEXT,
    session_id TEXT,
    department TEXT,
    location TEXT,
    device_type TEXT,
    success BOOLEAN DEFAULT 0,
    -- Network fields
    port TEXT,
    protocol TEXT,
    file_path TEXT,
    -- Security fields
    severity TEXT,
    error TEXT,
    -- Raw data (for source tracing in Phase 3)
    raw_json TEXT,
    -- Timestamps
    created_at TEXT DEFAULT (datetime('now')),
    -- Indexes
    UNIQUE(hostname, timestamp, event_type)  -- Prevent duplicates
);

-- Performance indexes
CREATE INDEX idx_events_timestamp ON events(timestamp);
CREATE INDEX idx_events_hostname ON events(hostname);
CREATE INDEX idx_events_event_type ON events(event_type);
CREATE INDEX idx_events_source_ip ON events(source_ip);
CREATE INDEX idx_events_user ON events(user_name);
CREATE INDEX idx_events_success ON events(success);
```

### Implementation
1. Create `backend/internal/database/sqlite.go` — database connection and migrations
2. Create `backend/internal/database/queries.go` — CRUD operations
3. Create `backend/internal/database/models.go` — Go structs matching schema
4. Add auto-migration on server startup
5. Implement `EventStore` interface:
   - `InsertEvents(events []Event) error`
   - `GetEvents(limit, offset int) ([]Event, int, error)`
   - `GetEventsByHost(hostname string) ([]Event, error)`
   - `GetEventsByType(eventType string) ([]Event, error)`
   - `SearchEvents(query string) ([]Event, error)`

## Verification
- SQLite database file created at `backend/narratorai.db`
- `SELECT COUNT(*) FROM events` returns 0 initially
- After import: returns correct count
- Queries by hostname, event_type, timestamp work with indexes
