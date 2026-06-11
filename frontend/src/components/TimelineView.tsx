'use client';

import { useState, useEffect, useMemo } from 'react';
import { Event } from '@/lib/types';
import { fetchIncidentEvents } from '@/lib/api';
import { formatTimestamp, cn } from '@/lib/utils';
import {
  Clock,
  ChevronDown,
  ChevronUp,
  ArrowDown,
  Shield,
  Network,
  FileText,
  Terminal,
  Hash,
} from 'lucide-react';

function getEventTypeColor(eventType: string): {
  dot: string;
  bg: string;
  text: string;
  border: string;
  icon: string;
} {
  const lower = eventType.toLowerCase();
  if (lower === 'authentication') {
    return {
      dot: 'bg-event-auth',
      bg: 'bg-event-auth/10',
      text: 'text-event-auth',
      border: 'border-event-auth/30',
      icon: 'text-event-auth',
    };
  }
  if (lower === 'process_activity' || lower === 'process') {
    return {
      dot: 'bg-event-process',
      bg: 'bg-event-process/10',
      text: 'text-event-process',
      border: 'border-event-process/30',
      icon: 'text-event-process',
    };
  }
  if (lower === 'network_activity' || lower === 'network') {
    return {
      dot: 'bg-event-network',
      bg: 'bg-event-network/10',
      text: 'text-event-network',
      border: 'border-event-network/30',
      icon: 'text-event-network',
    };
  }
  if (
    lower === 'file_activity' ||
    lower === 'file_create' ||
    lower === 'file_delete'
  ) {
    return {
      dot: 'bg-event-file',
      bg: 'bg-event-file/10',
      text: 'text-event-file',
      border: 'border-event-file/30',
      icon: 'text-event-file',
    };
  }
  return {
    dot: 'bg-zinc-500',
    bg: 'bg-zinc-500/10',
    text: 'text-muted-foreground',
    border: 'border-muted/30',
    icon: 'text-muted-foreground',
  };
}

function getEventIcon(eventType: string) {
  const lower = eventType.toLowerCase();
  if (lower === 'authentication') return Shield;
  if (lower === 'process_activity' || lower === 'process') return Terminal;
  if (lower === 'network_activity' || lower === 'network') return Network;
  if (
    lower === 'file_activity' ||
    lower === 'file_create' ||
    lower === 'file_delete'
  )
    return FileText;
  return Hash;
}

