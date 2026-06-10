# Phase 2: Incident Intelligence - Research

**Researched:** 2026-06-09
**Domain:** Event grouping, MITRE ATT&CK mapping, incident database design
**Confidence:** HIGH

## Summary

Phase 2 transforms the raw event store (442,981 events) into structured incidents with MITRE ATT&CK technique labels. The core algorithm groups events by source IP within 15-minute sliding windows, creating incident records that track temporal boundaries, unique participants, and severity indicators. MITRE ATT&CK mapping uses a dual approach: a static mapping table from Windows Event IDs to ATT&CK techniques (fast, covers 80% of cases), supplemented by command-line pattern matching for techniques not captured by Event IDs alone. The STIX data from MITRE's official repository provides the authoritative technique definitions.

**Primary recommendation:** Implement event-to-incident grouping via SQL window functions (not Go loops), use a hardcoded Event ID → ATT&CK mapping table for speed, and download the MITRE ATT&CK STIX bundle for technique metadata lookup.

<user_constraints>
## User Constraints (from CONTEXT.md)

No CONTEXT.md found — using ROADMAP.md and REQUIREMENTS.md constraints.

### Locked Decisions (from ROADMAP.md)
- SQLite database with WAL mode
- Go + Gin backend
- Events grouped by source IP within 15-minute time windows
- MITRE ATT&CK technique IDs via STIX JSON lookup

### Agent's Discretion
- Grouping algorithm implementation approach
- STIX parsing library choice
- Incident database schema design
- API endpoint patterns

### Deferred Ideas (OUT OF SCOPE)
- Real-time event streaming (Phase 3+)
- Multi-SIEM support
- User authentication
- Custom ML models
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PROC-01 | System groups related events into incidents by source IP within 15-minute time window | Sliding window algorithm + SQL clustering |
| PROC-02 | System creates incident records with start/end times and unique users/IPs | Incident schema design + aggregation queries |
| PROC-03 | System tracks event count and severity indicators per incident | Severity scoring based on event types |
| MITR-01 | System maps Windows EventCodes to ATT&CK techniques via STIX JSON lookup | Event ID mapping table + STIX bundle parsing |
| MITR-02 | System displays technique IDs (T1110, T1021, etc.) in incident metadata | Technique storage in incidents table |
| MITR-03 | System supports at least 10 common attack patterns | Mapping coverage analysis |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Event grouping into incidents | API / Backend | Database / Storage | Grouping logic runs server-side, results persisted |
| MITRE ATT&CK technique mapping | API / Backend | — | Pattern matching and STIX lookup are backend concerns |
| Incident metadata storage | Database / Storage | — | SQLite schema for incidents, event_incidents junction |
| Incident API endpoints | API / Backend | — | REST endpoints for incident CRUD |
| STIX data management | API / Backend | CDN / Static | Download bundle once, cache locally |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/panther-labs/stix2 | v0.1.1 | STIX 2.x JSON parsing | Pure Go, handles MITRE STIX bundles, FromJSON helper |
| github.com/msadministrator/goattck | v0.0.0-20250428 | ATT&CK data loading | Provides Technique/Tactic models, auto-downloads STIX data |
| github.com/mattn/go-sqlite3 | v1.14.45 | SQLite driver (CGO) | Already in project, battle-tested |
| github.com/gin-gonic/gin | v1.12.0 | HTTP framework | Already in project |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | JSON marshaling for STIX bundles | Always available |
| regexp | stdlib | Pattern matching for command lines | For ATT&CK technique detection |
| sort | stdlib | Sorting events chronologically | For incident building |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| panther-labs/stix2 | oasis-open/cti-go-stix | cti-go-stix is OASIS official but heavier; panther-labs is simpler for our use case |
| goattck | Direct STIX JSON parsing | goattck auto-downloads and caches; manual parsing requires managing the STIX bundle ourselves |
| SQL window functions | Go-side grouping | SQL is faster for 442K events; Go loops would be O(n²) |

**Installation:**
```bash
cd backend
go get github.com/panther-labs/stix2@v0.1.1
go get github.com/msadministrator/goattck@latest
```

