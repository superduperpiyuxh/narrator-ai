'use client';

import Link from 'next/link';
import { Incident } from '@/lib/types';
import { SeverityBadge } from './SeverityBadge';
import { ConfidenceBadge } from './ConfidenceBadge';
import { TechniqueBadge } from './TechniqueBadge';
import { formatTimestamp, truncate } from '@/lib/utils';

interface IncidentCardProps {
  incident: Incident;
}

export function IncidentCard({ incident }: IncidentCardProps) {
  const displayTechniques = incident.techniques?.slice(0, 3) || [];
  const remainingCount = (incident.techniques?.length || 0) - 3;

  return (
    <Link href={`/incidents/${incident.id}`}>
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 hover:border-zinc-600 transition-colors cursor-pointer h-full">
        {/* Top row */}
        <div className="flex items-center justify-between mb-3">
          <SeverityBadge severity={incident.severity} />
          <span className="text-xs font-mono text-zinc-500">#{incident.id}</span>
        </div>

        {/* Title */}
        <h3 className="text-lg font-semibold text-zinc-100 mb-2 line-clamp-1">
          {incident.title}
        </h3>

        {/* Description */}
        <p className="text-sm text-zinc-400 line-clamp-2 mb-4">
          {truncate(incident.description, 150)}
        </p>

        {/* Techniques */}
        <div className="flex flex-wrap gap-2 mb-4">
          {displayTechniques.map((tech) => (
            <TechniqueBadge
              key={tech.technique_id}
              techniqueId={tech.technique_id}
              name={tech.name}
            />
          ))}
          {remainingCount > 0 && (
            <span className="text-xs text-zinc-500 self-center">
              +{remainingCount} more
            </span>
          )}
        </div>

        {/* Bottom row */}
        <div className="flex items-center justify-between text-xs text-zinc-500">
          <span>{formatTimestamp(incident.start_time)}</span>
          <div className="flex items-center gap-3">
            <span className="bg-zinc-800 px-2 py-1 rounded">
              {incident.event_count} events
            </span>
            <ConfidenceBadge confidence={incident.confidence} />
          </div>
        </div>
      </div>
    </Link>
  );
}
