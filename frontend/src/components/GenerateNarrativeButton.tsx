'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Zap, Loader2 } from 'lucide-react';
import toast from 'react-hot-toast';

interface GenerateNarrativeButtonProps {
  incidentId: number;
}

export function GenerateNarrativeButton({ incidentId }: GenerateNarrativeButtonProps) {
  const [isGenerating, setIsGenerating] = useState(false);
  const router = useRouter();

  const handleGenerate = async () => {
    setIsGenerating(true);
    try {
      const res = await fetch(`http://localhost:8080/api/incidents/${incidentId}/narrative`, {
        method: 'POST',
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `HTTP ${res.status}`);
      }

      toast.success('Narrative generated!');
      router.refresh();
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to generate narrative';
      toast.error(msg);
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <button
      onClick={handleGenerate}
      disabled={isGenerating}
      className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-500 disabled:bg-zinc-700 disabled:text-zinc-400 text-white rounded-lg font-medium transition-colors"
      aria-label={`Generate AI narrative for incident ${incidentId}`}
      aria-busy={isGenerating}
    >
      {isGenerating ? (
        <>
          <Loader2 className="w-4 h-4 animate-spin" aria-hidden="true" />
          Generating...
        </>
      ) : (
        <>
          <Zap className="w-4 h-4" aria-hidden="true" />
          Generate Narrative
        </>
      )}
    </button>
  );
}
