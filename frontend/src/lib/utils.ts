import { clsx, type ClassValue } from 'clsx';

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs);
}

export function formatTimestamp(ts: string): string {
  if (!ts) return 'Unknown';
  const date = new Date(ts);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffSec < 60) return 'Just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatConfidence(score: number): string {
  return `${Math.round(score * 100)}%`;
}

export function getSeverityColor(severity: string): string {
  switch (severity.toLowerCase()) {
    case 'critical':
      return 'bg-red-500/20 text-red-400 border-red-500/30';
    case 'high':
      return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
    case 'medium':
      return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30';
    case 'low':
      return 'bg-green-500/20 text-green-400 border-green-500/30';
    default:
      return 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30';
  }
}

export function getSeverityDot(severity: string): string {
  switch (severity.toLowerCase()) {
    case 'critical':
      return 'bg-red-500';
    case 'high':
      return 'bg-orange-500';
    case 'medium':
      return 'bg-yellow-500';
    case 'low':
      return 'bg-green-500';
    default:
      return 'bg-zinc-500';
  }
}

export function getConfidenceColor(score: number): string {
  if (score >= 0.8) return 'text-green-400';
  if (score >= 0.5) return 'text-yellow-400';
  return 'text-red-400';
}

export function truncate(str: string, length: number): string {
  if (str.length <= length) return str;
  return str.slice(0, length) + '...';
}