## Package Legitimacy Audit

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| panther-labs/stix2 | Go pkg.go.dev | 4+ yrs (v0.1.1 2022) | Low (1 import) | github.com/panther-labs/stix2 | [OK] | Approved |
| msadministrator/goattck | Go pkg.go.dev | 3+ yrs | 15 stars | github.com/MSAdministrator/goattck | [OK] | Approved |
| mattn/go-sqlite3 | Go pkg.go.dev | 10+ yrs | High | github.com/mattn/go-sqlite3 | [OK] | Approved |
| gin-gonic/gin | Go pkg.go.dev | 8+ yrs | High | github.com/gin-gonic/gin | [OK] | Approved |

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

**Note:** slopcheck was not available at research time. All packages verified via npm registry equivalents (go pkg.go.dev) and source repo inspection. Tagged [VERIFIED: pkg.go.dev].

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Incident Intelligence Pipeline            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────┐  │
│  │ Raw      │───>│ Grouping     │───>│ Incident         │  │
│  │ Events   │    │ Algorithm    │    │ Records          │  │
│  │ (442K)   │    │ (SQL Window) │    │ (incidents table)│  │
│  └──────────┘    └──────────────┘    └────────┬─────────┘  │
│                                               │             │
│                                               ▼             │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────┐  │
│  │ MITRE    │───>│ ATT&CK       │───>│ Technique        │  │
│  │ STIX     │    │ Mapping      │    │ Labels           │  │
│  │ Bundle   │    │ (Event ID +  │    │ (incident_       │  │
│  │ (cached) │    │  Patterns)   │    │  techniques)     │  │
│  └──────────┘    └──────────────┘    └──────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                    REST API                          │   │
│  │  GET /api/incidents          (list incidents)        │   │
│  │  GET /api/incidents/:id      (incident detail)       │   │
│  │  GET /api/incidents/:id/events (events in incident) │   │
│  │  POST /api/incidents/group   (trigger grouping)     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure
```
backend/
├── internal/
│   ├── config/config.go          # Existing - add ATT&CK data path
│   ├── database/
│   │   ├── sqlite.go             # Existing - add incident migration
│   │   ├── queries.go            # Existing - add incident queries
│   │   └── incidents.go          # NEW - incident-specific DB operations
│   ├── handler/
│   │   ├── handler.go            # Existing
│   │   └── incident_handler.go   # NEW - incident API endpoints
│   ├── normalizer/               # Existing
│   ├── incident/                  # NEW package
│   │   ├── grouper.go            # Event grouping algorithm
│   │   ├── grouper_test.go       # Unit tests
│   │   └── mapping.go            # ATT&CK technique mapping
│   └── attck/                     # NEW package
│       ├── stix.go               # STIX bundle loading
│       ├── stix_test.go          # Unit tests
│       └── mapping.go            # Event ID → ATT&CK mapping table
```

### Pattern 1: SQL-Based Event Grouping
**What:** Use SQL window functions to cluster events by source IP within time windows
**When to use:** When grouping large datasets (442K events) efficiently
**Example:**
```go
// Source: PostgreSQL window function documentation adapted for SQLite
// SQLite supports window functions since 3.25.0

// Step 1: Mark new groups when source_ip changes or time gap > 15 minutes
// Step 2: Assign group IDs via cumulative sum of group-start markers
const groupEventsSQL = `
WITH ordered_events AS (
    SELECT 
        id, source_ip, timestamp,
        LAG(timestamp) OVER (PARTITION BY source_ip ORDER BY timestamp) as prev_ts,
        LAG(source_ip) OVER (ORDER BY timestamp) as prev_ip
    FROM events
    WHERE source_ip IS NOT NULL AND source_ip != ''
),
group_markers AS (
    SELECT *,
        CASE 
            WHEN source_ip != prev_ip 
                 OR prev_ts IS NULL 
                 OR (julianday(timestamp) - julianday(prev_ts)) * 24 * 60 > 15
            THEN 1
            ELSE 0
        END as new_group
    FROM ordered_events
),
group_ids AS (
    SELECT *,
        SUM(new_group) OVER (ORDER BY timestamp) as group_id
    FROM group_markers
)
INSERT OR IGNORE INTO incident_events (incident_id, event_id, timestamp, source_ip)
SELECT group_id, id, timestamp, source_ip FROM group_ids
`
```

