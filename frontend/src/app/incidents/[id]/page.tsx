'use client';

import { useEffect, useState, useCallback } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { StoryCard } from '@/components/StoryCard';
import { SeverityBadge } from '@/components/SeverityBadge';
import { TechniqueBadge } from '@/components/TechniqueBadge';
import { GenerateNarrativeButton } from '@/components/GenerateNarrativeButton';
import { TimelineView } from '@/components/TimelineView';
import { ArrowLeft, Shield, Clock, ChevronDown, ChevronUp } from 'lucide-react';
import { KillChain } from '@/components/KillChain';
import { fetchIncident, fetchNarrative, getFeedback } from '@/lib/api';
import type { Incident, Narrative, Feedback } from '@/lib/types';

export default function IncidentDetailPage() {
  const params = useParams();
  const incidentId = Number(params.id);

  const [incident, setIncident] = useState<Incident | null>(null);
  const [narrative, setNarrative] = useState<Narrative | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeView, setActiveView] = useState<'narrative' | 'timeline'>('narrative');
  const [timelineOpen, setTimelineOpen] = useState(false);

  const loadNarrative = useCallback(async () => {
    try {
      const narRes = await fetchNarrative(incidentId);
      if (narRes.narrative) {
        setNarrative(narRes.narrative);
        const fbRes = await getFeedback(narRes.narrative.id);
        setFeedback(fbRes.feedback);
      }
    } catch {
      // narrative may not exist yet
    }
  }, [incidentId]);

  useEffect(() => {
    const load = async () => {
      try {
        const incRes = await fetchIncident(incidentId);
        setIncident(incRes.incident);
        await loadNarrative();
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Failed to load');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [incidentId, loadNarrative]);

  if (loading) {
    return (
      <main className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-muted-foreground" role="status">Loading incident...</div>
      </main>
    );
  }

  if (!incident && error) {
    return (
      <div className="min-h-screen bg-background p-6">
        <div className="max-w-7xl mx-auto" id="main-content">
          <Link href="/" className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6">
            <ArrowLeft className="w-4 h-4" aria-hidden="true" />
            Back to Incidents
          </Link>
          <div className="bg-card border border-border rounded-xl p-12 text-center" role="alert">
            <Shield className="w-12 h-12 text-muted-foreground/60 mx-auto mb-4" aria-hidden="true" />
            <h2 className="text-xl font-medium text-foreground/80 mb-2">Incident not found</h2>
            <p className="text-muted-foreground">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (!incident) return null;

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto" id="main-content">
        <Link href="/" className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6">
          <ArrowLeft className="w-4 h-4" aria-hidden="true" />
          Back to Incidents
        </Link>

        <header className="mb-6">
          <div className="flex items-center gap-3 mb-2 flex-wrap">
            <h1 className="text-2xl font-bold text-foreground">{incident.title}</h1>
            <SeverityBadge severity={incident.severity} />
            <span className="text-sm font-mono text-muted-foreground">#{incident.id}</span>
          </div>
          <p className="text-muted-foreground">{incident.description}</p>
        </header>

        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6" role="region" aria-label="Incident metadata">
          <div className="bg-card border border-border rounded-lg p-3">
            <div className="text-xs text-muted-foreground mb-1">Source IP</div>
            <div className="font-mono text-sm text-foreground/80">{incident.source_ip}</div>
          </div>
          <div className="bg-card border border-border rounded-lg p-3">
            <div className="text-xs text-muted-foreground mb-1">Time Range</div>
            <div className="text-sm text-foreground/80">
              {incident.start_time ? new Date(incident.start_time).toLocaleString() : '-'}
            </div>
          </div>
          <div className="bg-card border border-border rounded-lg p-3">
            <div className="text-xs text-muted-foreground mb-1">Events</div>
            <div className="text-sm text-foreground/80">{incident.event_count}</div>
          </div>
          <div className="bg-card border border-border rounded-lg p-3">
            <div className="text-xs text-muted-foreground mb-1">Unique Users</div>
            <div className="text-sm text-foreground/80">{incident.unique_users?.length || 0}</div>
          </div>
          <div className="bg-card border border-border rounded-lg p-3">
            <div className="text-xs text-muted-foreground mb-1">Unique IPs</div>
            <div className="text-sm text-foreground/80">{incident.unique_ips?.length || 0}</div>
          </div>
        </div>

        {incident.techniques && incident.techniques.length > 0 && (
          <section className="mb-6">
            <h2 className="text-sm font-medium text-muted-foreground mb-3">MITRE ATT&CK Techniques</h2>
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

        {incident.techniques && incident.techniques.length > 0 && (
          <section className="mb-6">
            <h2 className="text-sm font-medium text-muted-foreground mb-3">Kill Chain</h2>
            <div className="bg-card border border-border rounded-xl p-4">
              <KillChain techniques={incident.techniques} />
            </div>
          </section>
        )}

        <div className="flex items-center gap-2 mb-6" role="tablist" aria-label="Incident view">
          <button
            onClick={() => setActiveView('narrative')}
            className={activeView === 'narrative'
              ? 'px-4 py-2 text-sm font-medium rounded-lg bg-surface text-foreground border border-border'
              : 'px-4 py-2 text-sm font-medium rounded-lg text-muted-foreground hover:text-foreground/80 hover:bg-card transition-colors'
            }
            role="tab"
            aria-selected={activeView === 'narrative'}
            aria-controls="narrative-panel"
          >
            Narrative
          </button>
          <button
            onClick={() => setActiveView('timeline')}
            className={activeView === 'timeline'
              ? 'px-4 py-2 text-sm font-medium rounded-lg bg-surface text-foreground border border-border'
              : 'px-4 py-2 text-sm font-medium rounded-lg text-muted-foreground hover:text-foreground/80 hover:bg-card transition-colors'
            }
            role="tab"
            aria-selected={activeView === 'timeline'}
            aria-controls="timeline-panel"
          >
            <span className="flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5" aria-hidden="true" />
              Timeline
            </span>
          </button>
        </div>

        {activeView === 'narrative' && (
          <div id="narrative-panel" role="tabpanel" aria-label="Narrative view">
            {narrative ? (
              <StoryCard narrative={narrative} incidentId={incidentId} existingFeedback={feedback} />
            ) : (
              <div className="bg-card border border-border rounded-xl p-12 text-center">
                <Shield className="w-12 h-12 text-muted-foreground/60 mx-auto mb-4" aria-hidden="true" />
                <h3 className="text-lg font-medium text-foreground/80 mb-2">No narrative generated</h3>
                <p className="text-muted-foreground mb-6">
                  Generate an AI narrative for this incident to see the attack story.
                </p>
                <GenerateNarrativeButton incidentId={incidentId} onGenerated={loadNarrative} />
              </div>
            )}
          </div>
        )}

        {activeView === 'timeline' && (
          <div id="timeline-panel" role="tabpanel" aria-label="Timeline view">
            <div className="bg-card border border-border rounded-xl overflow-hidden">
              <button
                onClick={() => setTimelineOpen(!timelineOpen)}
                className="w-full flex items-center justify-between px-6 py-4 hover:bg-surface/50 transition-colors"
                aria-expanded={timelineOpen}
                aria-controls="timeline-content"
              >
                <div className="flex items-center gap-3">
                  <Clock className="w-5 h-5 text-muted-foreground" aria-hidden="true" />
                  <h3 className="text-lg font-semibold text-foreground">Event Timeline</h3>
                </div>
                {timelineOpen ? (
                  <ChevronUp className="w-5 h-5 text-muted-foreground" aria-hidden="true" />
                ) : (
                  <ChevronDown className="w-5 h-5 text-muted-foreground" aria-hidden="true" />
                )}
              </button>
              {timelineOpen && (
                <div id="timeline-content" className="px-6 pb-6">
                  <TimelineView incidentId={incidentId} />
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
