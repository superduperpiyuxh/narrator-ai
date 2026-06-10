import { fetchIncidents, fetchIncidentStats } from '@/lib/api';
import { IncidentCard } from '@/components/IncidentCard';
import { DashboardControls } from '@/components/DashboardControls';
import { Shield } from 'lucide-react';

export default async function HomePage({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const params = await searchParams;
  const page = Number(params.page) || 1;
  const severity = typeof params.severity === 'string' ? params.severity : '';
  const status = typeof params.status === 'string' ? params.status : '';
  const sourceIP = typeof params.source_ip === 'string' ? params.source_ip : '';
  const limit = 24;
  const offset = (page - 1) * limit;

  let incidents: Awaited<ReturnType<typeof fetchIncidents>>['incidents'] = [];
  let total = 0;
  let stats: Awaited<ReturnType<typeof fetchIncidentStats>> | null = null;
  let error: string | null = null;

  try {
    const data = await fetchIncidents(limit, offset, severity || undefined, status || undefined, sourceIP || undefined);
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

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="min-h-screen bg-zinc-950 p-6">
      <div className="max-w-7xl mx-auto" id="main-content">
        {/* Header */}
        <header className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Shield className="w-8 h-8 text-blue-500" aria-hidden="true" />
            <h1 className="text-3xl font-bold text-zinc-100">Nexus</h1>
          </div>
          <p className="text-zinc-400">Security Incident Dashboard</p>
        </header>

        {/* Stats bar */}
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

        {/* Error state */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 mb-6" role="alert">
            <p className="text-red-400">Failed to load incidents: {error}</p>
            <p className="text-sm text-zinc-500 mt-2">
              Make sure the backend is running on port 8080
            </p>
          </div>
        )}

        {/* Search and Filters */}
        <DashboardControls
          currentPage={page}
          totalPages={totalPages}
          total={total}
          currentSeverity={severity}
          currentStatus={status}
          currentSourceIP={sourceIP}
        />

        {/* Incidents grid */}
        {incidents.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" role="list" aria-label="Incidents">
            {incidents.map((incident) => (
              <IncidentCard key={incident.id} incident={incident} />
            ))}
          </div>
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" aria-hidden="true" />
            <h2 className="text-lg font-medium text-zinc-300 mb-2">
              No incidents found
            </h2>
            <p className="text-zinc-500 mb-4">
              {severity || status || sourceIP
                ? 'Try adjusting your filters.'
                : 'Import data from the backend to get started.'}
            </p>
            {!severity && !status && !sourceIP && (
              <code className="text-xs text-zinc-600 bg-zinc-800 px-3 py-1 rounded">
                curl -X POST http://localhost:8080/api/import
              </code>
            )}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <nav className="flex items-center justify-center gap-2 mt-8" aria-label="Pagination">
            {page > 1 && (
              <a
                href={`/?page=${page - 1}${severity ? `&severity=${severity}` : ''}${status ? `&status=${status}` : ''}${sourceIP ? `&source_ip=${sourceIP}` : ''}`}
                className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm"
              >
                Previous
              </a>
            )}
            <span className="text-sm text-zinc-500 px-4">
              Page {page} of {totalPages} ({total.toLocaleString()} incidents)
            </span>
            {page < totalPages && (
              <a
                href={`/?page=${page + 1}${severity ? `&severity=${severity}` : ''}${status ? `&status=${status}` : ''}${sourceIP ? `&source_ip=${sourceIP}` : ''}`}
                className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors text-sm"
              >
                Next
              </a>
            )}
          </nav>
        )}
      </div>
    </div>
  );
}