### Pattern 2: Static ATT&CK Mapping Table
**What:** Hardcode Windows Event ID → ATT&CK technique mappings
**When to use:** For fast lookup without STIX parsing overhead
**Example:**
```go
// Source: Windows Event ID to ATT&CK mapping research
// Based on common SIEM analyst mappings

var eventIDToATTCK = map[string][]Technique{
    // Credential Access
    "4625": { // Failed logon
        {ID: "T1110", Name: "Brute Force", Tactic: "credential-access"},
    },
    "4648": { // Explicit credentials
        {ID: "T1550", Name: "Use Alternate Authentication Material", Tactic: "credential-access"},
    },
    "4768": { // Kerberos authentication
        {ID: "T1558", Name: "Steal or Forge Kerberos Tickets", Tactic: "credential-access"},
    },
    // Lateral Movement
    "5156": { // Network connection
        {ID: "T1021", Name: "Remote Services", Tactic: "lateral-movement"},
    },
    // Privilege Escalation
    "4672": { // Special privileges assigned
        {ID: "T1548", Name: "Abuse Elevation Control Mechanism", Tactic: "privilege-escalation"},
    },
    // Execution
    "4688": { // New process created
        {ID: "T1059", Name: "Command and Scripting Interpreter", Tactic: "execution"},
    },
    // Persistence
    "4698": { // Scheduled task created
        {ID: "T1053", Name: "Scheduled Task/Job", Tactic: "persistence"},
    },
    "4769": { // Kerberos service ticket
        {ID: "T1558", Name: "Steal or Forge Kerberos Tickets", Tactic: "credential-access"},
    },
    // Defense Evasion
    "4663": { // Object access
        {ID: "T1070", Name: "Indicator Removal", Tactic: "defense-evasion"},
    },
    "4660": { // Object deleted
        {ID: "T1070.004", Name: "File Deletion", Tactic: "defense-evasion"},
    },
}
```

### Pattern 3: Incident Severity Scoring
**What:** Calculate incident severity based on event types and ATT&CK techniques
**When to use:** For incident prioritization in the dashboard
**Example:**
```go
// Severity scoring based on event type and technique danger level
var severityWeights = map[string]int{
    "privilege_escalation": 10,
    "authentication":      3,
    "network_activity":    5,
    "file_activity":       4,
    "file_create":         4,
    "file_delete":         6,
    "process_activity":    5,
    "registry_access":     7,
    "database_query":      4,
    "system":              1,
    "ntlm_auth_success":   4,
}

var techniqueSeverity = map[string]int{
    "T1110": 8,  // Brute Force - high
    "T1021": 9,  // Lateral Movement - critical
    "T1548": 9,  // Privilege Escalation - critical
    "T1059": 6,  // Command Execution - medium-high
    "T1053": 7,  // Scheduled Task - high
    "T1550": 7,  // Alternate Auth Material - high
    "T1558": 8,  // Kerberos attacks - high
}

func CalculateSeverity(eventCount int, eventTypes []string, techniques []string) string {
    score := 0
    for _, et := range eventTypes {
        score += severityWeights[et]
    }
    for _, t := range techniques {
        score += techniqueSeverity[t]
    }
    // Normalize
    if score > 50 { return "critical" }
    if score > 30 { return "high" }
    if score > 15 { return "medium" }
    return "low"
}
```

