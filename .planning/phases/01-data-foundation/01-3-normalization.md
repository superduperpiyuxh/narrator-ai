# Plan 01-3: Event Field Normalization

## Goal
Normalize event fields during ingestion: parse timestamps, map event types to enum, cast booleans.

## Depends On
01-2 (SQLite schema)

## Requirements
ING-03: Parse and normalize event fields

## Technical Details

### Normalization Rules

1. **Timestamp Normalization**
   - Input: `"2025-12-21 09:19:00"` (string)
   - Output: `"2025-12-21T09:19:00Z"` (ISO 8601)
   - Handle: Missing timestamps → use import time

2. **Event Type Enum**
   Map raw strings to canonical categories:
   ```
   process_start → process_activity
   process_termination → process_activity
   network_connection → network_activity
   network_connection_failed → network_activity
   login_attempt → authentication
   login_success → authentication
   login_failure → authentication
   admin_action → privilege_escalation
   file_access → file_activity
   web_browsing → web_activity
   kerberos_auth_success → authentication
   api_key_auth_success → authentication
   api_key_auth_failure → authentication
   certificate_auth_success → authentication
   oauth_token_success → authentication
   system_event → system
   ```

3. **Boolean Conversion**
   - `"true"` → `1`
   - `"false"` → `0`
   - `""` / `null` → `0`

4. **IP Address Validation**
   - Validate format: `xxx.xxx.xxx.xxx`
   - Invalid → `NULL`

5. **Null Handling**
   - Empty string `""` → `NULL`
   - Missing fields → `NULL`

### Implementation
1. Create `backend/internal/normalizer/normalizer.go`
2. Implement `NormalizeEvent(raw map[string]interface{}) Event`
3. Add normalization step in ingestion pipeline
4. Unit tests for each normalization rule

## Verification
- Timestamps stored as ISO 8601
- Event types are normalized categories
- Success field is boolean (0/1)
- Invalid IPs stored as NULL
- Empty strings stored as NULL
