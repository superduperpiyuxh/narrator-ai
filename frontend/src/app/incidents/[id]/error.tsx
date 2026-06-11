'use client';

import { useEffect } from 'react';
import { Shield } from 'lucide-react';
import Link from 'next/link';

export default function IncidentError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('Incident error:', error);
  }, [error]);

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-6">
      <div className="bg-card border border-border rounded-xl p-12 text-center max-w-md">
        <Shield className="w-12 h-12 text-red-500 mx-auto mb-4" aria-hidden="true" />
        <h2 className="text-xl font-medium text-foreground mb-2">Failed to load incident</h2>
        <p className="text-muted-foreground mb-6">
          {error.message || 'Could not load the incident details.'}
        </p>
        <div className="flex gap-3 justify-center">
          <button
            onClick={reset}
            className="px-4 py-2 bg-primary hover:bg-primary/90 text-white rounded-lg transition-colors"
          >
            Try again
          </button>
          <Link
            href="/"
            className="px-4 py-2 bg-surface hover:bg-surface-hover text-foreground/80 rounded-lg transition-colors"
          >
            Back to Dashboard
          </Link>
        </div>
      </div>
    </div>
  );
}
