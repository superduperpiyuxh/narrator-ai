# Phase 4: Story Dashboard - Research

**Researched:** 2026-06-10
**Domain:** Frontend dashboard, UI/UX for SOC analysts, source tracing visualization
**Confidence:** HIGH

## Summary

Phase 4 builds the analyst-facing dashboard that displays incidents as browsable story cards with source tracing and feedback capabilities. The frontend uses Next.js 16 with App Router, React 19 Server Components for data fetching, and Client Components for interactive elements. The dark theme follows SOC analyst UX conventions (dark backgrounds, high-contrast severity indicators, monospace for technical data). Source tracing uses a hover-to-reveal pattern where hovering a narrative sentence displays the corresponding raw event in a side panel. The feedback system requires a new backend endpoint and database table for thumbs up/down plus notes.

**Primary recommendation:** Use Server Components for the incident list page (data fetching from Go backend), Client Components for interactive story cards and hover-to-reveal, and a simple React Context for state management. Add a `/api/feedback` endpoint to the Go backend with a new `feedback` table.

<user_constraints>
## User Constraints (from CONTEXT.md)

No CONTEXT.md found — using ROADMAP.md and REQUIREMENTS.md constraints.

### Locked Decisions (from ROADMAP.md)
- Next.js 16 with App Router for frontend
- React 19 for UI components
- Tailwind CSS 4 for styling
- Dark theme for SOC analysts
- Hover-to-reveal source tracing pattern
- Thumbs up/down + notes feedback on narrative quality
- Every narrative sentence links to source event IDs (from Phase 3)

### Agent's Discretion
- Component architecture (pages, layouts, components)
- State management approach (Context vs SWR vs React Query)
- Story card visual design
- Feedback form UX
- Responsive design considerations
- Loading states and error handling patterns

### Deferred Ideas (OUT OF SCOPE)
- Real-time streaming updates
- User authentication (single analyst view for demo)
- Mobile responsiveness (web-only for v1)
- Export functionality
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DASH-01 | System displays incident list view with severity indicators | Server Component fetching from `/api/incidents` with Tailwind severity badges |
| DASH-02 | System shows story cards with narrative sentences and confidence scores | Client Component with hover state management and confidence visualization |
| DASH-03 | System provides raw event viewer with source tracing (hover sentence → see raw event) | Hover event handler + side panel component + `/api/narratives/:id` endpoint |
| DASH-04 | System captures analyst feedback (thumbs up/down + notes) | New feedback endpoint + database table + form component |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Incident list rendering | Browser / Client | Frontend Server (SSR) | Server Component fetches data, renders list |
| Story card interactivity | Browser / Client | — | Client Component handles hover/click states |
| Source tracing hover-to-reveal | Browser / Client | API / Backend | Client handles UI, backend provides event data |
| Feedback submission | Browser / Client | API / Backend | Client form, backend stores in database |
| Data fetching from backend | Frontend Server (SSR) | API / Backend | Server Components fetch directly from Go API |
| State management | Browser / Client | — | React Context or SWR for client-side state |
| Dark theme styling | Browser / Client | — | Tailwind CSS classes for visual presentation |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| next | 16.2.7 | React framework with App Router | Already installed, best DX for React 19 |
| react | 19.2.4 | UI library | Already installed, Server Components support |
| tailwindcss | ^4 | Utility-first CSS | Already installed, rapid dark theme development |
| swr | ^2 | Client-side data fetching | Recommended by Next.js docs, lightweight, great DX |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| react-hot-toast | ^2 | Toast notifications | For feedback submission success/error |
| lucide-react | ^0.400 | Icon library | For thumbs up/down, severity icons, navigation |
| clsx | ^2 | Conditional classnames | For dynamic Tailwind classes based on state |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| SWR | React Query (TanStack Query) | SWR is lighter, React Query has more features; SWR recommended by Next.js docs |
| Tailwind CSS | CSS Modules | Tailwind faster for rapid prototyping, dark mode built-in |
| Lucide icons | Heroicons | Lucide has more security-related icons, smaller bundle |
| React Context | Zustand | Context simpler for this scope, Zustand for larger apps |

**Installation:**
```bash
cd frontend
npm install swr lucide-react clsx react-hot-toast
```

**Version verification:**
```bash
npm view swr version
npm view lucide-react version
npm view clsx version
npm view react-hot-toast version
```

