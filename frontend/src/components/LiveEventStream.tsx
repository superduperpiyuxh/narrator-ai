'use client';

import { useState, useCallback, useRef } from 'react';
import { useSSE } from '@/hooks/useSSE';
import { cn, formatTimestamp, getEventDot } from '@/lib/utils';
import { Radio, ChevronDown, ChevronUp } from 'lucide-react';

interface StreamEvent {
  id?: number;
  timestamp: string;
  hostname: string;
  event_type: string;
  source_ip: string;
  user_name?: string;
}

export function LiveEventStream() {
  const [events, setEvents] = useState<StreamEvent[]>([]);
  const [isExpanded, setIsExpanded] = useState(true);
  const maxEvents = 50;
  const listRef = useRef<HTMLDivElement>(null);

  const handleEvent = useCallback((event: StreamEvent) => {
    setEvents(prev => {
      const next = [event, ...prev];
      return next.slice(0, maxEvents);
    });
  }, []);

  const { isConnected, eventCount } = useSSE({
    onEvent: handleEvent,
    enabled: true,
  });

  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden">
      {/* Header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between px-4 py-3 hover:bg-surface-hover transition-colors"
      >
        <div className="flex items-center gap-2">
          <div className="relative">
            <Radio className={cn(
              "w-4 h-4",
              isConnected ? "text-severity-low" : "text-muted-foreground"
            )} />
            {isConnected && (
              <span className="absolute -top-0.5 -right-0.5 w-2 h-2 bg-severity-low rounded-full animate-pulse" />
            )}
          </div>
          <span className="text-sm font-medium text-foreground">Live Events</span>
          {eventCount > 0 && (
            <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded-full">
              {eventCount}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className={cn(
            "text-xs px-2 py-0.5 rounded-full",
            isConnected
              ? "bg-severity-low/10 text-severity-low"
              : "bg-muted text-muted-foreground"
          )}>
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
          {isExpanded ? (
            <ChevronUp className="w-4 h-4 text-muted-foreground" />
          ) : (
            <ChevronDown className="w-4 h-4 text-muted-foreground" />
          )}
        </div>
      </button>

      {/* Event list */}
      {isExpanded && (
        <div
          ref={listRef}
          className="border-t border-border max-h-64 overflow-y-auto"
          aria-live="polite"
          aria-label="Live event stream"
        >
          {events.length === 0 ? (
            <div className="px-4 py-6 text-center text-muted-foreground text-sm">
              {isConnected ? 'Waiting for events...' : 'Connecting to event stream...'}
            </div>
          ) : (
            <div className="divide-y divide-border">
              {events.map((event, i) => (
                <div
                  key={`${event.timestamp}-${event.source_ip}-${i}`}
                  className={cn(
                    "px-4 py-2.5 flex items-center gap-3 text-xs font-mono animate-fade-in",
                    i === 0 && "bg-surface-hover/50"
                  )}
                >
                  <span className={cn("w-2 h-2 rounded-full flex-shrink-0", getEventDot(event.event_type))} />
                  <span className="text-muted-foreground w-16 flex-shrink-0">
                    {formatTimestamp(event.timestamp)}
                  </span>
                  <span className="text-foreground truncate flex-1">
                    {event.hostname}
                  </span>
                  <span className="text-muted-foreground truncate max-w-[120px]">
                    {event.source_ip}
                  </span>
                  <span className="text-muted-foreground truncate max-w-[100px]">
                    {event.event_type}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
