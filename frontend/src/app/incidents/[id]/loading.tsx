export default function IncidentLoading() {
  return (
    <div className="min-h-screen bg-zinc-950 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="h-4 w-32 bg-zinc-800 rounded mb-6 animate-pulse" />
        <div className="h-8 w-96 bg-zinc-800 rounded mb-4 animate-pulse" />
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-16 bg-zinc-800 rounded-lg animate-pulse" />
          ))}
        </div>
        <div className="h-10 w-64 bg-zinc-800 rounded mb-6 animate-pulse" />
        <div className="h-48 bg-zinc-800 rounded-xl animate-pulse" />
      </div>
    </div>
  );
}
