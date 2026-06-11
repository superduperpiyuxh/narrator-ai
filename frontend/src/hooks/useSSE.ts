'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { API_BASE } from '@/lib/api';

interface SSEEvent {
  id?: number;
  timestamp: string;
  hostname: string;
  event_type: string;
  event_id: string;
  user_name: string;
  source_ip: string;
  command_line?: string;
  process_name?: string;
}

interface UseSSEOptions {
  onEvent?: (event: SSEEvent) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  enabled?: boolean;
}

export function useSSE({ onEvent, onConnect, onDisconnect, enabled = true }: UseSSEOptions = {}) {
  const [isConnected, setIsConnected] = useState(false);
  const [eventCount, setEventCount] = useState(0);
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttempts = useRef(0);

  const connect = useCallback(() => {
    if (!enabled) return;

    const token = typeof window !== 'undefined' ? localStorage.getItem('nexus_token') : null;
    if (!token) return;

    // EventSource doesn't support custom headers, so we pass token as query param
    const url = `${API_BASE}/api/v1/stream?token=${encodeURIComponent(token)}`;

    try {
      const es = new EventSource(url);
      eventSourceRef.current = es;

      es.onopen = () => {
        setIsConnected(true);
        reconnectAttempts.current = 0;
        onConnect?.();
      };

      es.addEventListener('event', (e) => {
        try {
          const data = JSON.parse(e.data) as SSEEvent;
          setEventCount(prev => prev + 1);
          onEvent?.(data);
        } catch {
          // ignore parse errors
        }
      });

      es.onerror = () => {
        setIsConnected(false);
        es.close();
        eventSourceRef.current = null;
        onDisconnect?.();

        // Reconnect with backoff
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000);
        reconnectAttempts.current++;
        reconnectTimeoutRef.current = setTimeout(connect, delay);
      };
    } catch {
      // Connection failed, retry
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000);
      reconnectAttempts.current++;
      reconnectTimeoutRef.current = setTimeout(connect, delay);
    }
  }, [enabled, onEvent, onConnect, onDisconnect]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsConnected(false);
  }, []);

  useEffect(() => {
    if (enabled) {
      connect();
    }
    return disconnect;
  }, [enabled, connect, disconnect]);

  return { isConnected, eventCount, disconnect };
}