## Package Legitimacy Audit

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| swr | npm | 4+ years | 5M+/week | github.com/vercel/swr | OK | Approved |
| lucide-react | npm | 3+ years | 2M+/week | github.com/lucide-icons/lucide | OK | Approved |
| clsx | npm | 7+ years | 10M+/week | github.com/lukeed/clsx | OK | Approved |
| react-hot-toast | npm | 3+ years | 500K+/week | github.com/timolins/react-hot-toast | OK | Approved |

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Browser / Client                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ IncidentList │  │ StoryCard   │  │ RawEventViewer     │ │
│  │ (Server)     │  │ (Client)    │  │ (Client)           │ │
│  └──────┬───────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         │                 │                    │             │
│         │    ┌────────────┴────────────┐       │             │
│         │    │   Hover State Manager   │       │             │
│         │    │   (React Context/SWR)   │       │             │
│         │    └────────────┬────────────┘       │             │
│         │                 │                    │             │
└─────────┼─────────────────┼────────────────────┼─────────────┘
          │                 │                    │
          ▼                 ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│              Frontend Server (SSR)                           │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Server Components fetch data from Go backend        │   │
│  │  - Incident list page                                │   │
│  │  - Incident detail page                              │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
          │                 │                    │
          ▼                 ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│                    API / Backend                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ /api/incidents│ │ /api/narratives│ │ /api/feedback      │ │
│  │ (GET list)   │ │ (GET source) │  │ (POST new)         │ │
│  └──────┬───────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         │                 │                    │             │
└─────────┼─────────────────┼────────────────────┼─────────────┘
          │                 │                    │
          ▼                 ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│               Database / Storage                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ incidents   │  │ narratives  │  │ feedback            │ │
│  │ events      │  │ sentences   │  │ (NEW)               │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure
```
frontend/src/
├── app/
│   ├── layout.tsx              # Root layout with dark theme, fonts
│   ├── page.tsx                # Dashboard home (incident list)
│   ├── globals.css             # Tailwind imports + custom theme
│   ├── incidents/
│   │   └── [id]/
│   │       └── page.tsx        # Incident detail with story card
│   └── loading.tsx             # Global loading state
├── components/
│   ├── IncidentCard.tsx        # Incident list item card
│   ├── StoryCard.tsx           # Narrative sentences with hover
│   ├── NarrativeSentence.tsx   # Individual sentence with source tracing
│   ├── RawEventViewer.tsx      # Side panel for raw events
│   ├── ConfidenceBadge.tsx     # Confidence score visualization
│   ├── SeverityBadge.tsx       # Severity indicator (low/med/high/critical)
│   ├── TechniqueBadge.tsx      # MITRE ATT&CK technique label
│   ├── FeedbackForm.tsx        # Thumbs up/down + notes form
│   ├── FeedbackButton.tsx      # Toggle feedback form visibility
│   └── LoadingSkeleton.tsx     # Skeleton loading states
├── lib/
│   ├── api.ts                  # API client functions
│   ├── types.ts                # TypeScript interfaces
│   └── utils.ts                # Utility functions (formatDate, etc.)
└── hooks/
    └── useFeedback.ts          # Feedback submission hook
```

### Pattern 1: Server Component Data Fetching
**What:** Fetch data in Server Components, pass to Client Components via props
**When to use:** Initial page load data that doesn't need client-side re-fetching
**Example:**
```tsx
// Source: Next.js 16 docs - Server and Client Components
// app/page.tsx (Server Component)
import IncidentCard from '@/components/IncidentCard'

async function getIncidents() {
  const res = await fetch('http://localhost:8080/api/incidents?limit=50')
  if (!res.ok) throw new Error('Failed to fetch incidents')
  return res.json()
}

export default async function DashboardPage() {
  const { incidents, total } = await getIncidents()

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100">
      <h1 className="text-2xl font-bold mb-6">Security Incidents</h1>
      <div className="grid gap-4">
        {incidents.map((incident) => (
          <IncidentCard key={incident.id} incident={incident} />
        ))}
      </div>
    </div>
  )
}
```

