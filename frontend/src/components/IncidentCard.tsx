'use client';

import { memo } from 'react';
import { Incident } from '@/lib/types';
import { SeverityBadge } from './SeverityBadge';
import { ConfidenceBadge } from './ConfidenceBadge';
import { TechniqueBadge } from './TechniqueBadge';
import { cn, formatTimestamp, getSeverityDot } from '@/lib/utils';

interface IncidentCardProps {
  incident: Incident;
  selected?: boolean;
}

export const IncidentCard = memo(function IncidentCard({ incident, selected = false }: IncidentCardProps) {
  const displayTechniques = incident.techniques?.slice(0, 3) || [];
  const remainingCount = (incident.techniques?.length || 0) - 3;

  const severityBorder: Record<string, string> = {
    critical: 'border-l-severity-critical',
    high: 'border-l-severity-high',
    medium: 'border-l-severity-medium',
    low: 'border-l-severity-low',
  };

  return (
    <a
      href={`/incidents/${incident.id}`}
      className={cn(
        'block bg-card border rounded-xl p-6 transition-colors cursor-pointer h-full border-l-4 hover:bg-surface-hover focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2',
        selected
          ? 'border-primary ring-1 ring-primary/30'
          : 'border-border',
        severityBorder[incident.severity] || 'border-l-muted-foreground'
      )}
      aria-label={`Incident ${incident.id}: ${incident.title}, severity ${incident.severity}, ${incident.event_count} events`}
    >
      {/* Top row */}
      <div className="flex items-center justify-between mb-3">
        <SeverityBadge severity={incident.severity} />
        <span className="text-xs font-mono text-muted-foreground">#{incident.id}</span>
      </div>

      {/* Title */}
      <h3 className="text-lg font-semibold text-foreground mb-2 line-clamp-1">
        {incident.title}
      </h3>

      {/* Description */}
      <p className="text-sm text-muted-foreground line-clamp-2 mb-4">
        {incident.description || incident.title}
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
          <span className="text-xs text-muted-foreground self-center">
            +{remainingCount} more
          </span>
        )}
      </div>

      {/* Bottom row */}
      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <time dateTime={incident.start_time}>
          {formatTimestamp(incident.start_time)}
        </time>
        <div className="flex items-center gap-3">
          <span className="bg-muted px-2 py-1 rounded">
            {incident.event_count} events
          </span>
          <ConfidenceBadge confidence={incident.confidence} />
        </div>
      </div>
    </a>
  );
});
