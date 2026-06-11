import { cn } from '@/lib/utils';

interface TechniqueBadgeProps {
  techniqueId: string;
  name: string;
  className?: string;
}

export function TechniqueBadge({ techniqueId, name, className }: TechniqueBadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-mono border border-border bg-surface/50 text-foreground/80',
        className
      )}
    >
      <span className="text-muted-foreground">{techniqueId}</span>
      <span className="text-muted-foreground">{name}</span>
    </span>
  );
}