function formatDuration(startMs: number, endMs: number): string {
  const diff = endMs - startMs;
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

interface TimelineEventProps {
  event: Event;
  isLast: boolean;
}

function TimelineEvent({ event, isLast }: TimelineEventProps) {
  const [expanded, setExpanded] = useState(false);
  const colors = getEventTypeColor(event.event_type);
  const EventIcon = getEventIcon(event.event_type);

  return (
    <div className="relative flex gap-4 group">
      {/* Vertical line + dot */}
      <div className="flex flex-col items-center">
        <div
          className={cn(
            'w-3 h-3 rounded-full border-2 border-background z-10 flex-shrink-0 mt-1',
            colors.dot
          )}
        />
        {!isLast && (
          <div className="w-px flex-1 bg-muted-foreground/60 min-h-[2rem]" />
        )}
      </div>

      {/* Event card */}
      <div
        className={cn(
          'flex-1 rounded-lg border p-3 mb-3 transition-colors',
          colors.bg,
          colors.border,
          'hover:border-border'
        )}
      >
        {/* Header row */}
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <EventIcon
              className={cn('w-3.5 h-3.5', colors.icon)}
              aria-hidden="true"
            />
            <span className={cn('text-xs font-medium uppercase tracking-wide', colors.text)}>
              {event.event_type.replace(/_/g, ' ')}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-muted-foreground/60">
              #{event.id}
            </span>
            <button
              onClick={() => setExpanded(!expanded)}
              className="text-muted-foreground/60 hover:text-muted-foreground transition-colors p-0.5"
              aria-label={expanded ? 'Collapse event details' : 'Expand event details'}
              aria-expanded={expanded}
            >
              {expanded ? (
                <ChevronUp className="w-3.5 h-3.5" aria-hidden="true" />
              ) : (
                <ChevronDown className="w-3.5 h-3.5" aria-hidden="true" />
              )}
            </button>
          </div>
        </div>

        {/* Timestamp */}
        <div className="flex items-center gap-1.5 mb-2">
          <Clock className="w-3 h-3 text-muted-foreground" aria-hidden="true" />
          <span className="text-xs text-muted-foreground">
            {formatTimestamp(event.timestamp)}
          </span>
        </div>

        {/* Key fields */}
        <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-xs font-mono">
          {event.hostname && (
            <>
              <span className="text-muted-foreground">Host</span>
              <span className="text-foreground/80 truncate">{event.hostname}</span>
            </>
          )}
          {event.source_ip && (
            <>
              <span className="text-muted-foreground">Source IP</span>
              <span className="text-foreground/80 truncate">{event.source_ip}</span>
            </>
          )}
          {event.process_name && (
            <>
              <span className="text-muted-foreground">Process</span>
              <span className="text-foreground/80 truncate">{event.process_name}</span>
            </>
          )}
          {event.user_name && (
            <>
              <span className="text-muted-foreground">User</span>
              <span className="text-foreground/80 truncate">{event.user_name}</span>
            </>
          )}
          {event.dest_ip && (
            <>
              <span className="text-muted-foreground">Dest IP</span>
              <span className="text-foreground/80 truncate">{event.dest_ip}</span>
            </>
          )}
        </div>

        {/* Expanded details */}
        {expanded && (
          <div className="mt-3 pt-3 border-t border-border space-y-1 text-xs font-mono">
            {event.command_line && (
              <div>
                <span className="text-muted-foreground">Command: </span>
                <span className="text-muted-foreground break-all">{event.command_line}</span>
              </div>
            )}
            {event.parent_process && (
              <div>
                <span className="text-muted-foreground">Parent Process: </span>
                <span className="text-muted-foreground">{event.parent_process}</span>
              </div>
            )}
            {event.log_type && (
              <div>
                <span className="text-muted-foreground">Log Type: </span>
                <span className="text-muted-foreground">{event.log_type}</span>
              </div>
            )}
            {event.session_id && (
              <div>
                <span className="text-muted-foreground">Session ID: </span>
                <span className="text-muted-foreground">{event.session_id}</span>
              </div>
            )}
            {event.file_path && (
              <div>
                <span className="text-muted-foreground">File Path: </span>
                <span className="text-muted-foreground break-all">{event.file_path}</span>
              </div>
            )}
            {event.protocol && (
              <div>
                <span className="text-muted-foreground">Protocol: </span>
                <span className="text-muted-foreground">{event.protocol}</span>
              </div>
            )}
            {event.port && (
              <div>
                <span className="text-muted-foreground">Port: </span>
                <span className="text-muted-foreground">{event.port}</span>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

interface TimelineViewProps {
  incidentId: number;
}

export function TimelineView({ incidentId }: TimelineViewProps) {
  const [events, setEvents] = useState<Event[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAll, setShowAll] = useState(false);
  const VISIBLE_LIMIT = 50;

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        setLoading(true);
        setError(null);
        const res = await fetchIncidentEvents(incidentId);
        if (!cancelled) {
          setEvents(res.events || []);
          setTotal(res.total || 0);
        }
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : 'Failed to load events');
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, [incidentId]);

  const summary = useMemo(() => {
    if (events.length === 0) return null;

    const timestamps = events
      .map((e) => new Date(e.timestamp).getTime())
      .filter((t) => !isNaN(t));
    const minTime = Math.min(...timestamps);
    const maxTime = Math.max(...timestamps);

    const actors = new Set<string>();
    events.forEach((e) => {
      if (e.user_name) actors.add(e.user_name);
      if (e.source_ip) actors.add(e.source_ip);
      if (e.hostname) actors.add(e.hostname);
    });

    return {
      total: events.length,
      timeSpan: formatDuration(minTime, maxTime),
      uniqueActors: actors.size,
      start: timestamps.length > 0 ? new Date(minTime).toLocaleString() : '-',
      end: timestamps.length > 0 ? new Date(maxTime).toLocaleString() : '-',
    };
  }, [events]);

  if (loading) {
    return (
      <div className="space-y-4" aria-label="Loading timeline events" aria-busy="true">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="flex gap-4">
            <div className="w-3 h-3 rounded-full bg-surface animate-pulse mt-1" />
            <div className="flex-1 h-24 bg-card rounded-lg animate-pulse" />
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4" role="alert">
        <p className="text-red-400 text-sm">Failed to load timeline: {error}</p>
      </div>
    );
  }

  if (events.length === 0) {
    return (
      <div className="bg-card border border-border rounded-lg p-8 text-center">
        <Clock className="w-8 h-8 text-muted-foreground/60 mx-auto mb-3" aria-hidden="true" />
        <p className="text-muted-foreground text-sm">No events found for this incident</p>
      </div>
    );
  }

  const visibleEvents = showAll ? events : events.slice(0, VISIBLE_LIMIT);
  const hasMore = events.length > VISIBLE_LIMIT && !showAll;

  return (
    <div>
      {/* Summary bar */}
      {summary && (
        <div
          className="flex flex-wrap items-center gap-4 mb-6 p-3 bg-card border border-border rounded-lg"
          role="region"
          aria-label="Timeline summary"
        >
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-muted-foreground" aria-hidden="true" />
            <span className="text-sm font-medium text-foreground/80">
              {summary.total.toLocaleString()} events
            </span>
          </div>
          <div className="text-muted-foreground/60">|</div>
          <div className="text-xs text-muted-foreground">
            <span className="text-muted-foreground">{summary.timeSpan}</span> span
          </div>
          <div className="text-muted-foreground/60">|</div>
          <div className="text-xs text-muted-foreground">
            <span className="text-muted-foreground">{summary.uniqueActors}</span> unique actors
          </div>
          <div className="text-muted-foreground/60">|</div>
          <div className="text-xs text-muted-foreground">
            {summary.start} &rarr; {summary.end}
          </div>
        </div>
      )}

      {/* Event type legend */}
      <div className="flex flex-wrap gap-3 mb-6" aria-label="Event type legend">
        {[
          { type: 'authentication', color: 'bg-event-auth', label: 'Authentication' },
          { type: 'process', color: 'bg-event-process', label: 'Process' },
          { type: 'network', color: 'bg-event-network', label: 'Network' },
          { type: 'file', color: 'bg-event-file', label: 'File' },
          { type: 'other', color: 'bg-zinc-500', label: 'Other' },
        ].map(({ type, color, label }) => (
          <div key={type} className="flex items-center gap-1.5">
            <div className={cn('w-2 h-2 rounded-full', color)} />
            <span className="text-xs text-muted-foreground">{label}</span>
          </div>
        ))}
      </div>

      {/* Timeline */}
      <div
        className="max-h-[calc(100vh-28rem)] overflow-y-auto pr-1"
        role="list"
        aria-label={`Incident timeline with ${events.length} events`}
      >
        {visibleEvents.map((event, index) => (
          <div key={event.id} role="listitem">
            <TimelineEvent
              event={event}
              isLast={index === visibleEvents.length - 1}
            />
          </div>
        ))}
      </div>

      {/* Show more button */}
      {hasMore && (
        <button
          onClick={() => setShowAll(true)}
          className="mt-4 flex items-center gap-2 mx-auto text-sm text-muted-foreground hover:text-foreground/80 transition-colors px-4 py-2 rounded-lg bg-card border border-border hover:border-border"
        >
          <ArrowDown className="w-4 h-4" aria-hidden="true" />
          Show all {events.length} events ({events.length - VISIBLE_LIMIT} more)
        </button>
      )}
    </div>
  );
}