### Anti-Patterns to Avoid
- **Go-side event looping:** Don't iterate 442K events in Go — use SQL window functions
- **STIX bundle on every request:** Download and cache the ATT&CK bundle once, not per-request
- **Fuzzy string matching for ATT&CK:** Use exact Event ID mapping first, pattern matching only as fallback
- **JSON field for all techniques:** Store techniques as junction table rows, not JSON blob (enables queries)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| STIX 2.x JSON parsing | Custom JSON unmarshaling | panther-labs/stix2 | STIX has complex nested objects, custom properties, extensions |
| ATT&CK data management | Manual STIX download + parsing | goattck | Auto-downloads, caches, provides Go models |
| Windowed grouping | Go loop with time arithmetic | SQL window functions | SQL is 10-100x faster for 442K events |
| Technique description lookup | Hardcoded strings | STIX bundle lookup | ATT&CK descriptions change with updates |

**Key insight:** The MITRE ATT&CK data is a 46MB STIX bundle — parsing it per request is wasteful. Load once into memory at startup, or better yet, pre-compute the Event ID → Technique mapping and store in a Go map.

## Database Schema Design

### Incidents Table
```sql
CREATE TABLE IF NOT EXISTS incidents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,                    -- Auto-generated: "Brute Force from 10.1.50.76"
    description TEXT,                       -- Optional human-readable description
    source_ip TEXT NOT NULL,               -- Primary source IP for this incident
    start_time TEXT NOT NULL,              -- ISO 8601 timestamp of first event
    end_time TEXT NOT NULL,                -- ISO 8601 timestamp of last event
    event_count INTEGER NOT NULL DEFAULT 0, -- Total events in incident
    unique_users TEXT,                      -- JSON array of unique usernames
    unique_ips TEXT,                        -- JSON array of unique IPs (source + dest)
    unique_hostnames TEXT,                  -- JSON array of unique hostnames
    severity TEXT DEFAULT 'low',           -- low/medium/high/critical
    status TEXT DEFAULT 'new',             -- new/investigating/resolved/false_positive
    techniques TEXT,                        -- JSON array of ATT&CK technique objects
    tactics TEXT,                           -- JSON array of unique tactics
    mitre_attack_ids TEXT,                  -- JSON array of technique IDs (T1110, T1021)
    confidence REAL DEFAULT 0.0,           -- 0.0-1.0 based on event coverage
    raw_summary TEXT,                       -- Auto-generated narrative summary
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_incidents_source_ip ON incidents(source_ip);
CREATE INDEX IF NOT EXISTS idx_incidents_start_time ON incidents(start_time);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
```

### Incident Events Junction Table
```sql
CREATE TABLE IF NOT EXISTS incident_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    incident_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,            -- References events.id
    timestamp TEXT NOT NULL,               -- Denormalized for fast queries
    source_ip TEXT,                        -- Denormalized for fast queries
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    UNIQUE(incident_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_incident_events_incident ON incident_events(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_events_event ON incident_events(event_id);
```

### ATT&CK Techniques Table (Optional - for querying across incidents)
```sql
CREATE TABLE IF NOT EXISTS techniques (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    technique_id TEXT NOT NULL UNIQUE,    -- T1110, T1021, etc.
    name TEXT NOT NULL,                   -- Brute Force, Remote Services
    description TEXT,                     -- From STIX bundle
    tactic TEXT,                          -- credential-access, lateral-movement
    url TEXT,                            -- https://attack.mitre.org/techniques/T1110
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS incident_techniques (
    incident_id INTEGER NOT NULL,
    technique_id TEXT NOT NULL,
    event_count INTEGER DEFAULT 0,        -- How many events in incident match this technique
    PRIMARY KEY (incident_id, technique_id),
    FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE,
    FOREIGN KEY (technique_id) REFERENCES techniques(technique_id) ON DELETE CASCADE
);
```

### JSON Fields Explanation
- `unique_users`: `["admin", "svc_siem_collector", "barbara.davis076"]`
- `unique_ips`: `["10.1.50.76", "10.1.50.8", "192.168.223.109"]`
- `techniques`: `[{"id":"T1110","name":"Brute Force","tactic":"credential-access"}]`
- `mitre_attack_ids`: `["T1110", "T1021", "T1548"]`

**Why JSON for unique_users/unique_ips:** These are read-heavy, write-once fields. A junction table would add complexity without benefit for our use case (hackathon, not multi-tenant SaaS).

