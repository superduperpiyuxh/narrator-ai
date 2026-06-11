'use client';

import { useEffect } from 'react';
import { Shield } from 'lucide-react';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('Application error:', error);
  }, [error]);

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center p-6">
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-12 text-center max-w-md">
        <Shield className="w-12 h-12 text-red-500 mx-auto mb-4" aria-hidden="true" />
        <h2 className="text-xl font-medium text-zinc-100 mb-2">Something went wrong</h2>
        <p className="text-zinc-400 mb-6">
          {error.message || 'An unexpected error occurred.'}
        </p>
        <button
          onClick={reset}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors"
        >
          Try again
        </button>
      </div>
    </div>
  );
}
