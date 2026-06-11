'use client';

import { useState, useCallback } from 'react';
import useSWR from 'swr';
import { API_BASE } from '@/lib/api';
import { Feedback } from '@/lib/types';

function getAuthHeaders() {
  const token = localStorage.getItem('nexus_token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return headers;
}

const fetcher = (url: string) => fetch(url, { headers: getAuthHeaders() }).then((res) => res.json());

export function useFeedback(narrativeId: number, incidentId?: number) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { data, mutate } = useSWR<{ feedback: Feedback | null }>(
    `${API_BASE}/api/feedback/${narrativeId}`,
    fetcher
  );

  const submitFeedback = useCallback(
    async (rating: 'up' | 'down', notes: string = '') => {
      setIsSubmitting(true);
      try {
        const res = await fetch(`${API_BASE}/api/feedback`, {
          method: 'POST',
          headers: getAuthHeaders(),
          body: JSON.stringify({
            narrative_id: narrativeId,
            incident_id: incidentId || 0,
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
    [narrativeId, incidentId, mutate]
  );

  return {
    feedback: data?.feedback || null,
    isSubmitting,
    submitFeedback,
    hasSubmitted: !!data?.feedback,
  };
}
