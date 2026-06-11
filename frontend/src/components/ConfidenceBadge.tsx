import { cn, formatConfidence, getConfidenceColor } from '@/lib/utils';

interface ConfidenceBadgeProps {
  confidence: number;
  className?: string;
}

export function ConfidenceBadge({ confidence, className }: ConfidenceBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border border-border',
        getConfidenceColor(confidence),
        className
      )}
      aria-label={`Confidence: ${formatConfidence(confidence)}`}
    >
      {formatConfidence(confidence)} confidence
    </span>
  );
}
