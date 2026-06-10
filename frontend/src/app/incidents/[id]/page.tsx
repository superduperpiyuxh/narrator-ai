'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { StoryCard } from '@/components/StoryCard';
import { SeverityBadge } from '@/components/SeverityBadge';
import { TechniqueBadge } from '@/components/TechniqueBadge';
import { GenerateNarrativeButton } from '@/components/GenerateNarrativeButton';
import { ArrowLeft, Shield } from 'lucide-react';
import type { Incident, Narrative, Feedback } from '@/lib/types';

const API_BASE = 'http://localhost:8080';

function getHeaders() {
  const token = localStorage.getItem('nexus_token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return headers;
}

export default function IncidentDetailPage() {
  const params = useParams();
  const incidentId = Number(params.id);

  const [incident, setIncident] = useState<Incident | null>(null);
  const [narrative, setNarrative] = useState<Narrative | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      const headers = getHeaders();
      try {
        const incRes = await fetch(`${API_BASE}/api/incidents/${incidentId}`, { headers });
        if (!incRes.ok) throw new Error('Failed to load incident');
        const incData = await incRes.json();
        setIncident(incData.incident);

        const narRes = await fetch(`${API_BASE}/api/incidents/${incidentId}/narrative`, { headers });
        if (narRes.ok) {
          const narData = await narRes.json();
          if (narData.narrative) {
            setNarrative(narData.narrative);
            const fbRes = await fetch(`${API_BASE}/api/feedback/${narData.narrative.id}`, { headers });
            if (fbRes.ok) {
              const fbData = await fbRes.json();
              setFeedback(fbData.feedback);
            }
          }
        }
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [incidentId]);

  if (loading) {
    return (
      <main className="min-h-screen bg-zinc-950 flex items-center justify-center">
        <div className="text-zinc-400" role="status">Loading incident...</div>
      </main>
    );
  }

  if (!incident && error) {
    return (
      <div className="min-h-screen bg-zinc-950 p-6">
        <div className="max-w-7xl mx-auto" id="main-content">
          <Link href="/" className="inline-flex items-center gap-2 text-zinc-400 hover:text-zinc-200 mb-6">
            <ArrowLeft className="w-4 h-4" aria-hidden="true" />
            Back to Incidents
          </Link>
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center" role="alert">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" aria-hidden="true" />
            <h2 className="text-xl font-medium text-zinc-300 mb-2">Incident not found</h2>
            <p className="text-zinc-500">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (!incident) return null;

  return (
    <div className="min-h-screen bg-zinc-950 p-6">
      <div className="max-w-7xl mx-auto" id="main-content">
        <Link href="/" className="inline-flex items-center gap-2 text-zinc-400 hover:text-zinc-200 mb-6">
          <ArrowLeft className="w-4 h-4" aria-hidden="true" />
          Back to Incidents
        </Link>

        <header className="mb-6">
          <div className="flex items-center gap-3 mb-2 flex-wrap">
            <h1 className="text-2xl font-bold text-zinc-100">{incident.title}</h1>
            <SeverityBadge severity={incident.severity} />
            <span className="text-sm font-mono text-zinc-500">#{incident.id}</span>
          </div>
          <p className="text-zinc-400">{incident.description}</p>
        </header>

        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6" role="region" aria-label="Incident metadata">
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Source IP</div>
            <div className="font-mono text-sm text-zinc-300">{incident.source_ip}</div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Time Range</div>
            <div className="text-sm text-zinc-300">
              {incident.start_time ? new Date(incident.start_time).toLocaleString() : '-'}
            </div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Events</div>
            <div className="text-sm text-zinc-300">{incident.event_count}</div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Unique Users</div>
            <div className="text-sm text-zinc-300">{incident.unique_users?.length || 0}</div>
          </div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-3">
            <div className="text-xs text-zinc-500 mb-1">Unique IPs</div>
            <div className="text-sm text-zinc-300">{incident.unique_ips?.length || 0}</div>
          </div>
        </div>

        {incident.techniques && incident.techniques.length > 0 && (
          <section className="mb-6">
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
          </section>
        )}

        {narrative ? (
          <StoryCard narrative={narrative} incidentId={incidentId} existingFeedback={feedback} />
        ) : (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center">
            <Shield className="w-12 h-12 text-zinc-600 mx-auto mb-4" aria-hidden="true" />
            <h3 className="text-lg font-medium text-zinc-300 mb-2">No narrative generated</h3>
            <p className="text-zinc-500 mb-6">
              Generate an AI narrative for this incident to see the attack story.
            </p>
            <GenerateNarrativeButton incidentId={incidentId} />
          </div>
        )}
      </div>
    </div>
  );
}
