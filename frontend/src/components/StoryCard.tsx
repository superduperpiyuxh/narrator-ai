'use client';

import { useState, useCallback } from 'react';
import { Narrative, Sentence, Feedback } from '@/lib/types';
import { NarrativeSentence } from './NarrativeSentence';
import { RawEventViewer } from './RawEventViewer';
import { ConfidenceBadge } from './ConfidenceBadge';
import { FeedbackButton } from './FeedbackButton';
import { Cpu } from 'lucide-react';

interface StoryCardProps {
  narrative: Narrative;
  incidentId: number;
  existingFeedback?: Feedback | null;
}

export function StoryCard({ narrative, incidentId, existingFeedback }: StoryCardProps) {
  const [hoveredSentenceIndex, setHoveredSentenceIndex] = useState<number | null>(null);
  const [selectedEventIds, setSelectedEventIds] = useState<number[]>([]);

  let sentences: Sentence[] = [];
  try {
    sentences = typeof narrative.sentences === 'string'
      ? JSON.parse(narrative.sentences)
      : narrative.sentences || [];
  } catch {
    sentences = [];
  }

  const handleSentenceHover = useCallback((eventIds: number[]) => {
    setHoveredSentenceIndex(null);
    setSelectedEventIds(eventIds);
  }, []);

  const handleSentenceLeave = useCallback(() => {
    setSelectedEventIds([]);
    setHoveredSentenceIndex(null);
  }, []);

  return (
    <div className="flex gap-6">
      {/* Main story card */}
      <div className="flex-1 bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-zinc-800">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-zinc-100">Attack Narrative</h3>
            <div className="flex items-center gap-3">
              <ConfidenceBadge confidence={narrative.confidence} />
              <span className="text-xs text-zinc-500 bg-zinc-800 px-2 py-1 rounded">
                {narrative.model_used}
              </span>
            </div>
          </div>
        </div>

        {/* Summary */}
        {narrative.summary && (
          <div className="px-6 py-4 border-b border-zinc-800">
            <p className="text-sm text-zinc-300 italic">{narrative.summary}</p>
          </div>
        )}

        {/* Sentences */}
        <div className="px-6 py-4 space-y-3">
          {sentences.length > 0 ? (
            sentences.map((sentence, index) => (
              <NarrativeSentence
                key={index}
                sentence={sentence}
                index={index}
                isHovered={hoveredSentenceIndex === index}
                onHover={handleSentenceHover}
                onLeave={handleSentenceLeave}
              />
            ))
          ) : (
            <p className="text-zinc-500 text-sm">No sentences available</p>
          )}
        </div>

        {/* Metadata footer */}
        <div className="px-6 py-3 border-t border-zinc-800 bg-zinc-950/50">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4 text-xs font-mono text-zinc-500">
              <span>{narrative.tokens_used.toLocaleString()} tokens</span>
              <span>{narrative.generation_time_ms}ms</span>
              <span>T={narrative.temperature}</span>
              <Cpu className="w-3 h-3" />
            </div>
            <FeedbackButton
              narrativeId={narrative.id}
              incidentId={incidentId}
              existingFeedback={existingFeedback || null}
            />
          </div>
        </div>
      </div>

      {/* Source events panel */}
      <div className="w-80 flex-shrink-0">
        <div className="sticky top-4">
          <h4 className="text-sm font-medium text-zinc-400 mb-3">Source Events</h4>
          <RawEventViewer
            eventIds={selectedEventIds}
            narrativeId={narrative.id}
          />
        </div>
      </div>
    </div>
  );
}