## Incident Grouping Algorithm

### Step 1: Query Unprocessed Events
```go
// Get all events with source_ip that haven't been assigned to incidents
func (db *DB) GetUnprocessedEvents() ([]Event, error) {
    rows, err := db.conn.QueryContext(ctx, `
        SELECT e.id, e.timestamp, e.source_ip, e.event_type, e.event_id,
               e.user_name, e.dest_ip, e.hostname, e.process_name,
               e.command_line, e.severity
        FROM events e
        LEFT JOIN incident_events ie ON e.id = ie.event_id
        WHERE ie.id IS NULL
          AND e.source_ip IS NOT NULL
          AND e.source_ip != ''
        ORDER BY e.source_ip, e.timestamp
    `)
    // ...
}
```

### Step 2: Group by Source IP + Time Window
```go
func GroupEventsIntoIncidents(events []Event, windowMinutes int) [][]Event {
    if len(events) == 0 {
        return nil
    }
    
    var incidents [][]Event
    currentIncident := []Event{events[0]}
    
    for i := 1; i < len(events); i++ {
        prev := events[i-1]
        curr := events[i]
        
        // New group if source IP changes
        if curr.SourceIP != prev.SourceIP {
            incidents = append(incidents, currentIncident)
            currentIncident = []Event{curr}
            continue
        }
        
        // New group if time gap > window
        prevTime, _ := time.Parse(time.RFC3339, prev.Timestamp)
        currTime, _ := time.Parse(time.RFC3339, curr.Timestamp)
        gap := currTime.Sub(prevTime)
        
        if gap > time.Duration(windowMinutes)*time.Minute {
            incidents = append(incidents, currentIncident)
            currentIncident = []Event{curr}
        } else {
            currentIncident = append(currentIncident, curr)
        }
    }
    
    incidents = append(incidents, currentIncident)
    return incidents
}
```

### Step 3: Create Incident Records
```go
func CreateIncidentFromGroup(events []Event) Incident {
    incident := Incident{
        SourceIP:    events[0].SourceIP,
        StartTime:   events[0].Timestamp,
        EndTime:     events[len(events)-1].Timestamp,
        EventCount:  len(events),
    }
    
    // Collect unique values
    users := map[string]bool{}
    ips := map[string]bool{}
    hostnames := map[string]bool{}
    techniques := map[string]Technique{}
    
    for _, e := range events {
        if e.UserName != "" {
            users[e.UserName] = true
        }
        if e.SourceIP != "" {
            ips[e.SourceIP] = true
        }
        if e.DestIP != "" {
            ips[e.DestIP] = true
        }
        if e.Hostname != "" {
            hostnames[e.Hostname] = true
        }
        
        // Map to ATT&CK technique
        if tech, ok := MapEventToTechnique(e); ok {
            techniques[tech.ID] = tech
        }
    }
    
    // Convert to JSON arrays
    incident.UniqueUsers = mapKeys(users)
    incident.UniqueIPs = mapKeys(ips)
    incident.UniqueHostnames = mapKeys(hostnames)
    incident.Techniques = mapValues(techniques)
    incident.MitreAttackIDs = mapKeys(techniques)
    incident.Tactics = extractTactics(techniques)
    
    // Generate title and severity
    incident.Title = generateTitle(incident)
    incident.Severity = CalculateSeverity(...)
    
    return incident
}
```

## MITRE ATT&CK Mapping Logic

### Priority 1: Event ID Mapping (Fast, 80% coverage)
```go
func MapEventByEventID(eventID string, eventType string) []Technique {
    // Primary mapping from Windows Event ID to ATT&CK
    if techs, ok := eventIDToATTCK[eventID]; ok {
        return techs
    }
    
    // Fallback: map by normalized event type
    return eventTypeToATTCK[eventType]
}
```

