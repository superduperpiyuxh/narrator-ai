export default function IncidentLoading() {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        <div className="h-4 w-32 bg-surface rounded mb-6 animate-pulse" />
        <div className="h-8 w-96 bg-surface rounded mb-4 animate-pulse" />
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-16 bg-surface rounded-lg animate-pulse" />
          ))}
        </div>
        <div className="h-10 w-64 bg-surface rounded mb-6 animate-pulse" />
        <div className="h-48 bg-surface rounded-xl animate-pulse" />
      </div>
    </div>
  );
}
