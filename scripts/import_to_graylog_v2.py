#!/usr/bin/env python3
"""Import security events into Graylog via REST API."""

import json
import sys
import requests
from datetime import datetime

GRAYLOG_URL = "http://localhost:9000"
GRAYLOG_USER = "admin"
GRAYLOG_PASS = "admin"

def send_to_graylog(events, index="default"):
    """Send events to Graylog via GELF HTTP input."""
    batch = []
    for event in events:
        gelf = {
            "version": "1.1",
            "host": event.get("hostname", "unknown"),
            "short_message": f"{event.get('event_type', 'unknown')}: {event.get('command_line', event.get('description', ''))[:200]}",
            "timestamp": 0,
        }
        
        # Add all event fields with underscore prefix
        for k, v in event.items():
            if v and k != "timestamp":
                gelf[f"_{k}"] = str(v)[:500]
        
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
        
        batch.append(json.dumps(gelf))
    
    # Send via GELF HTTP
    try:
        resp = requests.post(
            "http://localhost:12201/gelf",
            data="\n".join(batch),
            headers={"Content-Type": "application/x-gelf"},
            timeout=30
        )
        return resp.ok
    except Exception as e:
        print(f"Error: {e}")
        return False

def check_input_status():
    """Check if GELF HTTP input is running."""
    try:
        resp = requests.get(
            f"{GRAYLOG_URL}/api/system/inputs",
            auth=(GRAYLOG_USER, GRAYLOG_PASS)
        )
        if resp.ok:
            for inp in resp.json().get("inputs", []):
                if "GELF HTTP" in inp.get("title", ""):
                    return True, inp["id"]
        return False, None
    except Exception as e:
        print(f"Error checking input: {e}")
        return False, None

def create_gelf_input():
    """Create a new GELF HTTP input."""
    payload = {
        "title": "GELF HTTP Import",
        "type": "org.graylog2.inputs.gelf.http.GELFHttpInput",
        "global": True,
        "configuration": {
            "bind_address": "0.0.0.0",
            "port": 12201,
            "recv_buffer_size": 1048576,
            "number_worker_threads": 4,
            "tcp_keepalive": False,
            "tls": False,
            "max_message_size": 2097152,
            "max_header_size": 65536
        }
    }
    
    try:
        resp = requests.post(
            f"{GRAYLOG_URL}/api/system/inputs",
            json=payload,
            auth=(GRAYLOG_USER, GRAYLOG_PASS),
            headers={"Content-Type": "application/json", "X-Requested-By": "python-import"}
        )
        if resp.ok:
            inp = resp.json()
            print(f"Created input: {inp['id']}")
            return True, inp["id"]
        else:
            print(f"Failed to create input: {resp.status_code}")
            return False, None
    except Exception as e:
        print(f"Error creating input: {e}")
        return False, None

if __name__ == "__main__":
    # Check if input exists
    input_exists, input_id = check_input_status()
    
    if not input_exists:
        print("Creating GELF HTTP input...")
        input_exists, input_id = create_gelf_input()
    
    if not input_exists:
        print("Failed to setup input. Make sure Graylog is running.")
        sys.exit(1)
    
    print(f"Using input: {input_id}")
    print("Importing events...")
    
    # Import events from JSON files
    total_imported = 0
    files = [
        "/home/piyuxhh/hackathon/data/sample_json_20260301/20251221.json",
        "/home/piyuxhh/hackathon/data/sample_json_20260301/20251222.json"
    ]
    
    for filepath in files:
        print(f"\nProcessing {filepath}...")
        batch = []
        file_count = 0
        
        try:
            with open(filepath) as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    
                    try:
                        event = json.loads(line)
                        batch.append(event)
                        file_count += 1
                        
                        # Send in batches of 100
                        if len(batch) >= 100:
                            if send_to_graylog(batch):
                                total_imported += len(batch)
                                print(f"  Imported {total_imported} events...")
                            batch = []
                    except json.JSONDecodeError:
                        continue
                    except Exception as e:
                        print(f"  Error processing event: {e}")
                        continue
            
            # Send remaining events
            if batch:
                if send_to_graylog(batch):
                    total_imported += len(batch)
                    print(f"  Imported {total_imported} events...")
        
        except Exception as e:
            print(f"Error processing file {filepath}: {e}")
    
    print(f"\n✅ Import complete! Total: {total_imported} events")
    print("Check Graylog at http://localhost:9000")
