export default function Loading() {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header skeleton */}
        <div className="h-16 bg-card rounded-lg animate-pulse" />
        
        {/* Stats skeleton */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-24 bg-card rounded-lg animate-pulse" />
          ))}
        </div>
        
        {/* Content skeleton */}
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-32 bg-card rounded-lg animate-pulse" />
          ))}
        </div>
      </div>
    </div>
  );
}