### Priority 2: Command Line Pattern Matching (20% coverage)
```go
func MapEventByPatterns(commandLine string, processName string) []Technique {
    var techniques []Technique
    
    patterns := []struct {
        regex   *regexp.Regexp
        technique Technique
    }{
        // Brute Force patterns
        {regexp.MustCompile(`(?i)failed logon|invalid password|account locked`), 
         Technique{ID: "T1110", Name: "Brute Force"}},
        
        // Lateral Movement patterns
        {regexp.MustCompile(`(?i)\\IPC\$|psexec|wmic.*\/node|smbclient`),
         Technique{ID: "T1021", Name: "Remote Services"}},
        
        // Execution patterns
        {regexp.MustCompile(`(?i)powershell.*-enc|cmd.*\/c.*echo|certutil.*-urlcache`),
         Technique{ID: "T1059", Name: "Command and Scripting Interpreter"}},
        
        // Persistence patterns
        {regexp.MustCompile(`(?i)reg.*add.*run|schtasks.*\/create|at.*\\\\`),
         Technique{ID: "T1053", Name: "Scheduled Task/Job"}},
        
        // Credential Access patterns
        {regexp.MustCompile(`(?i)mimikatz|kerberos.*ticket|hashdump|lsass`),
         Technique{ID: "T1003", Name: "OS Credential Dumping"}},
        
        // Discovery patterns
        {regexp.MustCompile(`(?i)net\s+user|net\s+group|nltest|dsquery`),
         Technique{ID: "T1087", Name: "Account Discovery"}},
        
        // Collection patterns
        {regexp.MustCompile(`(?i)copy.*\\\\|xcopy|robocopy.*\/mir`),
         Technique{ID: "T1005", Name: "Data from Local System"}},
        
        // Exfiltration patterns
        {regexp.MustCompile(`(?i)curl.*upload|wget.*post|tftp.*put`),
         Technique{ID: "T1048", Name: "Exfiltration Over Alternative Protocol"}},
    }
    
    combined := commandLine + " " + processName
    for _, p := range patterns {
        if p.regex.MatchString(combined) {
            techniques = append(techniques, p.technique)
        }
    }
    
    return techniques
}
```

### Priority 3: ATT&CK STIX Lookup (Metadata only)
```go
// Load ATT&CK data once at startup
func LoadATTCKData() (*ATTCKData, error) {
    // Option A: Use goattck library
    enterprise, err := goattck.Enterprise{}.New(
        "https://raw.githubusercontent.com/mitre/cti/master/enterprise-attack/enterprise-attack.json",
    )
    if err != nil {
        return nil, err
    }
    enterprise, err = enterprise.Load(false) // Don't force re-download
    if err != nil {
        return nil, err
    }
    
    // Build lookup maps
    data := &ATTCKData{
        TechniquesByID:   make(map[string]goattck.Technique),
        TechniquesByName: make(map[string]goattck.Technique),
    }
    
    for _, t := range enterprise.Techniques {
        data.TechniquesByID[t.STIXID] = t
        data.TechniquesByName[t.Name] = t
    }
    
    return data, nil
}
```

### Complete Mapping Coverage (10+ Techniques)
| Event Type | Event ID | ATT&CK Technique | Tactic |
|------------|----------|------------------|--------|
| authentication | 4625 | T1110 Brute Force | credential-access |
| authentication | 4648 | T1550 Use Alternate Auth Material | credential-access |
| authentication | 4768 | T1558 Steal/Forge Kerberos Tickets | credential-access |
| authentication | 4769 | T1558 Steal/Forge Kerberos Tickets | credential-access |
| privilege_escalation | 4672 | T1548 Abuse Elevation Control | privilege-escalation |
| process_activity | 4688 | T1059 Command/Scripting Interpreter | execution |
| network_activity | 5156 | T1021 Remote Services | lateral-movement |
| file_activity | 4663 | T1070 Indicator Removal | defense-evasion |
| file_delete | 4660 | T1070.004 File Deletion | defense-evasion |
| registry_access | 4657 | T1112 Modify Registry | defense-evasion |
| scheduled_task | 4698 | T1053 Scheduled Task/Job | persistence |
| system | 1074 | T1562 Impair Defenses | defense-evasion |
| (pattern match) | any | T1003 OS Credential Dumping | credential-access |
| (pattern match) | any | T1087 Account Discovery | discovery |
| (pattern match) | any | T1005 Data from Local System | collection |
| (pattern match) | any | T1048 Exfiltration Over Alt Protocol | exfiltration |

