'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Search, ArrowRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { fetchIncidents } from '@/lib/api';
import type { Incident } from '@/lib/types';

interface CommandAction {
  id: string;
  label: string;
  description?: string;
  icon: React.ReactNode;
  action: () => void;
  category: string;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
}

export function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const staticActions: CommandAction[] = [
    {
      id: 'nav-dashboard',
      label: 'Dashboard',
      description: 'Go to main dashboard',
      icon: <Search className="w-4 h-4" />,
      action: () => { router.push('/'); onClose(); },
      category: 'Navigation',
    },
    {
      id: 'nav-settings',
      label: 'Settings',
      description: 'Manage API keys and preferences',
      icon: <Search className="w-4 h-4" />,
      action: () => { router.push('/settings'); onClose(); },
      category: 'Navigation',
    },
  ];

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await fetchIncidents(20, 0, undefined, undefined, undefined);
        const filtered = (res.incidents || []).filter(
          (inc) =>
            inc.title.toLowerCase().includes(query.toLowerCase()) ||
            inc.source_ip.includes(query) ||
            inc.severity.toLowerCase().includes(query.toLowerCase())
        );
        setResults(filtered);
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 200);

    return () => clearTimeout(timer);
  }, [query]);

  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setResults([]);
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [isOpen]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      const total = results.length + staticActions.length;

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex((prev) => (prev + 1) % total);
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex((prev) => (prev - 1 + total) % total);
          break;
        case 'Enter':
          e.preventDefault();
          if (selectedIndex < results.length) {
            router.push(`/incidents/${results[selectedIndex].id}`);
            onClose();
          } else {
            const actionIdx = selectedIndex - results.length;
            staticActions[actionIdx]?.action();
          }
          break;
        case 'Escape':
          onClose();
          break;
      }
    },
    [results, staticActions, selectedIndex, router, onClose]
  );

  if (!isOpen) return null;

  const totalItems = results.length + staticActions.length;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]">
      <div
        className="absolute inset-0 bg-overlay backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      <div
        className="relative w-full max-w-lg bg-card border border-border rounded-2xl shadow-2xl overflow-hidden animate-fade-in"
        role="dialog"
        aria-label="Command palette"
        onKeyDown={handleKeyDown}
      >
        <div className="flex items-center gap-3 px-4 border-b border-border">
          <Search className="w-5 h-5 text-muted-foreground flex-shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => { setQuery(e.target.value); setSelectedIndex(0); }}
            placeholder="Search incidents, navigate..."
            className="flex-1 bg-transparent py-4 text-foreground placeholder:text-muted-foreground outline-none text-sm"
            aria-label="Search command palette"
          />
          <kbd className="hidden sm:inline-flex items-center gap-1 text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded border border-border">
            ESC
          </kbd>
        </div>

        <div className="max-h-80 overflow-y-auto">
          {query.trim() === '' ? (
            <div className="p-2">
              {staticActions.map((action, i) => (
                <button
                  key={action.id}
                  onClick={action.action}
                  className={cn(
                    "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors",
                    selectedIndex === results.length + i
                      ? "bg-surface-hover text-foreground"
                      : "text-muted-foreground hover:bg-surface-hover hover:text-foreground"
                  )}
                >
                  <span className="text-muted-foreground">{action.icon}</span>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium">{action.label}</div>
                    {action.description && (
                      <div className="text-xs text-muted-foreground truncate">{action.description}</div>
                    )}
                  </div>
                  <ArrowRight className="w-3 h-3 text-muted-foreground" />
                </button>
              ))}
            </div>
          ) : loading ? (
            <div className="p-8 text-center text-muted-foreground text-sm">
              Searching...
            </div>
          ) : results.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground text-sm">
              No incidents found for &quot;{query}&quot;
            </div>
          ) : (
            <div className="p-2">
              <div className="px-3 py-1.5 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Incidents
              </div>
              {results.map((incident, i) => (
                <button
                  key={incident.id}
                  onClick={() => {
                    router.push(`/incidents/${incident.id}`);
                    onClose();
                  }}
                  className={cn(
                    "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors",
                    selectedIndex === i
                      ? "bg-surface-hover text-foreground"
                      : "text-muted-foreground hover:bg-surface-hover hover:text-foreground"
                  )}
                >
                  <span className={cn(
                    "w-2 h-2 rounded-full flex-shrink-0",
                    incident.severity === 'critical' ? 'bg-severity-critical' :
                    incident.severity === 'high' ? 'bg-severity-high' :
                    incident.severity === 'medium' ? 'bg-severity-medium' :
                    'bg-severity-low'
                  )} />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium truncate">{incident.title}</div>
                    <div className="text-xs text-muted-foreground">
                      {incident.source_ip} · {incident.event_count} events · {incident.severity}
                    </div>
                  </div>
                  <ArrowRight className="w-3 h-3 text-muted-foreground flex-shrink-0" />
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="px-4 py-2 border-t border-border flex items-center gap-4 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <kbd className="bg-muted px-1 rounded">↑↓</kbd> navigate
          </span>
          <span className="flex items-center gap-1">
            <kbd className="bg-muted px-1 rounded">↵</kbd> select
          </span>
          <span className="flex items-center gap-1">
            <kbd className="bg-muted px-1 rounded">esc</kbd> close
          </span>
        </div>
      </div>
    </div>
  );
}
