'use client';

import { useEffect, useState, useRef, useCallback } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { API_BASE, isAuthenticated, clearToken } from '@/lib/api';
import { IncidentCard } from '@/components/IncidentCard';
import { DashboardControls, type DashboardControlsHandle } from '@/components/DashboardControls';
import { KeyboardShortcutsModal } from '@/components/KeyboardShortcutsModal';
import { Shield, Keyboard } from 'lucide-react';
import { TechniqueHeatmap } from '@/components/TechniqueHeatmap';
import type { Incident, IncidentStats } from '@/lib/types';

function getHeaders() {
  const token = localStorage.getItem('nexus_token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return headers;
}

export default function HomePage() {
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

  // Sync filter state from URL search params (for browser back/forward)
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

  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [showShortcuts, setShowShortcuts] = useState(false);
  const [pendingKey, setPendingKey] = useState<string | null>(null);
  const dashboardControlsRef = useRef<DashboardControlsHandle>(null);

  useEffect(() => {
    const loadData = async () => {
      const headers = getHeaders();
      try {
        const params = new URLSearchParams();
        params.set('limit', String(limit));
        params.set('offset', String((page - 1) * limit));
        if (severity) params.set('severity', severity);
        if (status) params.set('status', status);
        if (sourceIP) params.set('source_ip', sourceIP);

        const [incRes, statsRes] = await Promise.all([
          fetch(`${API_BASE}/api/incidents?${params.toString()}`, { headers }),
          fetch(`${API_BASE}/api/incidents/stats`, { headers }),
        ]);

        if (incRes.status === 401 || statsRes.status === 401) {
          clearToken();
          router.push('/login');
          return;
        }

        if (incRes.ok) {
          const data = await incRes.json();
          setIncidents(data.incidents || []);
          setTotal(data.total);
        }
        if (statsRes.ok) {
          setStats(await statsRes.json());
        }
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load');
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [page, severity, status, sourceIP]);

  // Reset selected index when incidents change
  useEffect(() => {
    setSelectedIndex(-1);
  }, [incidents]);

  // Keyboard shortcuts handler
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Don't handle shortcuts when typing in input fields
      const target = e.target as HTMLElement;
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable
      ) {
        // Allow Escape to blur input fields
        if (e.key === 'Escape') {
          (target as HTMLElement).blur();
          dashboardControlsRef.current?.clearSearch();
        }
        return;
      }

      // Handle modal open state
      if (showShortcuts) {
        if (e.key === 'Escape' || e.key === '?') {
          setShowShortcuts(false);
          e.preventDefault();
        }
        return;
      }

      // Handle pending g key for go-to shortcuts
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

      // Single key shortcuts
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
    return (
      <main className="min-h-screen bg-zinc-950 flex items-center justify-center">
        <div className="text-zinc-400" role="status">Loading dashboard...</div>
      </main>
    );
  }

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="min-h-screen bg-zinc-950 p-6">
      <div className="max-w-7xl mx-auto" id="main-content">
        <header className="mb-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Shield className="w-8 h-8 text-blue-500" aria-hidden="true" />
              <div>
                <h1 className="text-3xl font-bold text-zinc-100">Nexus</h1>
                <p className="text-zinc-400">Security Incident Dashboard</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={() => setShowShortcuts(true)}
                className="p-2 text-zinc-500 hover:text-zinc-300 transition-colors rounded"
                aria-label="Show keyboard shortcuts"
                title="Keyboard shortcuts (?)"
              >
                <Keyboard className="w-5 h-5" />
              </button>
              {isAuthenticated() ? (
                <>
                  <Link href="/settings" className="px-3 py-1.5 bg-zinc-800 text-zinc-300 rounded hover:bg-zinc-700 transition-colors text-sm">
                    Settings
                  </Link>
                  <button onClick={handleLogout} className="px-3 py-1.5 bg-zinc-800 text-zinc-300 rounded hover:bg-zinc-700 transition-colors text-sm">
                    Sign Out
                  </button>
                </>
              ) : (
                <>
                  <Link href="/login" className="px-3 py-1.5 bg-zinc-800 text-zinc-300 rounded hover:bg-zinc-700 transition-colors text-sm">
                    Sign In
                  </Link>
                  <Link href="/signup" className="px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-500 transition-colors text-sm">
                    Sign Up
                  </Link>
                </>
              )}
            </div>
          </div>
        </header>

        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8" role="region" aria-label="Incident statistics">
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
              <div className="text-2xl font-bold text-zinc-100">{stats.total_incidents.toLocaleString()}</div>
              <div className="text-sm text-zinc-500">Total Incidents</div>
            </div>
            {stats.by_severity?.critical && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-red-400">{stats.by_severity.critical.toLocaleString()}</div>
                <div className="text-sm text-zinc-500">Critical</div>
              </div>
            )}
            {stats.by_severity?.high && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-orange-400">{stats.by_severity.high.toLocaleString()}</div>
                <div className="text-sm text-zinc-500">High</div>
              </div>
            )}
            {stats.by_severity?.medium && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-yellow-400">{stats.by_severity.medium.toLocaleString()}</div>
                <div className="text-sm text-zinc-500">Medium</div>
              </div>
            )}
          </div>
        )}

        <div className="mb-6">
          <TechniqueHeatmap />
        </div>

        {error && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 mb-6" role="alert">
            <p className="text-red-400">Failed to load incidents: {error}</p>
          </div>
        )}

        <DashboardControls
          ref={dashboardControlsRef}
          currentPage={page}
          totalPages={totalPages}
          total={total}
          currentSeverity={severity}
          currentStatus={status}
          currentSourceIP={sourceIP}
        />

        {incidents.length > 0 ? (
          <>
            <div aria-live="polite" className="sr-only">
              {total} incidents loaded, page {page} of {totalPages}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" role="list" aria-label="Incidents">
            {incidents.map((incident, index) => (
              <IncidentCard
                key={incident.id}
                incident={incident}
                selected={index === selectedIndex}
              />
            ))}
          </div>
          </>
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" aria-hidden="true" />
            <h2 className="text-lg font-medium text-zinc-300 mb-2">No incidents found</h2>
            <p className="text-zinc-500">
              {severity || status || sourceIP ? 'Try adjusting your filters.' : 'Import data to get started.'}
            </p>
          </div>
        )}

        {totalPages > 1 && (
          <nav className="flex items-center justify-center gap-2 mt-8" aria-label="Pagination">
            {page > 1 && (
              <button onClick={() => setPage(page - 1)} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm">
                Previous
              </button>
            )}
            <span className="text-sm text-zinc-500 px-4">
              Page {page} of {totalPages} ({total.toLocaleString()} incidents)
            </span>
            {page < totalPages && (
              <button onClick={() => setPage(page + 1)} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm">
                Next
              </button>
            )}
          </nav>
        )}
      </div>

      <KeyboardShortcutsModal
        isOpen={showShortcuts}
        onClose={() => setShowShortcuts(false)}
      />
    </div>
  );
}