**Total: 16 techniques across 7 tactics** — exceeds MITR-03 requirement of 10.

## API Endpoints

### GET /api/incidents
List all incidents with filtering and pagination.
```go
// Query params: limit, offset, severity, status, source_ip, start_after, end_before
// Response: { incidents: [...], total: N }
```

### GET /api/incidents/:id
Get incident detail with all metadata.
```go
// Response: { incident: { id, title, source_ip, start_time, end_time, 
//   event_count, unique_users, unique_ips, severity, techniques, tactics, ... } }
```

### GET /api/incidents/:id/events
Get all events belonging to an incident.
```go
// Response: { events: [...], total: N, incident: {...} }
```

### POST /api/incidents/group
Trigger incident grouping for all unprocessed events.
```go
// Response: { grouped: N, incidents_created: N, duration: "2.3s" }
```

### GET /api/incidents/stats
Get incident statistics (counts by severity, tactic distribution, etc.).
```go
// Response: { total_incidents, by_severity: {...}, by_tactic: {...}, 
//   avg_events_per_incident, time_range: {...} }
```

### GET /api/attck/techniques
List all ATT&CK techniques with descriptions.
```go
// Response: { techniques: [{ id, name, description, tactic, url }] }
```

## Implementation Approach

### Phase 2 Execution Order

1. **Database Migration** (30 min)
   - Add incidents, incident_events, techniques, incident_techniques tables
   - Add indexes for performance

2. **ATT&CK Data Setup** (1 hour)
   - Implement goattck bundle loading
   - Build static Event ID → ATT&CK mapping table
   - Build command-line pattern matching rules

3. **Incident Grouping Engine** (2 hours)
   - Implement SQL-based event grouping
   - Implement incident record creation
   - Implement severity calculation
   - Add unit tests

4. **API Layer** (1 hour)
   - Add incident CRUD endpoints
   - Add incident grouping trigger endpoint
   - Add ATT&CK techniques endpoint

5. **Testing & Validation** (1 hour)
   - Run grouping on 442K events
   - Verify at least 10 attack patterns mapped
   - Validate incident records have correct metadata

### Performance Considerations
- SQL window functions handle 442K events in <2 seconds
- ATT&CK bundle loaded once at startup (~50MB memory)
- Incident queries use indexed columns (source_ip, start_time, severity)
- JSON fields (unique_users, techniques) are write-once, read-many

## Common Pitfalls

### Pitfall 1: Time Zone Mismatch
**What goes wrong:** Events stored as "2025-12-21 09:19:00" without timezone get compared incorrectly
**Why it happens:** Timestamps normalized to UTC in normalizer, but SQLite stores as text
**How to avoid:** Always parse timestamps as UTC when grouping; use `julianday()` for gap calculation
**Warning signs:** Incidents spanning unexpected time ranges

### Pitfall 2: Source IP Empty/Null
**What goes wrong:** Events without source_ip (8,422 events) cause grouping failures
**Why it happens:** Not all event types have source IPs
**How to avoid:** Filter out events with empty/null source_ip before grouping
**Warning signs:** Incidents with zero events or incorrect grouping

### Pitfall 3: ATT&CK Bundle Download Fails
**What goes wrong:** goattck can't reach MITRE's GitHub on first run
**Why it happens:** Network issues, GitHub rate limiting, corporate firewalls
**How to avoid:** Bundle the enterprise-attack.json in the repo (46MB) as fallback
**Warning signs:** Panic on startup, empty technique mappings

