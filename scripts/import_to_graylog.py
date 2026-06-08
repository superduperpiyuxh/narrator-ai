#!/usr/bin/env python3
"""Import cyber_simulation security events into Graylog via GELF HTTP input."""

import json
import sys
import requests
from datetime import datetime

GRAYLOG_URL = "http://localhost:9000"
GRAYLOG_USER = "admin"
GRAYLOG_PASS = "admin"
GELF_INPUT_PORT = 12201

def setup_gelf_input():
    """Create a GELF HTTP input in Graylog."""
    # Check if input already exists
    resp = requests.get(f"{GRAYLOG_URL}/api/system/inputs", auth=(GRAYLOG_USER, GRAYLOG_PASS))
    if resp.ok:
        for inp in resp.json().get("inputs", []):
            if inp.get("title") == "GELF HTTP for demo import":
                print(f"Input already exists: {inp['id']}")
                return inp["id"]

    # Create GELF HTTP input
    payload = {
        "title": "GELF HTTP for demo import",
        "type": "org.graylog2.inputs.gelf.http.GELFHttpInput",
        "global": True,
        "configuration": {
            "bind_address": "0.0.0.0",
            "port": GELF_INPUT_PORT,
            "recv_buffer_size": 1048576,
            "number_worker_threads": 4,
            "tcp_keepalive": False,
            "tls": False,
            "max_message_size": 2097152,
            "max_header_size": 65536
        }
    }
    resp = requests.post(
        f"{GRAYLOG_URL}/api/system/inputs",
        json=payload,
        auth=(GRAYLOG_USER, GRAYLOG_PASS),
        headers={"Content-Type": "application/json", "X-Requested-By": "python-import"}
    )
    if resp.ok:
        inp = resp.json()
        print(f"Created GELF input: {inp['id']}")
        return inp["id"]
    else:
        print(f"Failed to create input: {resp.status_code} {resp.text}")
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

def import_file(filepath, batch_size=50):
    """Import a JSON file into Graylog via GELF HTTP."""
    total = 0
    errors = 0
    batch = []

    print(f"Importing {filepath}...")
    with open(filepath) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                event = json.loads(line)
                gelf = convert_to_gelf(event)
                batch.append(json.dumps(gelf))
                total += 1

                if len(batch) >= batch_size:
                    send_batch(batch)
                    batch = []
                    print(f"  Imported {total} events...")
            except json.JSONDecodeError:
                errors += 1
            except Exception as e:
                errors += 1

    if batch:
        send_batch(batch)

    print(f"Done: {total} imported, {errors} errors")
    return total

def send_batch(batch):
    """Send a batch of GELF messages to Graylog."""
    try:
        resp = requests.post(
            f"http://localhost:{GELF_INPUT_PORT}/gelf",
            data="\n".join(batch),
            headers={"Content-Type": "application/x-gelf"},
            timeout=30
        )
        if not resp.ok:
            print(f"  Batch send failed: {resp.status_code}")
    except Exception as e:
        print(f"  Batch send error: {e}")

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
