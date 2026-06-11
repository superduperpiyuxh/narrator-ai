'use client';

import { useState, useEffect, useMemo } from 'react';
import { Grid3x3, ChevronDown, ChevronUp, ExternalLink } from 'lucide-react';
import { API_BASE, isAuthenticated } from '@/lib/api';
import type { Technique, Incident } from '@/lib/types';
import { cn } from '@/lib/utils';

const TACTIC_ORDER = [
  'initial-access',
  'execution',
  'persistence',
  'privilege-escalation',
  'defense-evasion',
  'credential-access',
  'discovery',
  'lateral-movement',
  'collection',
  'command-and-control',
  'exfiltration',
  'impact',
  'reconnaissance',
  'resource-development',
];

function tacticLabel(tactic: string): string {
  return tactic
    .split('-')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

function getColorClass(count: number): string {
  if (count === 0) return 'bg-zinc-900';
  if (count <= 2) return 'bg-green-900';
  if (count <= 5) return 'bg-yellow-900';
  if (count <= 10) return 'bg-orange-900';
  return 'bg-red-900';
}

export function TechniqueHeatmap() {
  const [techniques, setTechniques] = useState<Technique[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expanded, setExpanded] = useState(false);
  const [hoveredId, setHoveredId] = useState<string | null>(null);
  const [counts, setCounts] = useState<Map<string, number>>(new Map());

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        if (!isAuthenticated()) {
          if (!cancelled) {
            setTechniques([]);
            setLoading(false);
          }
          return;
        }

        const token = localStorage.getItem('nexus_token');
        const headers: Record<string, string> = { 'Content-Type': 'application/json' };
        if (token) headers['Authorization'] = `Bearer ${token}`;

        const [techRes, incRes] = await Promise.all([
          fetch(`${API_BASE}/api/techniques`, { headers }),
          fetch(`${API_BASE}/api/incidents?limit=1000`, { headers }),
        ]);

        if (techRes.status === 401 || incRes.status === 401) {
          localStorage.removeItem('nexus_token');
          window.location.href = '/login';
          return;
        }

        if (!techRes.ok) {
          if (!cancelled) {
            setError('Failed to load techniques');
            setLoading(false);
          }
          return;
        }

        const techData = await techRes.json();
        const allTechs: Technique[] = techData.techniques || [];

        let techniqueCounts = new Map<string, number>();
        if (incRes.ok) {
          const incData = await incRes.json();
          const incidents: Incident[] = incData.incidents || [];
          for (const inc of incidents) {
            for (const t of inc.techniques || []) {
              techniqueCounts.set(
                t.technique_id,
                (techniqueCounts.get(t.technique_id) || 0) + 1
              );
            }
          }
        }

        if (!cancelled) {
          setTechniques(allTechs);
          setCounts(techniqueCounts);
          setLoading(false);
        }
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : 'Failed to load heatmap data');
          setLoading(false);
        }
      }
    };
    load();
    return () => { cancelled = true; };
  }, []);

  // Group techniques by tactic
  const grouped = useMemo(() => {
    const map = new Map<string, Technique[]>();
    for (const t of techniques) {
      const key = t.tactic || 'unknown';
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(t);
    }
    // Sort within each tactic by ID
    for (const list of map.values()) {
      list.sort((a, b) => a.id.localeCompare(b.id));
    }
    return map;
  }, [techniques]);

  // Sorted tactic keys
  const sortedTactics = useMemo(() => {
    const keys = [...grouped.keys()];
    keys.sort((a, b) => {
      const ai = TACTIC_ORDER.indexOf(a);
      const bi = TACTIC_ORDER.indexOf(b);
      if (ai === -1 && bi === -1) return a.localeCompare(b);
      if (ai === -1) return 1;
      if (bi === -1) return 1;
      return ai - bi;
    });
    return keys;
  }, [grouped]);

  if (loading) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-8">
        <div className="flex items-center gap-3 mb-4">
          <Grid3x3 className="w-5 h-5 text-zinc-500 animate-pulse" />
          <span className="text-sm text-zinc-500">Loading MITRE ATT&CK heatmap...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4" role="alert">
        <p className="text-red-400 text-sm">Failed to load heatmap: {error}</p>
      </div>
    );
  }

  const totalTechniques = techniques.length;

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      {/* Collapsible header */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-5 py-4 hover:bg-zinc-800/50 transition-colors"
        aria-expanded={expanded}
      >
        <div className="flex items-center gap-3">
          <Grid3x3 className="w-5 h-5 text-blue-500" aria-hidden="true" />
          <div className="text-left">
            <h2 className="text-lg font-semibold text-zinc-100">MITRE ATT&CK Heatmap</h2>
            <p className="text-xs text-zinc-500">
              {totalTechniques} techniques across {sortedTactics.length} tactics
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          {/* Legend */}
          <div className="hidden md:flex items-center gap-2 text-[10px] text-zinc-500">
            <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 rounded bg-zinc-900 border border-zinc-700" /> 0</span>
            <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 rounded bg-green-900 border border-zinc-700" /> 1-2</span>
            <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 rounded bg-yellow-900 border border-zinc-700" /> 3-5</span>
            <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 rounded bg-orange-900 border border-zinc-700" /> 6-10</span>
            <span className="flex items-center gap-1"><span className="inline-block w-3 h-3 rounded bg-red-900 border border-zinc-700" /> 10+</span>
          </div>
          {expanded ? (
            <ChevronUp className="w-5 h-5 text-zinc-500" aria-hidden="true" />
          ) : (
            <ChevronDown className="w-5 h-5 text-zinc-500" aria-hidden="true" />
          )}
        </div>
      </button>

      {/* Collapsible body */}
      {expanded && (
        <div className="px-5 pb-5 space-y-5" role="region" aria-label="MITRE ATT&CK technique heatmap grid">
          {sortedTactics.map((tactic) => {
            const techs = grouped.get(tactic) || [];
            return (
              <div key={tactic}>
                <h3 className="text-xs font-semibold text-zinc-400 uppercase tracking-wider mb-2">
                  {tacticLabel(tactic)}
                </h3>
                <div className="flex flex-wrap gap-1.5">
                  {techs.map((tech) => {
                    const count = counts.get(tech.id) || 0;
                    const colorClass = getColorClass(count);
                    const isHovered = hoveredId === tech.id;
                    return (
                      <div key={tech.id} className="relative">
                        <a
                          href={tech.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={cn(
                            'block w-16 h-10 rounded border border-zinc-700/50 transition-all duration-150',
                            colorClass,
                            isHovered && 'ring-2 ring-blue-500 scale-110 z-10'
                          )}
                          onMouseEnter={() => setHoveredId(tech.id)}
                          onMouseLeave={() => setHoveredId(null)}
                          aria-label={`${tech.id}: ${tech.name} — ${count} incident${count !== 1 ? 's' : ''}`}
                        >
                          <div className="flex items-center justify-center h-full">
                            <span className="text-[9px] font-mono text-zinc-400 truncate px-0.5">
                              {tech.id}
                            </span>
                          </div>
                        </a>
                        {/* Tooltip */}
                        {isHovered && (
                          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 z-20 pointer-events-none">
                            <div className="bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 shadow-xl whitespace-nowrap">
                              <div className="text-[10px] font-mono text-blue-400 mb-0.5">{tech.id}</div>
                              <div className="text-xs text-zinc-200 font-medium">{tech.name}</div>
                              <div className="text-[10px] text-zinc-500 mt-0.5">
                                {count} incident{count !== 1 ? 's' : ''}
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            );
          })}

          {/* Summary footer */}
          <div className="flex items-center justify-between pt-3 border-t border-zinc-800">
            <div className="text-xs text-zinc-500">
              {totalTechniques} techniques
            </div>
            <a
              href="https://attack.mitre.org/"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-xs text-zinc-500 hover:text-zinc-300 transition-colors"
            >
              MITRE ATT&CK <ExternalLink className="w-3 h-3" />
            </a>
          </div>
        </div>
      )}
    </div>
  );
}