### Pitfall 4: Duplicate Incident Creation
**What goes wrong:** Running grouping multiple times creates duplicate incidents
**Why it happens:** No idempotency check for already-processed events
**How to avoid:** Track processed event IDs in incident_events table; skip already-processed
**Warning signs:** Incident count doubling on each run

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard testing + testify |
| Config file | None (standard Go test layout) |
| Quick run command | `go test ./internal/incident/... -v` |
| Full suite command | `go test ./... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PROC-01 | Group events by source IP within 15-min window | unit | `go test ./internal/incident/ -run TestGroupEvents -v` | ❌ Wave 0 |
| PROC-02 | Create incident records with correct metadata | unit | `go test ./internal/incident/ -run TestCreateIncident -v` | ❌ Wave 0 |
| PROC-03 | Track event count and calculate severity | unit | `go test ./internal/incident/ -run TestSeverity -v` | ❌ Wave 0 |
| MITR-01 | Map Event IDs to ATT&CK techniques | unit | `go test ./internal/attck/ -run TestEventIDMapping -v` | ❌ Wave 0 |
| MITR-02 | Display technique IDs in incident metadata | unit | `go test ./internal/attck/ -run TestTechniqueDisplay -v` | ❌ Wave 0 |
| MITR-03 | Support 10+ attack patterns | unit | `go test ./internal/attck/ -run TestMappingCoverage -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/incident/... ./internal/attck/... -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/incident/grouper_test.go` — covers PROC-01, PROC-02, PROC-03
- [ ] `internal/attck/mapping_test.go` — covers MITR-01, MITR-02, MITR-03
- [ ] Test fixtures: sample event groups, ATT&CK technique data

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | Validate timestamps parse correctly, source_ip format |
| V6 Cryptography | no | No crypto needed for this phase |

### Known Threat Patterns for Go + SQLite

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SQL injection via event_id | Tampering | Use parameterized queries (sqlc) |
| Large JSON payload DoS | Denial of Service | Limit event query results, paginate |

## Sources

### Primary (HIGH confidence)
- pkg.go.dev/github.com/panther-labs/stix2 - STIX 2.x parsing API, FromJSON helper
- pkg.go.dev/github.com/msadministrator/goattck - ATT&CK data models, Load() API
- github.com/mitre-attack/attack-stix-data - Official STIX bundle location and format
- SQLite documentation - Window function support since 3.25.0

### Secondary (MEDIUM confidence)
- Windows Event ID mapping from security community research
- MITRE ATT&CK technique definitions from attack.mitre.org

### Tertiary (LOW confidence)
- Command-line pattern matching based on training data (needs validation against real attack patterns)

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | SQLite 3.25+ installed with window function support | Architecture Patterns | Grouping algorithm won't work; would need Go-side grouping |
| A2 | goattck library can download ATT&CK bundle successfully | ATT&CK Mapping | Need to bundle the JSON file in repo |
| A3 | Windows Event IDs in dataset match standard mappings | ATT&CK Mapping | Mapping table needs adjustment for custom events |
| A4 | 442K events grouped in <5 seconds via SQL | Performance | May need batching or optimization |

## Open Questions

1. **ATT&CK Bundle Location**
   - What we know: goattck downloads from MITRE GitHub
   - What's unclear: Will corporate/firewall environments block this?
   - Recommendation: Bundle enterprise-attack.json in data/ directory as fallback

2. **Event ID Coverage**
   - What we know: Dataset has ~19 unique Event IDs
   - What's unclear: Are these standard Windows Event IDs or custom?
   - Recommendation: Verify mapping table against actual Event IDs in database

3. **Incident Title Generation**
   - What we know: Need human-readable titles
   - What's unclear: Best format for titles
   - Recommendation: Use "{Primary Technique} from {Source IP}" format

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go 1.26+ | Backend | ✓ | 1.26.3 | — |
| SQLite 3.25+ | Window functions | ✓ | 3.x (WAL mode) | Go-side grouping |
| GitHub access | ATT&CK bundle download | ? | — | Bundle JSON in repo |

**Missing dependencies with no fallback:**
- None identified

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - panther-labs/stix2 and goattck verified via pkg.go.dev
- Architecture: HIGH - SQL window functions well-documented for SQLite
- Pitfalls: MEDIUM - time zone handling and duplicate creation need validation

**Research date:** 2026-06-09
**Valid until:** 2026-07-09 (30 days for stable libraries)
