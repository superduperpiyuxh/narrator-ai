'use client';

import { useState, useCallback } from 'react';
import useSWR from 'swr';
import { Feedback } from '@/lib/types';

const fetcher = (url: string) => fetch(url).then((res) => res.json());

export function useFeedback(narrativeId: number) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { data, mutate } = useSWR<{ feedback: Feedback | null }>(
    `/api/feedback/${narrativeId}`,
    fetcher
  );

  const submitFeedback = useCallback(
    async (rating: 'up' | 'down', notes: string = '') => {
      setIsSubmitting(true);
      try {
        const res = await fetch('/api/feedback', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            narrative_id: narrativeId,
            incident_id: 0, // Will be set by parent component
            rating: rating === 'up' ? 1 : -1,
            notes,
          }),
        });

        if (!res.ok) {
          throw new Error('Failed to submit feedback');
        }

        const { feedback } = await res.json();
        mutate({ feedback });
        return feedback;
      } finally {
        setIsSubmitting(false);
      }
    },
    [narrativeId, mutate]
  );

  return {
    feedback: data?.feedback || null,
    isSubmitting,
    submitFeedback,
    hasSubmitted: !!data?.feedback,
  };
}
