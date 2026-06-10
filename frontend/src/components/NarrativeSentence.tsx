'use client';

import { useState, useRef, useCallback } from 'react';
import { Sentence } from '@/lib/types';
import { TechniqueBadge } from './TechniqueBadge';
import { ConfidenceBadge } from './ConfidenceBadge';
import { formatTimestamp, cn } from '@/lib/utils';

interface NarrativeSentenceProps {
  sentence: Sentence;
  index: number;
  isHovered: boolean;
  onHover: (eventIds: number[]) => void;
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
    onHover(sentence.source_event_ids);
  }, [sentence.source_event_ids, onHover]);

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
          ? 'bg-zinc-800 border-l-2 border-blue-500 text-zinc-50'
          : 'text-zinc-200 hover:bg-zinc-800/50'
      )}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <p className="text-sm leading-relaxed">{sentence.text}</p>
      <div className="flex items-center gap-3 mt-2">
        <span className="text-xs text-zinc-500">
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
