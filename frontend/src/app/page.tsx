'use client';

import { useEffect, useState, useRef, useCallback, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { isAuthenticated, clearToken, fetchIncidents, fetchIncidentStats } from '@/lib/api';
import { DashboardControls, type DashboardControlsHandle } from '@/components/DashboardControls';
import { KeyboardShortcutsModal } from '@/components/KeyboardShortcutsModal';
import { CommandPalette } from '@/components/CommandPalette';
import { useCommandPalette } from '@/hooks/useCommandPalette';
import { LiveEventStream } from '@/components/LiveEventStream';
import { TechniqueHeatmap } from '@/components/TechniqueHeatmap';
import { Shield, Keyboard, Search, X, ChevronLeft, ChevronRight } from 'lucide-react';
import { cn, formatTimestamp, getSeverityDot, getSeverityBorder } from '@/lib/utils';
import type { Incident, IncidentStats } from '@/lib/types';

export default function HomePage() {
  return (
    <Suspense fallback={<LoadingSkeleton />}>
      <HomePageContent />
    </Suspense>
  );
}

function LoadingSkeleton() {
  return (
    <main className="min-h-screen bg-background flex items-center justify-center">
      <div className="text-muted-foreground" role="status">Loading dashboard...</div>
    </main>
  );
}

function HomePageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [total, setTotal] = useState(0);
  const [stats, setStats] = useState<IncidentStats | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [severity, setSeverity] = useState(searchParams.get('severity') || '');
  const [status, setStatus] = useState(searchParams.get('status') || '');
  const [sourceIP, setSourceIP] = useState(searchParams.get('source_ip') || '');
  const limit = 24;

  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [showShortcuts, setShowShortcuts] = useState(false);
  const [pendingKey, setPendingKey] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const dashboardControlsRef = useRef<DashboardControlsHandle>(null);
  const sidebarListRef = useRef<HTMLDivElement>(null);
  const { isOpen: isPaletteOpen, open: openPalette, close: closePalette } = useCommandPalette();

  useEffect(() => {
    const sev = searchParams.get('severity') || '';
    const st = searchParams.get('status') || '';
    const ip = searchParams.get('source_ip') || '';
    const p = parseInt(searchParams.get('page') || '1', 10);
    setSeverity(sev);
    setStatus(st);
    setSourceIP(ip);
    setPage(isNaN(p) ? 1 : p);
  }, [searchParams]);

  useEffect(() => {
    const loadData = async () => {
      try {
        const offset = (page - 1) * limit;
        const [incRes, statsRes] = await Promise.all([
          fetchIncidents(limit, offset, severity || undefined, status || undefined, sourceIP || undefined),
          fetchIncidentStats(),
        ]);
        setIncidents(incRes.incidents || []);
        setTotal(incRes.total);
        setStats(statsRes);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load');
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [page, severity, status, sourceIP]);

  useEffect(() => {
    setSelectedIndex(-1);
  }, [incidents]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable
      ) {
        if (e.key === 'Escape') {
          (target as HTMLElement).blur();
          dashboardControlsRef.current?.clearSearch();
        }
        return;
      }

      if (showShortcuts) {
        if (e.key === 'Escape' || e.key === '?') {
          setShowShortcuts(false);
          e.preventDefault();
        }
        return;
      }

      if (pendingKey === 'g') {
        setPendingKey(null);
        if (e.key === 'h') {
          router.push('/');
          e.preventDefault();
        } else if (e.key === 's') {
          router.push('/settings');
          e.preventDefault();
        }
        return;
      }

      switch (e.key) {
        case '/':
          e.preventDefault();
          dashboardControlsRef.current?.focusSearch();
          break;
        case '?':
          e.preventDefault();
          setShowShortcuts(true);
          break;
        case 'j':
          e.preventDefault();
          setSelectedIndex((prev) => Math.min(prev + 1, incidents.length - 1));
          break;
        case 'k':
          e.preventDefault();
          setSelectedIndex((prev) => Math.max(prev - 1, 0));
          break;
        case 'Enter':
          e.preventDefault();
          if (selectedIndex >= 0 && selectedIndex < incidents.length) {
            router.push(`/incidents/${incidents[selectedIndex].id}`);
          }
          break;
        case 'g':
          e.preventDefault();
          setPendingKey('g');
          break;
        case 'Escape':
          e.preventDefault();
          setSelectedIndex(-1);
          break;
      }
    },
    [showShortcuts, pendingKey, selectedIndex, incidents, router]
  );

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  useEffect(() => {
    if (selectedIndex >= 0 && sidebarListRef.current) {
      const item = sidebarListRef.current.children[selectedIndex] as HTMLElement | undefined;
      if (item) {
        item.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
      }
    }
  }, [selectedIndex]);

  const handleLogout = () => {
    clearToken();
    window.location.reload();
  };

  if (loading) {
    return <LoadingSkeleton />;
  }

  const totalPages = Math.ceil(total / limit);

  const statCards = stats
    ? [
        {
          label: 'Total Incidents',
          value: stats.total_incidents,
          color: 'text-foreground',
          sparkColor: 'text-muted-foreground',
          trend: '+12%',
        },
        stats.by_severity?.critical
          ? { label: 'Critical', value: stats.by_severity.critical, color: 'text-severity-critical', sparkColor: 'text-severity-critical/40', trend: '+5%' }
          : null,
        stats.by_severity?.high
          ? { label: 'High', value: stats.by_severity.high, color: 'text-severity-high', sparkColor: 'text-severity-high/40', trend: '-3%' }
          : null,
        stats.by_severity?.medium
          ? { label: 'Medium', value: stats.by_severity.medium, color: 'text-severity-medium', sparkColor: 'text-severity-medium/40', trend: '+8%' }
          : null,
      ].filter(Boolean)
    : [];

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <CommandPalette isOpen={isPaletteOpen} onClose={closePalette} />

      {/* Header */}
      <header className="h-14 border-b border-border bg-card flex items-center justify-between px-4 flex-shrink-0 z-10">
        <div className="flex items-center gap-3">
          <Shield className="w-6 h-6 text-primary" aria-hidden="true" />
          <h1 className="text-lg font-bold text-foreground tracking-tight">Nexus</h1>
          <button
            onClick={openPalette}
            className="hidden sm:flex items-center gap-2 ml-4 px-3 py-1.5 bg-surface border border-border rounded-lg text-sm text-muted-foreground hover:bg-surface-hover transition-colors"
            aria-label="Open command palette"
          >
            <Search className="w-3.5 h-3.5" />
            <span>Search incidents...</span>
            <kbd className="text-[10px] bg-muted px-1.5 py-0.5 rounded border border-border ml-4">Ctrl+K</kbd>
          </button>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowShortcuts(true)}
            className="p-2 text-muted-foreground hover:text-foreground transition-colors rounded-md"
            aria-label="Show keyboard shortcuts"
            title="Keyboard shortcuts (?)"
          >
            <Keyboard className="w-4 h-4" />
          </button>
          {isAuthenticated() ? (
            <>
              <Link
                href="/settings"
                className="px-3 py-1.5 bg-surface border border-border text-sm text-muted-foreground rounded-md hover:bg-surface-hover hover:text-foreground transition-colors"
              >
                Settings
              </Link>
              <button
                onClick={handleLogout}
                className="px-3 py-1.5 bg-surface border border-border text-sm text-muted-foreground rounded-md hover:bg-surface-hover hover:text-foreground transition-colors"
              >
                Sign Out
              </button>
            </>
          ) : (
            <>
              <Link
                href="/login"
                className="px-3 py-1.5 bg-surface border border-border text-sm text-muted-foreground rounded-md hover:bg-surface-hover hover:text-foreground transition-colors"
              >
                Sign In
              </Link>
              <Link
                href="/signup"
                className="px-3 py-1.5 bg-primary text-primary-foreground text-sm rounded-md hover:bg-primary/90 transition-colors"
              >
                Sign Up
              </Link>
            </>
          )}
        </div>
      </header>

      {/* Main layout: sidebar + content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <aside
          className={cn(
            'flex-shrink-0 border-r border-border bg-sidebar-bg flex flex-col overflow-hidden transition-all duration-300',
            sidebarOpen ? 'w-[320px]' : 'w-0',
            'max-lg:absolute max-lg:inset-y-0 max-lg:left-0 max-lg:z-20 max-lg:shadow-2xl',
            !sidebarOpen && 'max-lg:w-0'
          )}
          aria-label="Incidents sidebar"
        >
          <div className="flex flex-col h-full w-[320px]">
            {/* Sidebar header */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-border flex-shrink-0">
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold text-foreground">Incidents</span>
                <span className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded-full">
                  {total}
                </span>
              </div>
              <button
                onClick={() => setSidebarOpen(false)}
                className="lg:hidden p-1 text-muted-foreground hover:text-foreground transition-colors"
                aria-label="Close sidebar"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Sidebar filters */}
            <div className="px-4 py-3 border-b border-border flex-shrink-0 space-y-2">
              <div className="flex gap-2">
                <select
                  value={severity}
                  onChange={(e) => {
                    const params = new URLSearchParams();
                    if (e.target.value) params.set('severity', e.target.value);
                    if (status) params.set('status', status);
                    if (sourceIP) params.set('source_ip', sourceIP);
                    router.push(`/?${params.toString()}`);
                  }}
                  className="flex-1 bg-surface border border-border rounded-md px-2 py-1.5 text-xs text-foreground focus:outline-none focus:border-primary"
                  aria-label="Filter by severity"
                >
                  <option value="">All Severity</option>
                  <option value="critical">Critical</option>
                  <option value="high">High</option>
                  <option value="medium">Medium</option>
                  <option value="low">Low</option>
                </select>
                <select
                  value={status}
                  onChange={(e) => {
                    const params = new URLSearchParams();
                    if (severity) params.set('severity', severity);
                    if (e.target.value) params.set('status', e.target.value);
                    if (sourceIP) params.set('source_ip', sourceIP);
                    router.push(`/?${params.toString()}`);
                  }}
                  className="flex-1 bg-surface border border-border rounded-md px-2 py-1.5 text-xs text-foreground focus:outline-none focus:border-primary"
                  aria-label="Filter by status"
                >
                  <option value="">All Status</option>
                  <option value="open">Open</option>
                  <option value="investigating">Investigating</option>
                  <option value="resolved">Resolved</option>
                </select>
              </div>
              <div className="relative">
                <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
                <input
                  type="text"
                  value={sourceIP}
                  onChange={(e) => {
                    const params = new URLSearchParams();
                    if (severity) params.set('severity', severity);
                    if (status) params.set('status', status);
                    if (e.target.value) params.set('source_ip', e.target.value);
                    router.push(`/?${params.toString()}`);
                  }}
                  placeholder="Filter by source IP..."
                  className="w-full bg-surface border border-border rounded-md pl-7 pr-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-primary"
                  aria-label="Filter by source IP"
                />
                {sourceIP && (
                  <button
                    onClick={() => {
                      const params = new URLSearchParams();
                      if (severity) params.set('severity', severity);
                      if (status) params.set('status', status);
                      router.push(`/?${params.toString()}`);
                    }}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    aria-label="Clear source IP filter"
                  >
                    <X className="w-3 h-3" />
                  </button>
                )}
              </div>
              {(severity || status || sourceIP) && (
                <button
                  onClick={() => router.push('/')}
                  className="text-[11px] text-muted-foreground hover:text-foreground underline transition-colors"
                >
                  Clear all filters
                </button>
              )}
            </div>

            {/* Incident list */}
            <div
              ref={sidebarListRef}
              className="flex-1 overflow-y-auto"
              role="list"
              aria-label="Incidents list"
            >
              {incidents.length > 0 ? (
                incidents.map((incident, index) => (
                  <a
                    key={incident.id}
                    href={`/incidents/${incident.id}`}
                    role="listitem"
                    className={cn(
                      'block border-l-3 px-4 py-3 transition-all duration-150 hover:bg-sidebar-hover',
                      getSeverityBorder(incident.severity),
                      index === selectedIndex
                        ? 'bg-sidebar-active'
                        : 'bg-transparent'
                    )}
                    aria-label={`Incident ${incident.id}: ${incident.title}, severity ${incident.severity}`}
                  >
                    <div className="flex items-start gap-3">
                      <span
                        className={cn(
                          'w-2 h-2 rounded-full mt-1.5 flex-shrink-0',
                          getSeverityDot(incident.severity)
                        )}
                        aria-hidden="true"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-0.5">
                          <h3 className="text-sm font-medium text-foreground truncate">
                            {incident.title}
                          </h3>
                        </div>
                        <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                          <span className="font-mono">{incident.source_ip}</span>
                          <span aria-hidden="true">&middot;</span>
                          <time dateTime={incident.start_time}>
                            {formatTimestamp(incident.start_time)}
                          </time>
                        </div>
                        <div className="flex items-center gap-2 mt-1">
                          <span className={cn(
                            'text-[10px] px-1.5 py-0.5 rounded-full capitalize',
                            incident.severity === 'critical' && 'bg-severity-critical/15 text-severity-critical',
                            incident.severity === 'high' && 'bg-severity-high/15 text-severity-high',
                            incident.severity === 'medium' && 'bg-severity-medium/15 text-severity-medium',
                            incident.severity === 'low' && 'bg-severity-low/15 text-severity-low'
                          )}>
                            {incident.severity}
                          </span>
                          <span className="text-[10px] text-muted-foreground">
                            {incident.event_count} events
                          </span>
                        </div>
                      </div>
                    </div>
                  </a>
                ))
              ) : (
                <div className="px-4 py-8 text-center">
                  <Shield className="w-8 h-8 text-muted-foreground mx-auto mb-2 opacity-50" aria-hidden="true" />
                  <p className="text-sm text-muted-foreground">No incidents found</p>
                  {(severity || status || sourceIP) && (
                    <button
                      onClick={() => router.push('/')}
                      className="text-xs text-primary hover:underline mt-1"
                    >
                      Clear filters
                    </button>
                  )}
                </div>
              )}
            </div>

            {/* Sidebar pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between px-4 py-2 border-t border-border flex-shrink-0">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page <= 1}
                  className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                  aria-label="Previous page"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <span className="text-[11px] text-muted-foreground">
                  {page}/{totalPages}
                </span>
                <button
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page >= totalPages}
                  className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                  aria-label="Next page"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            )}
          </div>
        </aside>

        {/* Sidebar toggle (visible when collapsed) */}
        {!sidebarOpen && (
          <button
            onClick={() => setSidebarOpen(true)}
            className="absolute left-0 top-1/2 -translate-y-1/2 z-30 bg-card border border-border border-l-0 rounded-r-md p-2 text-muted-foreground hover:text-foreground transition-colors max-lg:hidden"
            aria-label="Open sidebar"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        )}

        {/* Mobile sidebar overlay */}
        {sidebarOpen && (
          <div
            className="fixed inset-0 bg-overlay z-10 lg:hidden"
            onClick={() => setSidebarOpen(false)}
            aria-hidden="true"
          />
        )}

        {/* Main content area */}
        <main className="flex-1 overflow-y-auto" id="main-content">
          <div className="max-w-6xl mx-auto p-6 space-y-6">
            {/* Error banner */}
            {error && (
              <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4" role="alert">
                <p className="text-destructive text-sm">Failed to load incidents: {error}</p>
              </div>
            )}

            {/* Stats cards with sparklines */}
            {stats && (
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4" role="region" aria-label="Incident statistics">
                {statCards.map((card) => card && (
                  <div
                    key={card.label}
                    className="bg-card border border-border rounded-xl p-4 relative overflow-hidden group hover:border-border/80 transition-colors"
                  >
                    {/* Sparkline background */}
                    <div className={cn('absolute inset-0 opacity-[0.03] group-hover:opacity-[0.06] transition-opacity')}>
                      <svg className="w-full h-full" viewBox="0 0 100 40" preserveAspectRatio="none">
                        <polyline
                          points="0,35 15,30 25,32 35,20 50,22 60,15 70,18 85,10 100,12"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth="2"
                          className={card.sparkColor}
                        />
                      </svg>
                    </div>
                    <div className="relative">
                      <div className={cn('text-2xl font-bold', card.color)}>
                        {card.value.toLocaleString()}
                      </div>
                      <div className="flex items-center gap-2 mt-1">
                        <span className="text-xs text-muted-foreground">{card.label}</span>
                        {card.trend && (
                          <span className={cn(
                            'text-[10px] font-medium',
                            card.trend.startsWith('+') ? 'text-severity-low' : 'text-severity-critical'
                          )}>
                            {card.trend}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* TechniqueHeatmap */}
            <TechniqueHeatmap />

            {/* LiveEventStream */}
            <LiveEventStream />

            {/* DashboardControls - kept for advanced filtering */}
            <DashboardControls
              ref={dashboardControlsRef}
              currentPage={page}
              totalPages={totalPages}
              total={total}
              currentSeverity={severity}
              currentStatus={status}
              currentSourceIP={sourceIP}
            />
          </div>
        </main>
      </div>

      <KeyboardShortcutsModal
        isOpen={showShortcuts}
        onClose={() => setShowShortcuts(false)}
      />
    </div>
  );
}
