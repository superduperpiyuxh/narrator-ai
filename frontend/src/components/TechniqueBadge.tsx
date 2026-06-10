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
        'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-mono border border-zinc-700 bg-zinc-800/50 text-zinc-300',
        className
      )}
    >
      <span className="text-zinc-500">{techniqueId}</span>
      <span className="text-zinc-400">{name}</span>
    </span>
  );
}
