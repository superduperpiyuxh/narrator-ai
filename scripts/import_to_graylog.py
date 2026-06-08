#!/usr/bin/env python3
"""Import cyber_simulation security events into Graylog via GELF HTTP input."""

import json
import sys
import time
import requests
from datetime import datetime

GRAYLOG_URL = "http://localhost:9000"
GRAYLOG_USER = "admin"
GRAYLOG_PASS = "admin"
GELF_INPUT_PORT = 12201

def setup_gelf_input():
    """Check if GELF HTTP input exists, don't create if it does."""
    resp = requests.get(f"{GRAYLOG_URL}/api/system/inputs", auth=(GRAYLOG_USER, GRAYLOG_PASS))
    if resp.ok:
        for inp in resp.json().get("inputs", []):
            if "GELF" in inp.get("title", "").upper() and "HTTP" in inp.get("title", "").upper():
                print(f"Using existing input: {inp['id']}")
                return inp["id"]
    print("No GELF HTTP input found. Create one in Graylog UI first.")
    return None

def convert_to_gelf(event):
    """Convert a security event to GELF format."""
    gelf = {
        "version": "1.1",
        "host": event.get("hostname", "unknown"),
        "short_message": f"{event.get('event_type', 'unknown')}: {event.get('command_line', event.get('description', ''))}",
        "timestamp": 0,
        "_event_type": event.get("event_type", ""),
        "_user": event.get("user", event.get("account", "")),
        "_source_ip": event.get("source_ip", ""),
        "_dest_ip": event.get("dest_ip", ""),
        "_process_name": event.get("process_name", ""),
        "_command_line": event.get("command_line", ""),
        "_log_type": event.get("log_type", ""),
        "_event_id": event.get("event_id", ""),
        "_department": event.get("department", ""),
        "_location": event.get("location", ""),
        "_device_type": event.get("device_type", ""),
        "_success": event.get("success", ""),
        "_session_id": event.get("session_id", ""),
        "_parent_process": event.get("parent_process", ""),
    }

    # Parse timestamp
    ts = event.get("timestamp", "")
    if ts:
        try:
            dt = datetime.strptime(ts, "%Y-%m-%d %H:%M:%S")
            gelf["timestamp"] = dt.timestamp()
        except ValueError:
            try:
                dt = datetime.fromisoformat(ts.replace("Z", "+00:00"))
                gelf["timestamp"] = dt.timestamp()
            except ValueError:
                gelf["timestamp"] = datetime.now().timestamp()

    # Add attack fields if present
    for field in ["attack_id", "attack_type", "stage_number", "severity", "alert_name"]:
        if field in event and event[field]:
            gelf[f"_{field}"] = event[field]

    return gelf

def import_file(filepath):
    """Import a JSON file into Graylog via GELF HTTP (one event per request)."""
    total = 0
    errors = 0

    print(f"Importing {filepath}...")
    with open(filepath) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                event = json.loads(line)
                gelf = convert_to_gelf(event)
                if send_event(json.dumps(gelf)):
                    total += 1
                else:
                    errors += 1
                if total % 5000 == 0:
                    print(f"  Imported {total} events...")
            except json.JSONDecodeError:
                errors += 1
            except Exception as e:
                errors += 1

    print(f"Done: {total} imported, {errors} errors")
    return total

def send_event(gelf_json, retries=3):
    """Send a single GELF message to Graylog with retry logic."""
    for attempt in range(retries):
        try:
            resp = requests.post(
                f"http://localhost:{GELF_INPUT_PORT}/gelf",
                data=gelf_json,
                headers={"Content-Type": "application/x-gelf"},
                timeout=30
            )
            if resp.ok:
                return True
            else:
                print(f"  Error {resp.status_code}: {resp.text[:100]}")
        except requests.exceptions.ConnectionError:
            time.sleep(2)
        except Exception as e:
            time.sleep(1)
    return False

if __name__ == "__main__":
    print("Setting up Graylog GELF input...")
    input_id = setup_gelf_input()
    if not input_id:
        print("Failed to setup input. Make sure Graylog is running.")
        sys.exit(1)

    # Import both days
    import_file("/home/piyuxhh/hackathon/data/sample_json_20260301/20251221.json")
    import_file("/home/piyuxhh/hackathon/data/sample_json_20260301/20251222.json")

    print("\n✅ Import complete! Check Graylog at http://localhost:9000")
