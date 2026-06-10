import { fetchIncidents, fetchIncidentStats } from '@/lib/api';
import { IncidentCard } from '@/components/IncidentCard';
import { Shield } from 'lucide-react';

export default async function HomePage() {
  let incidents: Awaited<ReturnType<typeof fetchIncidents>>['incidents'] = [];
  let total = 0;
  let stats: Awaited<ReturnType<typeof fetchIncidentStats>> | null = null;
  let error: string | null = null;

  try {
    const data = await fetchIncidents(50);
    incidents = data.incidents || [];
    total = data.total;
  } catch (e) {
    error = e instanceof Error ? e.message : 'Failed to load incidents';
  }

  try {
    stats = await fetchIncidentStats();
  } catch {
    // Stats are optional
  }

  return (
    <div className="min-h-screen bg-zinc-950 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Shield className="w-8 h-8 text-blue-500" />
            <h1 className="text-3xl font-bold text-zinc-100">Nexus</h1>
          </div>
          <p className="text-zinc-400">Security Incident Dashboard</p>
        </div>

        {/* Stats bar */}
        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
              <div className="text-2xl font-bold text-zinc-100">
                {stats.total_incidents}
              </div>
              <div className="text-sm text-zinc-500">Total Incidents</div>
            </div>
            {stats.by_severity?.critical && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-red-400">
                  {stats.by_severity.critical}
                </div>
                <div className="text-sm text-zinc-500">Critical</div>
              </div>
            )}
            {stats.by_severity?.high && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-orange-400">
                  {stats.by_severity.high}
                </div>
                <div className="text-sm text-zinc-500">High</div>
              </div>
            )}
            {stats.by_severity?.medium && (
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                <div className="text-2xl font-bold text-yellow-400">
                  {stats.by_severity.medium}
                </div>
                <div className="text-sm text-zinc-500">Medium</div>
              </div>
            )}
          </div>
        )}

        {/* Error state */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 mb-6">
            <p className="text-red-400">Failed to load incidents: {error}</p>
            <p className="text-sm text-zinc-500 mt-2">
              Make sure the backend is running on port 8080
            </p>
          </div>
        )}

        {/* Incident count */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-zinc-100">
            Incidents ({total})
          </h2>
        </div>

        {/* Incidents grid */}
        {incidents.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {incidents.map((incident) => (
              <IncidentCard key={incident.id} incident={incident} />
            ))}
          </div>
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-zinc-300 mb-2">
              No incidents found
            </h3>
            <p className="text-zinc-500 mb-4">
              Import data from the backend to get started.
            </p>
            <code className="text-xs text-zinc-600 bg-zinc-800 px-3 py-1 rounded">
              curl -X POST http://localhost:8080/api/import
            </code>
          </div>
        )}
      </div>
    </div>
  );
}
