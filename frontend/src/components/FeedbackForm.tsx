'use client';

import { useState } from 'react';
import { ThumbsUp, ThumbsDown } from 'lucide-react';
import { useFeedback } from '@/hooks/useFeedback';
import { Feedback } from '@/lib/types';
import toast from 'react-hot-toast';
import { cn } from '@/lib/utils';

interface FeedbackFormProps {
  narrativeId: number;
  incidentId: number;
  existingFeedback: Feedback | null;
}

export function FeedbackForm({ narrativeId, incidentId, existingFeedback }: FeedbackFormProps) {
  const [rating, setRating] = useState<'up' | 'down' | null>(
    existingFeedback ? (existingFeedback.rating === 1 ? 'up' : 'down') : null
  );
  const [notes, setNotes] = useState(existingFeedback?.notes || '');
  const { submitFeedback, isSubmitting } = useFeedback(narrativeId, incidentId);

  const handleSubmit = async () => {
    if (!rating) return;

    try {
      await submitFeedback(rating, notes);
      toast.success('Feedback submitted');
    } catch {
      toast.error('Failed to submit feedback');
    }
  };

  if (existingFeedback) {
    return (
      <div className="bg-card border border-border rounded-xl p-6 mt-6">
        <div className="flex items-center gap-3 mb-3">
          <span className="text-sm text-muted-foreground">Your feedback:</span>
          <span
            className={cn(
              'flex items-center gap-1 px-2 py-1 rounded text-sm',
              existingFeedback.rating === 1
                ? 'bg-green-900/30 text-green-400'
                : 'bg-red-900/30 text-red-400'
            )}
          >
            {existingFeedback.rating === 1 ? (
              <ThumbsUp className="w-4 h-4" />
            ) : (
              <ThumbsDown className="w-4 h-4" />
            )}
            {existingFeedback.rating === 1 ? 'Helpful' : 'Not helpful'}
          </span>
        </div>
        {existingFeedback.notes && (
          <p className="text-sm text-foreground/80 bg-surface rounded p-3">
            {existingFeedback.notes}
          </p>
        )}
      </div>
    );
  }

  return (
    <div className="bg-card border border-border rounded-xl p-6 mt-6">
      <p className="text-sm text-muted-foreground mb-4">Was this narrative helpful?</p>

      {/* Rating buttons */}
      <div className="flex items-center gap-3 mb-4" role="radiogroup" aria-label="Narrative rating">
        <button
          onClick={() => setRating('up')}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg transition-colors',
            rating === 'up'
              ? 'bg-green-900/30 text-green-400 border border-green-700'
              : 'bg-surface text-muted-foreground hover:bg-surface-hover border border-transparent'
          )}
          role="radio"
          aria-checked={rating === 'up'}
          aria-label="Yes, this narrative was helpful"
        >
          <ThumbsUp className="w-4 h-4" aria-hidden="true" />
          Yes
        </button>
        <button
          onClick={() => setRating('down')}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg transition-colors',
            rating === 'down'
              ? 'bg-red-900/30 text-red-400 border border-red-700'
              : 'bg-surface text-muted-foreground hover:bg-surface-hover border border-transparent'
          )}
          role="radio"
          aria-checked={rating === 'down'}
          aria-label="No, this narrative was not helpful"
        >
          <ThumbsDown className="w-4 h-4" aria-hidden="true" />
          No
        </button>
      </div>

      {/* Notes textarea */}
      <label htmlFor={`feedback-notes-${narrativeId}`} className="sr-only">
        Optional feedback notes
      </label>
      <textarea
        id={`feedback-notes-${narrativeId}`}
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
        placeholder="Optional notes about this narrative..."
        className="w-full bg-surface border border-border rounded-lg p-3 text-sm text-foreground/80 placeholder-muted-foreground resize-none focus:outline-none focus:border-primary mb-4"
        rows={3}
        maxLength={1000}
      />

      {/* Submit button */}
      <button
        onClick={handleSubmit}
        disabled={!rating || isSubmitting}
        className={cn(
          'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
          rating && !isSubmitting
            ? 'bg-primary hover:bg-primary/90 text-white'
            : 'bg-surface text-muted-foreground cursor-not-allowed'
        )}
        aria-busy={isSubmitting}
      >
        {isSubmitting ? 'Submitting...' : 'Submit Feedback'}
      </button>
    </div>
  );
}