### Pattern 2: Client Component with SWR for Revalidation
**What:** Use SWR for client-side data fetching with caching and revalidation
**When to use:** Data that needs to be refreshed without page reload (feedback, hover events)
**Example:**
```tsx
// Source: Next.js 16 docs - Fetching Data (Community libraries)
// components/RawEventViewer.tsx
'use client'

import useSWR from 'swr'

const fetcher = (url: string) => fetch(url).then((r) => r.json())

export default function RawEventViewer({ narrativeId }: { narrativeId: number }) {
  const { data, error, isLoading } = useSWR(
    narrativeId ? `/api/narratives/${narrativeId}` : null,
    fetcher
  )

  if (isLoading) return <div className="animate-pulse bg-zinc-800 h-32 rounded" />
  if (error) return <div className="text-red-400">Failed to load events</div>

  return (
    <div className="bg-zinc-900 border border-zinc-700 rounded-lg p-4">
      <h3 className="text-sm font-mono text-zinc-400 mb-2">Source Events</h3>
      {data.events.map((event: any) => (
        <div key={event.id} className="font-mono text-xs text-zinc-300 mb-1">
          [{event.timestamp}] {event.event_type}: {event.source_ip}
        </div>
      ))}
    </div>
  )
}
```

### Pattern 3: Hover-to-Reveal Source Tracing
**What:** Hover over a narrative sentence to highlight and reveal source events
**When to use:** For DASH-03 requirement - source tracing visualization
**Example:**
```tsx
// components/NarrativeSentence.tsx
'use client'

import { useState } from 'react'
import { Sentence } from '@/lib/types'

interface NarrativeSentenceProps {
  sentence: Sentence
  onHover: (sourceEventIds: number[]) => void
  onLeave: () => void
}

export default function NarrativeSentence({
  sentence,
  onHover,
  onLeave
}: NarrativeSentenceProps) {
  const [isHovered, setIsHovered] = useState(false)

  return (
    <div
      className={`p-3 rounded-lg cursor-pointer transition-colors ${
        isHovered ? 'bg-zinc-800 border-l-2 border-blue-500' : 'hover:bg-zinc-800/50'
      }`}
      onMouseEnter={() => {
        setIsHovered(true)
        onHover(sentence.source_event_ids)
      }}
      onMouseLeave={() => {
        setIsHovered(false)
        onLeave()
      }}
    >
      <p className="text-zinc-200">{sentence.text}</p>
      <div className="flex items-center gap-2 mt-2 text-xs">
        <span className="text-zinc-500">{sentence.timestamp}</span>
        {sentence.technique && (
          <span className="bg-red-900/30 text-red-300 px-2 py-0.5 rounded font-mono">
            {sentence.technique}
          </span>
        )}
        <span className="text-zinc-600">
          Confidence: {(sentence.confidence * 100).toFixed(0)}%
        </span>
      </div>
    </div>
  )
}
```

### Pattern 4: Feedback Form with Optimistic Updates
**What:** Submit feedback with optimistic UI update, then confirm with server
**When to use:** For DASH-04 requirement - analyst feedback system
**Example:**
```tsx
// components/FeedbackForm.tsx
'use client'

import { useState } from 'react'
import { ThumbsUp, ThumbsDown } from 'lucide-react'

interface FeedbackFormProps {
  narrativeId: number
  incidentId: number
  onSubmit: (feedback: { rating: 'up' | 'down'; notes: string }) => void
}

export default function FeedbackForm({ narrativeId, incidentId, onSubmit }: FeedbackFormProps) {
  const [rating, setRating] = useState<'up' | 'down' | null>(null)
  const [notes, setNotes] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!rating) return

    setIsSubmitting(true)
    try {
      const res = await fetch('/api/feedback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          narrative_id: narrativeId,
          incident_id: incidentId,
          rating,
          notes
        })
      })

      if (res.ok) {
        onSubmit({ rating, notes })
        setRating(null)
        setNotes('')
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="bg-zinc-900 border border-zinc-700 rounded-lg p-4 mt-4">
      <p className="text-sm text-zinc-400 mb-3">Was this narrative helpful?</p>
      <div className="flex gap-2 mb-3">
        <button
          onClick={() => setRating('up')}
          className={`p-2 rounded ${
            rating === 'up'
              ? 'bg-green-900/30 text-green-400 border border-green-700'
              : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
          }`}
        >
          <ThumbsUp size={16} />
        </button>
        <button
          onClick={() => setRating('down')}
          className={`p-2 rounded ${
            rating === 'down'
              ? 'bg-red-900/30 text-red-400 border border-red-700'
              : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
          }`}
        >
          <ThumbsDown size={16} />
        </button>
      </div>
      <textarea
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
        placeholder="Optional notes about this narrative..."
        className="w-full bg-zinc-800 border border-zinc-600 rounded p-2 text-sm text-zinc-200 placeholder-zinc-500 resize-none"
        rows={3}
      />
      <button
        onClick={handleSubmit}
        disabled={!rating || isSubmitting}
        className="mt-2 px-4 py-2 bg-blue-600 text-white rounded text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-blue-500"
      >
        {isSubmitting ? 'Submitting...' : 'Submit Feedback'}
      </button>
    </div>
  )
}
```

