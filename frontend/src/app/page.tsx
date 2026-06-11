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
import { IncidentCard } from '@/components/IncidentCard';
import { Shield, Keyboard, Search, X, ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
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
  const dashboardControlsRef = useRef<DashboardControlsHandle>(null);
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

  const handleLogout = () => {
    clearToken();
    window.location.reload();
  };

  if (loading) {
    return <LoadingSkeleton />;
  }

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="min-h-screen bg-background">
      <CommandPalette isOpen={isPaletteOpen} onClose={closePalette} />

      {/* Header */}
      <header className="h-14 border-b border-border bg-card flex items-center justify-between px-4 flex-shrink-0 sticky top-0 z-20">
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

      {/* Main content — full width vertical flow */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-6 space-y-6" id="main-content">
        {/* Error banner */}
        {error && (
          <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4" role="alert">
            <p className="text-destructive text-sm">Failed to load incidents: {error}</p>
          </div>
        )}

        {/* Stats cards */}
        {stats && (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4" role="region" aria-label="Incident statistics">
            <StatCard
              label="Total Incidents"
              value={stats.total_incidents}
              color="text-foreground"
              sparkColor="text-muted-foreground"
            />
            {stats.by_severity?.critical != null && (
              <StatCard
                label="Critical"
                value={stats.by_severity.critical}
                color="text-severity-critical"
                sparkColor="text-severity-critical/40"
              />
            )}
            {stats.by_severity?.high != null && (
              <StatCard
                label="High"
                value={stats.by_severity.high}
                color="text-severity-high"
                sparkColor="text-severity-high/40"
              />
            )}
            {stats.by_severity?.medium != null && (
              <StatCard
                label="Medium"
                value={stats.by_severity.medium}
                color="text-severity-medium"
                sparkColor="text-severity-medium/40"
              />
            )}
          </div>
        )}

        {/* Two-column: Heatmap + Live Stream */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            <TechniqueHeatmap />
          </div>
          <div className="lg:col-span-1">
            <LiveEventStream />
          </div>
        </div>

        {/* Filter bar */}
        <div className="bg-card border border-border rounded-xl p-4">
          <div className="flex flex-wrap items-center gap-3">
            <span className="text-sm font-medium text-foreground">Filters</span>
            <div className="h-4 w-px bg-border" aria-hidden="true" />
            <select
              value={severity}
              onChange={(e) => {
                const params = new URLSearchParams();
                if (e.target.value) params.set('severity', e.target.value);
                if (status) params.set('status', status);
                if (sourceIP) params.set('source_ip', sourceIP);
                router.push(`/?${params.toString()}`);
              }}
              className="bg-surface border border-border rounded-md px-2.5 py-1.5 text-xs text-foreground focus:outline-none focus:border-primary transition-colors"
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
              className="bg-surface border border-border rounded-md px-2.5 py-1.5 text-xs text-foreground focus:outline-none focus:border-primary transition-colors"
              aria-label="Filter by status"
            >
              <option value="">All Status</option>
              <option value="open">Open</option>
              <option value="investigating">Investigating</option>
              <option value="resolved">Resolved</option>
            </select>
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
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
                placeholder="Source IP..."
                className="bg-surface border border-border rounded-md pl-8 pr-8 py-1.5 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-primary transition-colors w-40"
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
                className="text-xs text-primary hover:underline transition-colors"
              >
                Clear all
              </button>
            )}
            <div className="ml-auto text-xs text-muted-foreground">
              {total.toLocaleString()} incidents
            </div>
          </div>
        </div>

        {/* Incident grid */}
        {incidents.length > 0 ? (
          <>
            <div
              className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
              role="list"
              aria-label="Incidents"
            >
              {incidents.map((incident, index) => (
                <IncidentCard
                  key={incident.id}
                  incident={incident}
                  selected={index === selectedIndex}
                />
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <nav className="flex items-center justify-center gap-3 pt-2" aria-label="Pagination">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page <= 1}
                  className="flex items-center gap-1 px-3 py-1.5 bg-card border border-border text-sm text-muted-foreground rounded-lg hover:bg-surface-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                  aria-label="Previous page"
                >
                  <ChevronLeft className="w-4 h-4" />
                  Previous
                </button>
                <span className="text-sm text-muted-foreground">
                  Page {page} of {totalPages}
                </span>
                <button
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page >= totalPages}
                  className="flex items-center gap-1 px-3 py-1.5 bg-card border border-border text-sm text-muted-foreground rounded-lg hover:bg-surface-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                  aria-label="Next page"
                >
                  Next
                  <ChevronRight className="w-4 h-4" />
                </button>
              </nav>
            )}
          </>
        ) : (
          <div className="bg-card border border-border rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-muted-foreground mx-auto mb-4 opacity-50" aria-hidden="true" />
            <h2 className="text-lg font-medium text-foreground mb-2">No incidents found</h2>
            <p className="text-muted-foreground text-sm">
              {severity || status || sourceIP ? 'Try adjusting your filters.' : 'Import data to get started.'}
            </p>
            {(severity || status || sourceIP) && (
              <button
                onClick={() => router.push('/')}
                className="mt-3 text-sm text-primary hover:underline"
              >
                Clear filters
              </button>
            )}
          </div>
        )}
      </main>

      <KeyboardShortcutsModal
        isOpen={showShortcuts}
        onClose={() => setShowShortcuts(false)}
      />
    </div>
  );
}

function StatCard({
  label,
  value,
  color,
  sparkColor,
}: {
  label: string;
  value: number;
  color: string;
  sparkColor: string;
}) {
  return (
    <div className="bg-card border border-border rounded-xl p-4 relative overflow-hidden group hover:border-border/80 transition-colors">
      <div className={cn('absolute inset-0 opacity-[0.03] group-hover:opacity-[0.06] transition-opacity')}>
        <svg className="w-full h-full" viewBox="0 0 100 40" preserveAspectRatio="none">
          <polyline
            points="0,35 15,30 25,32 35,20 50,22 60,15 70,18 85,10 100,12"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            className={sparkColor}
          />
        </svg>
      </div>
      <div className="relative">
        <div className={cn('text-2xl font-bold', color)}>
          {value.toLocaleString()}
        </div>
        <div className="text-xs text-muted-foreground mt-1">{label}</div>
      </div>
    </div>
  );
}
