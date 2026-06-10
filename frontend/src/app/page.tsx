'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { isAuthenticated, clearToken } from '@/lib/api';
import { IncidentCard } from '@/components/IncidentCard';
import { DashboardControls } from '@/components/DashboardControls';
import { Shield } from 'lucide-react';
import type { Incident, IncidentStats } from '@/lib/types';

export default function HomePage() {
  const router = useRouter();
  const [checked, setChecked] = useState(false);
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [total, setTotal] = useState(0);
  const [stats, setStats] = useState<IncidentStats | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [severity, setSeverity] = useState('');
  const [status, setStatus] = useState('');
  const [sourceIP, setSourceIP] = useState('');
  const limit = 24;

  useEffect(() => {
    // Demo mode: no auth required, just show dashboard
    setChecked(true);
  }, [router]);

  useEffect(() => {
    if (!checked) return;

    const loadData = async () => {
      try {
        const API_BASE = 'http://localhost:8080';
        const token = localStorage.getItem('nexus_token');
        const headers: Record<string, string> = { 'Content-Type': 'application/json' };
        if (token) headers['Authorization'] = `Bearer ${token}`;

        const params = new URLSearchParams();
        params.set('limit', String(limit));
        params.set('offset', String((page - 1) * limit));
        if (severity) params.set('severity', severity);
        if (status) params.set('status', status);
        if (sourceIP) params.set('source_ip', sourceIP);

        const incRes = await fetch(`${API_BASE}/api/incidents?${params.toString()}`, { headers });
        if (incRes.ok) {
          const data = await incRes.json();
          setIncidents(data.incidents || []);
          setTotal(data.total);
        }

        const statsRes = await fetch(`${API_BASE}/api/incidents/stats`, { headers });
        if (statsRes.ok) {
          setStats(await statsRes.json());
        }
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load');
      }
    };
    loadData();
  }, [checked, page, severity, status, sourceIP]);

  const handleLogout = () => {
    clearToken();
    localStorage.removeItem('nexus_user');
    router.push('/login');
  };

  if (!checked) {
    return (
      <main className="min-h-screen bg-gray-950 flex items-center justify-center">
        <div className="text-gray-400" role="status">Loading...</div>
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
              {isAuthenticated() ? (
                <>
                  <Link
                    href="/settings"
                    className="px-3 py-1.5 bg-zinc-800 text-zinc-300 rounded hover:bg-zinc-700 transition-colors text-sm"
                  >
                    Settings
                  </Link>
                  <button
                    onClick={handleLogout}
                    className="px-3 py-1.5 bg-zinc-800 text-zinc-300 rounded hover:bg-zinc-700 transition-colors text-sm"
                  >
                    Sign Out
                  </button>
                </>
              ) : (
                <>
                  <Link
                    href="/login"
                    className="px-3 py-1.5 bg-zinc-800 text-zinc-300 rounded hover:bg-zinc-700 transition-colors text-sm"
                  >
                    Sign In
                  </Link>
                  <Link
                    href="/signup"
                    className="px-3 py-1.5 bg-blue-600 text-white rounded hover:bg-blue-500 transition-colors text-sm"
                  >
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
              <div className="text-2xl font-bold text-zinc-100">
                {stats.total_incidents.toLocaleString()}
              </div>
              <div className="text-sm text-zinc-500">Total Incidents</div>
            </div>
            {stats.by_severity?.critical && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-red-400">
                  {stats.by_severity.critical.toLocaleString()}
                </div>
                <div className="text-sm text-zinc-500">Critical</div>
              </div>
            )}
            {stats.by_severity?.high && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-orange-400">
                  {stats.by_severity.high.toLocaleString()}
                </div>
                <div className="text-sm text-zinc-500">High</div>
              </div>
            )}
            {stats.by_severity?.medium && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-yellow-400">
                  {stats.by_severity.medium.toLocaleString()}
                </div>
                <div className="text-sm text-zinc-500">Medium</div>
              </div>
            )}
          </div>
        )}

        {error && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 mb-6" role="alert">
            <p className="text-red-400">Failed to load incidents: {error}</p>
          </div>
        )}

        <DashboardControls
          currentPage={page}
          totalPages={totalPages}
          total={total}
          currentSeverity={severity}
          currentStatus={status}
          currentSourceIP={sourceIP}
        />

        {incidents.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" role="list" aria-label="Incidents">
            {incidents.map((incident) => (
              <IncidentCard key={incident.id} incident={incident} />
            ))}
          </div>
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" aria-hidden="true" />
            <h2 className="text-lg font-medium text-zinc-300 mb-2">No incidents found</h2>
            <p className="text-zinc-500 mb-4">
              {severity || status || sourceIP
                ? 'Try adjusting your filters.'
                : 'Import data from the backend to get started.'}
            </p>
          </div>
        )}

        {totalPages > 1 && (
          <nav className="flex items-center justify-center gap-2 mt-8" aria-label="Pagination">
            {page > 1 && (
              <button
                onClick={() => setPage(page - 1)}
                className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm"
              >
                Previous
              </button>
            )}
            <span className="text-sm text-zinc-500 px-4">
              Page {page} of {totalPages} ({total.toLocaleString()} incidents)
            </span>
            {page < totalPages && (
              <button
                onClick={() => setPage(page + 1)}
                className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm"
              >
                Next
              </button>
            )}
          </nav>
        )}
      </div>
    </div>
  );
}
