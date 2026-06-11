'use client';

import { useState } from 'react';
import useSWR from 'swr';
import { Event } from '@/lib/types';
import { Crosshair, ChevronDown, ChevronUp } from 'lucide-react';
import { formatTimestamp, cn } from '@/lib/utils';
import { API_BASE } from '@/lib/api';

const fetcher = (url: string) => {
  const token = localStorage.getItem('nexus_token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return fetch(`${API_BASE}${url}`, { headers }).then((res) => res.json());
};

interface RawEventViewerProps {
  eventIds: number[];
  narrativeId: number;
}

export function RawEventViewer({ eventIds, narrativeId }: RawEventViewerProps) {
  const [expandedEventId, setExpandedEventId] = useState<number | null>(null);

  const { data, error, isLoading } = useSWR<{ events: Event[] }>(
    `/api/narratives/${narrativeId}`,
    fetcher
  );

  if (eventIds.length === 0) {
    return (
      <div className="bg-background border border-border rounded-lg p-6 text-center">
        <Crosshair className="w-8 h-8 text-muted-foreground/60 mx-auto mb-3" aria-hidden="true" />
        <p className="text-muted-foreground text-sm">
          Hover a sentence to view source events
        </p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-2" aria-label="Loading source events" aria-busy="true">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="h-20 bg-card rounded-lg animate-pulse" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4" role="alert">
        <p className="text-red-400 text-sm">Failed to load source events</p>
      </div>
    );
  }

  const filteredEvents = (data?.events || []).filter((e) =>
    eventIds.includes(e.id)
  );

  if (filteredEvents.length === 0) {
    return (
      <div className="bg-background border border-border rounded-lg p-4">
        <p className="text-muted-foreground text-sm">No matching events found</p>
      </div>
    );
  }

  return (
    <div className="space-y-2 max-h-[calc(100vh-8rem)] overflow-y-auto" role="list" aria-label={`Source events for selected sentence (${filteredEvents.length} events)`}>
      {filteredEvents.map((event) => (
        <div
          key={event.id}
          className="bg-background border border-border rounded-lg p-3"
          role="listitem"
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-2">
            <span className="font-mono text-xs text-muted-foreground">#{event.id}</span>
            <span
              className={cn(
                'text-xs px-2 py-0.5 rounded',
                event.event_type === 'authentication'
                  ? 'bg-primary/20 text-primary'
                  : event.event_type === 'process'
                  ? 'bg-event-process/10 text-event-process'
                  : event.event_type === 'network'
                  ? 'bg-event-network/10 text-event-network'
                  : 'bg-muted-foreground/20 text-muted-foreground'
              )}
            >
              {event.event_type}
            </span>
          </div>

          {/* Details */}
          <div className="grid grid-cols-2 gap-1 text-xs font-mono">
            <div className="text-muted-foreground">Time:</div>
            <div className="text-muted-foreground">{formatTimestamp(event.timestamp)}</div>
            <div className="text-muted-foreground">Host:</div>
            <div className="text-muted-foreground">{event.hostname}</div>
            <div className="text-muted-foreground">Source:</div>
            <div className="text-muted-foreground">{event.source_ip}</div>
            <div className="text-muted-foreground">Dest:</div>
            <div className="text-muted-foreground">{event.dest_ip}</div>
          </div>

          {/* Process info */}
          {(event.process_name || event.command_line) && (
            <div className="mt-2 text-xs font-mono border-t border-border pt-2">
              {event.process_name && (
                <div>
                  <span className="text-muted-foreground">Process: </span>
                  <span className="text-muted-foreground">{event.process_name}</span>
                </div>
              )}
              {event.command_line && (
                <div className="mt-1">
                  <span className="text-muted-foreground">Command: </span>
                  <span className="text-muted-foreground break-all">{event.command_line}</span>
                </div>
              )}
            </div>
          )}

          {/* Expandable raw JSON */}
          {event.raw_json && (
            <button
              onClick={() =>
                setExpandedEventId(expandedEventId === event.id ? null : event.id)
              }
              className="mt-2 flex items-center gap-1 text-xs text-muted-foreground/60 hover:text-muted-foreground"
              aria-expanded={expandedEventId === event.id}
              aria-label={`${expandedEventId === event.id ? 'Collapse' : 'Expand'} raw JSON for event ${event.id}`}
            >
              {expandedEventId === event.id ? (
                <ChevronUp className="w-3 h-3" aria-hidden="true" />
              ) : (
                <ChevronDown className="w-3 h-3" aria-hidden="true" />
              )}
              Raw JSON
            </button>
          )}
          {expandedEventId === event.id && event.raw_json && (
            <pre className="mt-2 text-xs text-muted-foreground bg-card rounded p-2 overflow-x-auto max-h-48 overflow-y-auto">
              {JSON.stringify(event.raw_json, null, 2)}
            </pre>
          )}
        </div>
      ))}
    </div>
  );
}