### Anti-Patterns to Avoid
- **Don't fetch data in Client Components when Server Components can do it:** Server Components reduce client-side JavaScript and improve initial load
- **Don't use useState for data that comes from the server:** Use Server Components or SWR for server data
- **Don't forget the `'use client'` directive:** Interactive components need it, but keep it minimal
- **Don't hardcode API URLs:** Use environment variables or a central API client
- **Don't skip loading states:** Always show skeleton/spinner during data fetches

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Client-side data fetching | Custom fetch hook with useState/useEffect | SWR | Handles caching, revalidation, error states, loading states |
| Icon components | Custom SVG icons | Lucide React | Well-maintained, consistent, tree-shakeable |
| Conditional classNames | String concatenation | clsx | Handles edge cases, cleaner syntax |
| Toast notifications | Custom alert/toast system | react-hot-toast | Handles positioning, animations, stacking |
| Dark theme colors | Custom CSS variables | Tailwind dark: prefix | Built-in, consistent, optimized |

**Key insight:** The existing frontend already has Tailwind CSS 4 configured. Use the built-in `dark:` prefix for dark mode instead of creating custom theme systems. The `@theme inline` block in globals.css allows custom design tokens.

## Common Pitfalls

### Pitfall 1: Next.js 16 Params are Promises
**What goes wrong:** Trying to access `params.id` directly instead of awaiting it
**Why it happens:** Next.js 16 changed params to be Promise-based for consistency
**How to avoid:** Always destructure with `await`: `const { id } = await params`
**Warning signs:** TypeScript errors about params being Promise type

### Pitfall 2: Forgetting Server vs Client Component Boundaries
**What goes wrong:** Using useState/useEffect in Server Components, or trying to fetch data directly in Client Components
**Why it happens:** Confusion about which components run where
**How to avoid:** Mark interactive components with `'use client'`, keep data fetching in Server Components
**Warning signs:** "useState is not a function" errors, or client-side data fetching issues

### Pitfall 3: CORS Issues Between Frontend and Backend
**What goes wrong:** API requests fail due to CORS policy
**Why it happens:** Frontend on port 3000, backend on port 8080
**How to avoid:** Backend already has CORS configured for localhost:3000. For Server Components, use absolute URLs or configure Next.js rewrites
**Warning signs:** CORS errors in browser console

### Pitfall 4: Large Bundle Sizes from Client Components
**What goes wrong:** Sending too much JavaScript to the client
**Why it happens:** Marking entire pages as Client Components instead of specific interactive parts
**How to avoid:** Keep `'use client'` at the component level, not page level. Use Server Components for static content
**Warning signs:** Slow initial page load, large bundle size warnings

### Pitfall 5: Missing Loading States
**What goes wrong:** Users see blank screen or broken UI while data loads
**Why it happens:** Forgetting to add loading.tsx or Suspense boundaries
**How to avoid:** Add loading.tsx for route-level loading, Suspense for component-level
**Warning signs:** White flash on navigation, unresponsive UI during data fetches

## Code Examples

