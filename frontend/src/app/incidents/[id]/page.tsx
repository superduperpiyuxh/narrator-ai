import { fetchIncident, fetchNarrative } from '@/lib/api';
import { StoryCard } from '@/components/StoryCard';
import { SeverityBadge } from '@/components/SeverityBadge';
import { TechniqueBadge } from '@/components/TechniqueBadge';
import { ArrowLeft, Shield } from 'lucide-react';
import Link from 'next/link';

export default async function IncidentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const incidentId = Number(id);

  let incident = null;
  let narrative = null;
  let error: string | null = null;

  try {
    const data = await fetchIncident(incidentId);
    incident = data.incident;
  } catch (e) {
    error = e instanceof Error ? e.message : 'Failed to load incident';
  }

  if (incident) {
    try {
      const data = await fetchNarrative(incidentId);
      narrative = data.narrative;
    } catch {
      // Narrative might not exist yet
    }
  }

  if (!incident && error) {
    return (
      <div className="min-h-screen bg-zinc-950 p-6">
        <div className="max-w-7xl mx-auto">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-zinc-400 hover:text-zinc-200 mb-6"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to Incidents
          </Link>
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h2 className="text-xl font-medium text-zinc-300 mb-2">
              Incident not found
            </h2>
            <p className="text-zinc-500">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-zinc-950 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Back link */}
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-zinc-400 hover:text-zinc-200 mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Incidents
        </Link>

        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center gap-3 mb-2">
            <h1 className="text-2xl font-bold text-zinc-100">{incident?.title}</h1>
            <SeverityBadge severity={incident?.severity || ''} />
            <span className="text-sm font-mono text-zinc-500">#{incident?.id}</span>
          </div>
          <p className="text-zinc-400">{incident?.description}</p>
        </div>

        {/* Metadata row */}
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Source IP</div>
            <div className="font-mono text-sm text-zinc-300">{incident?.source_ip}</div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Time Range</div>
            <div className="text-sm text-zinc-300">
              {incident?.start_time ? new Date(incident.start_time).toLocaleString() : '-'}
            </div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Events</div>
            <div className="text-sm text-zinc-300">{incident?.event_count}</div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Unique Users</div>
            <div className="text-sm text-zinc-300">{incident?.unique_users?.length || 0}</div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Unique IPs</div>
            <div className="text-sm text-zinc-300">{incident?.unique_ips?.length || 0}</div>
          </div>
        </div>

        {/* MITRE Techniques */}
        {incident?.techniques && incident.techniques.length > 0 && (
          <div className="mb-6">
            <h2 className="text-sm font-medium text-zinc-400 mb-3">MITRE ATT&CK Techniques</h2>
            <div className="flex flex-wrap gap-2">
              {incident.techniques.map((tech) => (
                <TechniqueBadge
                  key={tech.technique_id}
                  techniqueId={tech.technique_id}
                  name={`${tech.name} (${tech.event_count})`}
                />
              ))}
            </div>
          </div>
        )}

        {/* Story Card */}
        {narrative ? (
          <StoryCard narrative={narrative} incidentId={incidentId} />
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-zinc-300 mb-2">
              No narrative generated
            </h3>
            <p className="text-zinc-500 mb-4">
              Generate an AI narrative for this incident to see the attack story.
            </p>
            <code className="text-xs text-zinc-600 bg-zinc-800 px-3 py-1 rounded">
              curl -X POST http://localhost:8080/api/incidents/{incidentId}/narrative
            </code>
          </div>
        )}
      </div>
    </div>
  );
}
