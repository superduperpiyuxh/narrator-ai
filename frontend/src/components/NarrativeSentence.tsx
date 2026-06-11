'use client';

import { useRef, useCallback } from 'react';
import { Sentence } from '@/lib/types';
import { TechniqueBadge } from './TechniqueBadge';
import { ConfidenceBadge } from './ConfidenceBadge';
import { formatTimestamp, cn } from '@/lib/utils';

interface NarrativeSentenceProps {
  sentence: Sentence;
  index: number;
  isHovered: boolean;
  onHover: (eventIds: number[], index: number) => void;
  onLeave: () => void;
}

export function NarrativeSentence({
  sentence,
  index,
  isHovered,
  onHover,
  onLeave,
}: NarrativeSentenceProps) {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleMouseEnter = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    onHover(sentence.source_event_ids, index);
  }, [sentence.source_event_ids, onHover, index]);

  const handleMouseLeave = useCallback(() => {
    timeoutRef.current = setTimeout(() => {
      onLeave();
    }, 200);
  }, [onLeave]);

  return (
    <div
      className={cn(
        'p-3 rounded-lg cursor-pointer transition-all duration-150',
        isHovered
          ? 'bg-surface border-l-2 border-primary text-foreground'
          : 'text-foreground/80 hover:bg-surface/50'
      )}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onHover(sentence.source_event_ids, index);
        }
      }}
      aria-label={`Narrative sentence ${index + 1}: ${sentence.text}`}
    >
      <p className="text-sm leading-relaxed">{sentence.text}</p>
      <div className="flex items-center gap-3 mt-2">
        <span className="text-xs text-muted-foreground">
          {formatTimestamp(sentence.timestamp)}
        </span>
        {sentence.technique && (
          <TechniqueBadge techniqueId={sentence.technique} name={sentence.technique} />
        )}
        <ConfidenceBadge confidence={sentence.confidence} />
      </div>
    </div>
  );
}
