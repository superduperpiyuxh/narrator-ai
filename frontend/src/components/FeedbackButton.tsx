'use client';

import { useState } from 'react';
import { MessageSquare, CheckCircle } from 'lucide-react';
import { FeedbackForm } from './FeedbackForm';
import { Feedback } from '@/lib/types';
import { cn } from '@/lib/utils';

interface FeedbackButtonProps {
  narrativeId: number;
  incidentId: number;
  existingFeedback: Feedback | null;
}

export function FeedbackButton({ narrativeId, incidentId, existingFeedback }: FeedbackButtonProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  if (existingFeedback) {
    return (
      <div className="flex items-center gap-2 text-green-400 text-sm" role="status">
        <CheckCircle className="w-4 h-4" aria-hidden="true" />
        Feedback submitted
      </div>
    );
  }

  return (
    <div>
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className={cn(
          'flex items-center gap-2 px-4 py-2 rounded-lg text-sm transition-colors',
          isExpanded
            ? 'bg-zinc-700 text-zinc-200'
            : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
        )}
        aria-expanded={isExpanded}
        aria-label="Rate this narrative"
      >
        <MessageSquare className="w-4 h-4" aria-hidden="true" />
        Rate this narrative
      </button>
      {isExpanded && (
        <FeedbackForm
          narrativeId={narrativeId}
          incidentId={incidentId}
          existingFeedback={existingFeedback}
        />
      )}
    </div>
  );
}
