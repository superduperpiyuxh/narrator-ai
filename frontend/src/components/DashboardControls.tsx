'use client';

import { useState, useRef, useImperativeHandle, forwardRef } from 'react';
import { useRouter } from 'next/navigation';
import { Search, X } from 'lucide-react';

export interface DashboardControlsHandle {
  focusSearch: () => void;
  clearSearch: () => void;
}

interface DashboardControlsProps {
  currentPage: number;
  totalPages: number;
  total: number;
  currentSeverity: string;
  currentStatus: string;
  currentSourceIP: string;
}

export const DashboardControls = forwardRef<DashboardControlsHandle, DashboardControlsProps>(
  function DashboardControls(
    {
      currentPage,
      totalPages,
      total,
      currentSeverity,
      currentStatus,
      currentSourceIP,
    },
    ref
  ) {
    const router = useRouter();
    const [searchIP, setSearchIP] = useState(currentSourceIP);
    const searchInputRef = useRef<HTMLInputElement>(null);

    useImperativeHandle(ref, () => ({
      focusSearch: () => {
        searchInputRef.current?.focus();
      },
      clearSearch: () => {
        setSearchIP('');
        searchInputRef.current?.focus();
      },
    }));

  const buildURL = (overrides: Record<string, string>) => {
    const params = new URLSearchParams();
    const values = {
      severity: currentSeverity,
      status: currentStatus,
      source_ip: currentSourceIP,
      page: '1',
      ...overrides,
    };
    if (values.severity) params.set('severity', values.severity);
    if (values.status) params.set('status', values.status);
    if (values.source_ip) params.set('source_ip', values.source_ip);
    if (values.page !== '1') params.set('page', values.page);
    const qs = params.toString();
    return `/${qs ? '?' + qs : ''}`;
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    router.push(buildURL({ source_ip: searchIP, page: '1' }));
  };

  const hasFilters = currentSeverity || currentStatus || currentSourceIP;

  return (
    <div className="mb-6 space-y-4">
      {/* Search and filter row */}
      <div className="flex flex-col sm:flex-row gap-3">
        {/* Search by source IP */}
        <form onSubmit={handleSearch} className="flex-1 flex gap-2">
          <label htmlFor="source-ip-search" className="sr-only">
            Search by source IP
          </label>
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" aria-hidden="true" />
            <input
              id="source-ip-search"
              ref={searchInputRef}
              type="text"
              value={searchIP}
              onChange={(e) => setSearchIP(e.target.value)}
              placeholder="Search by source IP..."
              className="w-full bg-card border border-border rounded-lg pl-10 pr-4 py-2 text-sm text-foreground/80 placeholder-muted-foreground focus:outline-none focus:border-primary"
            />
          </div>
          <button
            type="submit"
            className="px-4 py-2 bg-surface text-foreground/80 rounded-lg hover:bg-surface-hover transition-colors text-sm"
          >
            Search
          </button>
        </form>

        {/* Severity filter */}
        <div className="flex gap-2">
          <label htmlFor="severity-filter" className="sr-only">
            Filter by severity
          </label>
          <select
            id="severity-filter"
            value={currentSeverity}
            onChange={(e) => router.push(buildURL({ severity: e.target.value, page: '1' }))}
            className="bg-card border border-border rounded-lg px-3 py-2 text-sm text-foreground/80 focus:outline-none focus:border-primary"
          >
            <option value="">All Severities</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
      </div>

      {/* Active filters */}
      {hasFilters && (
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-muted-foreground">Active filters:</span>
          {currentSeverity && (
            <span className="inline-flex items-center gap-1 px-2 py-1 bg-surface rounded text-xs text-foreground/80">
              Severity: {currentSeverity}
              <button
                onClick={() => router.push(buildURL({ severity: '', page: '1' }))}
                className="text-muted-foreground hover:text-foreground/80"
                aria-label={`Remove severity filter: ${currentSeverity}`}
              >
                <X className="w-3 h-3" />
              </button>
            </span>
          )}
          {currentSourceIP && (
            <span className="inline-flex items-center gap-1 px-2 py-1 bg-surface rounded text-xs text-foreground/80">
              IP: {currentSourceIP}
              <button
                onClick={() => router.push(buildURL({ source_ip: '', page: '1' }))}
                className="text-muted-foreground hover:text-foreground/80"
                aria-label={`Remove IP filter: ${currentSourceIP}`}
              >
                <X className="w-3 h-3" />
              </button>
            </span>
          )}
          <button
            onClick={() => router.push('/')}
            className="text-xs text-muted-foreground hover:text-foreground/80 underline"
          >
            Clear all
          </button>
        </div>
      )}
    </div>
  );
});