### API Client Configuration
```typescript
// lib/api.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export async function fetchIncidents(limit = 50, offset = 0) {
  const res = await fetch(`${API_BASE}/api/incidents?limit=${limit}&offset=${offset}`)
  if (!res.ok) throw new Error('Failed to fetch incidents')
  return res.json()
}

export async function fetchIncident(id: number) {
  const res = await fetch(`${API_BASE}/api/incidents/${id}`)
  if (!res.ok) throw new Error('Failed to fetch incident')
  return res.json()
}

export async function fetchNarrative(incidentId: number) {
  const res = await fetch(`${API_BASE}/api/incidents/${incidentId}/narrative`)
  if (!res.ok) throw new Error('Failed to fetch narrative')
  return res.json()
}

export async function fetchNarrativeSourceEvents(narrativeId: number) {
  const res = await fetch(`${API_BASE}/api/narratives/${narrativeId}`)
  if (!res.ok) throw new Error('Failed to fetch source events')
  return res.json()
}

export async function submitFeedback(feedback: {
  narrative_id: number
  incident_id: number
  rating: 'up' | 'down'
  notes: string
}) {
  const res = await fetch(`${API_BASE}/api/feedback`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(feedback)
  })
  if (!res.ok) throw new Error('Failed to submit feedback')
  return res.json()
}
```

### TypeScript Types
```typescript
// lib/types.ts
export interface Incident {
  id: number
  title: string
  description: string
  source_ip: string
  start_time: string
  end_time: string
  event_count: number
  unique_users: string[]
  unique_ips: string[]
  unique_hostnames: string[]
  severity: 'low' | 'medium' | 'high' | 'critical'
  status: string
  techniques: TechniqueRef[]
  tactics: string[]
  mitre_attack_ids: string[]
  confidence: number
  raw_summary: string
  created_at: string
  updated_at: string
}

export interface TechniqueRef {
  technique_id: string
  name: string
  tactic: string
  event_count: number
}

export interface Narrative {
  id: number
  incident_id: number
  summary: string
  confidence: number
  sentences: Sentence[]
  model_used: string
  temperature: number
  tokens_used: number
  generation_time_ms: number
  created_at: string
  updated_at: string
}

export interface Sentence {
  text: string
  timestamp: string
  source_event_ids: number[]
  confidence: number
  technique?: string
}

export interface Event {
  id: number
  timestamp: string
  hostname: string
  event_type: string
  event_id: string
  user_name: string
  source_ip: string
  dest_ip: string
  process_name: string
  command_line: string
  parent_process: string
  log_type: string
  session_id: string
  department: string
  location: string
  device_type: string
  success: boolean
  port: string
  protocol: string
  file_path: string
  severity: string
  error: string
  raw_json: string
  created_at: string
}

export interface Feedback {
  id: number
  narrative_id: number
  incident_id: number
  rating: 'up' | 'down'
  notes: string
  created_at: string
}
```

### Dark Theme Tailwind Configuration
```css
/* app/globals.css */
@import "tailwindcss";

@theme inline {
  --color-background: #09090b;
  --color-foreground: #fafafa;
  --color-card: #18181b;
  --color-card-foreground: #fafafa;
  --color-muted: #27272a;
  --color-muted-foreground: #a1a1aa;
  --color-border: #27272a;
  --color-primary: #3b82f6;
  --color-primary-foreground: #fafafa;
  --color-destructive: #ef4444;
  --color-destructive-foreground: #fafafa;
  --color-success: #22c55e;
  --color-warning: #f59e0b;
  --font-sans: var(--font-geist-sans);
  --font-mono: var(--font-geist-mono);
}

body {
  background: var(--color-background);
  color: var(--color-foreground);
  font-family: var(--font-sans);
}

/* SOC analyst specific utilities */
.severity-low { @apply bg-green-900/30 text-green-400 border border-green-700; }
.severity-medium { @apply bg-yellow-900/30 text-yellow-400 border border-yellow-700; }
.severity-high { @apply bg-orange-900/30 text-orange-400 border border-orange-700; }
.severity-critical { @apply bg-red-900/30 text-red-400 border border-red-700; }
```

### Severity Badge Component
```tsx
// components/SeverityBadge.tsx
interface SeverityBadgeProps {
  severity: 'low' | 'medium' | 'high' | 'critical'
}

const severityStyles = {
  low: 'bg-green-900/30 text-green-400 border border-green-700',
  medium: 'bg-yellow-900/30 text-yellow-400 border border-yellow-700',
  high: 'bg-orange-900/30 text-orange-400 border border-orange-700',
  critical: 'bg-red-900/30 text-red-400 border border-red-700'
}

export default function SeverityBadge({ severity }: SeverityBadgeProps) {
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${severityStyles[severity]}`}>
      {severity.toUpperCase()}
    </span>
  )
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `tailwind.config.js` | `@theme inline` in CSS | Tailwind CSS 4 | Configuration moved to CSS |
| `getServerSideProps` | Server Components | Next.js 13+ | Data fetching in components |
| `useEffect` for data | SWR / React Query | 2020+ | Better caching, revalidation |
| Context API only | SWR for server state | 2020+ | Automatic revalidation |

