export function IncidentCardSkeleton() {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 animate-pulse">
      <div className="flex items-center justify-between mb-3">
        <div className="h-5 w-16 bg-zinc-800 rounded-full" />
        <div className="h-4 w-8 bg-zinc-800 rounded" />
      </div>
      <div className="h-6 w-3/4 bg-zinc-800 rounded mb-2" />
      <div className="h-4 w-full bg-zinc-800 rounded mb-4" />
      <div className="flex gap-2 mb-4">
        <div className="h-5 w-20 bg-zinc-800 rounded-full" />
        <div className="h-5 w-20 bg-zinc-800 rounded-full" />
      </div>
      <div className="flex items-center justify-between">
        <div className="h-4 w-24 bg-zinc-800 rounded" />
        <div className="h-4 w-16 bg-zinc-800 rounded" />
      </div>
    </div>
  );
}

export function StoryCardSkeleton() {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden animate-pulse">
      <div className="px-6 py-4 border-b border-zinc-800">
        <div className="h-6 w-32 bg-zinc-800 rounded" />
      </div>
      <div className="px-6 py-4 border-b border-zinc-800">
        <div className="h-4 w-full bg-zinc-800 rounded mb-2" />
        <div className="h-4 w-2/3 bg-zinc-800 rounded" />
      </div>
      <div className="px-6 py-4 space-y-3">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="p-3 rounded-lg bg-zinc-800/50">
            <div className="h-4 w-full bg-zinc-800 rounded mb-2" />
            <div className="h-4 w-3/4 bg-zinc-800 rounded" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function RawEventViewerSkeleton() {
  return (
    <div className="space-y-2 animate-pulse">
      {[...Array(3)].map((_, i) => (
        <div key={i} className="h-24 bg-zinc-900 rounded-lg" />
      ))}
    </div>
  );
}

export function PageSkeleton() {
  return (
    <div className="min-h-screen bg-zinc-950 p-6">
      <div className="max-w-7xl mx-auto space-y-6 animate-pulse">
        <div className="h-16 bg-zinc-900 rounded-lg" />
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-24 bg-zinc-900 rounded-lg" />
          ))}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[...Array(6)].map((_, i) => (
            <IncidentCardSkeleton key={i} />
          ))}
        </div>
      </div>
    </div>
  );
}