**Deprecated/outdated:**
- `getServerSideProps` / `getStaticProps`: Replaced by Server Components in App Router
- `pages/_app.js`: Replaced by `app/layout.tsx`
- Manual `useEffect` + `useState` for data fetching: Use SWR or React Query

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Backend CORS is configured for localhost:3000 | Common Pitfalls | API calls fail - need to verify backend config |
| A2 | Narrative sentences are stored as JSON string in database | Code Examples | Need to parse JSON in frontend - verify with Phase 3 implementation |
| A3 | No authentication required (demo mode) | User Constraints | May need to add auth headers - verify with project requirements |
| A4 | SWR is recommended by Next.js 16 docs | Standard Stack | Need to verify if React Query would be better choice |

**If this table is empty:** All claims in this research were verified or cited — no user confirmation needed.

## Open Questions

1. **Feedback table schema**
   - What we know: Need narrative_id, incident_id, rating, notes, created_at
   - What's unclear: Should we add user_id column for future auth support?
   - Recommendation: Add user_id column with NULL default for forward compatibility

2. **Source event viewer placement**
   - What we know: Hover sentence should show raw events
   - What's unclear: Side panel vs modal vs inline expansion?
   - Recommendation: Side panel (right side) for desktop, modal for mobile - standard SOC dashboard pattern

3. **Incident detail page URL structure**
   - What we know: Need `/incidents/[id]` route
   - What's unclear: Should we use client-side routing or full page navigation?
   - Recommendation: Full page navigation with Next.js Link for better UX and SEO

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Node.js | Frontend build | ✓ | 25.9.0 | — |
| npm | Package management | ✓ | 11.12.1 | — |
| Go backend | API endpoints | ✓ | Running on :8080 | — |
| Tailwind CSS | Styling | ✓ | ^4 installed | — |

**Missing dependencies with no fallback:**
- None - all dependencies are available

**Missing dependencies with fallback:**
- None - all dependencies are available

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest (via Next.js 16) |
| Config file | next.config.ts (built-in) |
| Quick run command | `npm run build` |
| Full suite command | `npm run build && npm run lint` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DASH-01 | Incident list displays with severity | smoke | `npm run build` | ❌ Wave 0 |
| DASH-02 | Story cards show sentences with confidence | smoke | `npm run build` | ❌ Wave 0 |
| DASH-03 | Hover reveals source events | manual | Browser test | ❌ Wave 0 |
| DASH-04 | Feedback form submits | smoke | `npm run build` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `npm run build`
- **Per wave merge:** `npm run build && npm run lint`
- **Phase gate:** Full build green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `npm install swr lucide-react clsx react-hot-toast` — install dependencies
- [ ] `app/globals.css` — update with dark theme tokens
- [ ] `lib/types.ts` — create TypeScript interfaces
- [ ] `lib/api.ts` — create API client

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | Form validation for feedback notes |
| V6 Cryptography | no | No sensitive data handling |

### Known Threat Patterns for {stack}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| XSS via narrative text | Tampering | React auto-escaping + DOMPurify if needed |
| CSRF on feedback endpoint | Tampering | CSRF token or SameSite cookies |
| API key exposure | Information Disclosure | Environment variables only |

## Sources

### Primary (HIGH confidence)
- Next.js 16 docs (node_modules/next/dist/docs/) - Server Components, Fetching Data, CSS
- Tailwind CSS 4 docs - @theme inline configuration
- React 19 docs - Server Components, Suspense

### Secondary (MEDIUM confidence)
- SWR documentation - Client-side data fetching patterns
- Lucide React documentation - Icon library

### Tertiary (LOW confidence)
- SOC dashboard UX patterns - Dark theme conventions

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All packages verified via npm and official docs
- Architecture: HIGH - Based on Next.js 16 official patterns
- Pitfalls: HIGH - Common issues documented in Next.js docs

**Research date:** 2026-06-10
**Valid until:** 2026-07-10 (30 days for stable stack)
