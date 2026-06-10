import type {
  Incident,
  IncidentsResponse,
  IncidentResponse,
  Narrative,
  NarrativeResponse,
  Event,
  EventsResponse,
  IncidentEventsResponse,
  NarrativeSourceEventsResponse,
  IncidentStats,
  Feedback,
  FeedbackResponse,
  CreateFeedbackRequest,
  TechniquesResponse,
} from './types';

export const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('nexus_token');
}

async function fetchAPI<T>(url: string, options?: RequestInit): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(error.error || `HTTP ${res.status}`);
  }

  return res.json();
}

// Auth
export async function signup(email: string, password: string) {
  return fetchAPI<{ token: string; user: { id: string; email: string; api_key: string } }>(
    '/api/auth/signup',
    { method: 'POST', body: JSON.stringify({ email, password }) }
  );
}

export async function login(email: string, password: string) {
  return fetchAPI<{ token: string; user: { id: string; email: string; api_key: string } }>(
    '/api/auth/login',
    { method: 'POST', body: JSON.stringify({ email, password }) }
  );
}

export async function getMe() {
  return fetchAPI<{ user: { id: string; email: string; api_key: string } }>('/api/auth/me');
}

export async function getSettings() {
  return fetchAPI<{ openrouter_key: string; api_key: string }>('/api/auth/settings');
}

export async function updateSettings(openrouter_key: string) {
  return fetchAPI<{ message: string }>('/api/auth/settings', {
    method: 'PUT',
    body: JSON.stringify({ openrouter_key }),
  });
}

export function setToken(token: string) {
  localStorage.setItem('nexus_token', token);
}

export function clearToken() {
  localStorage.removeItem('nexus_token');
}

export function isAuthenticated(): boolean {
  return !!getToken();
}

// Incidents
export async function fetchIncidents(
  limit = 50,
  offset = 0,
  severity?: string,
  status?: string,
  sourceIP?: string
): Promise<IncidentsResponse> {
  const params = new URLSearchParams();
  params.set('limit', String(limit));
  params.set('offset', String(offset));
  if (severity) params.set('severity', severity);
  if (status) params.set('status', status);
  if (sourceIP) params.set('source_ip', sourceIP);
  return fetchAPI<IncidentsResponse>(`/api/incidents?${params.toString()}`);
}

export async function fetchIncident(id: number): Promise<IncidentResponse> {
  return fetchAPI<IncidentResponse>(`/api/incidents/${id}`);
}

export async function fetchIncidentEvents(id: number): Promise<IncidentEventsResponse> {
  return fetchAPI<IncidentEventsResponse>(`/api/incidents/${id}/events`);
}

export async function fetchIncidentStats(): Promise<IncidentStats> {
  return fetchAPI<IncidentStats>('/api/incidents/stats');
}

// Narratives
export async function fetchNarrative(incidentId: number): Promise<NarrativeResponse> {
  return fetchAPI<NarrativeResponse>(`/api/incidents/${incidentId}/narrative`);
}

export async function generateNarrative(incidentId: number): Promise<NarrativeResponse> {
  return fetchAPI<NarrativeResponse>(`/api/incidents/${incidentId}/narrative`, {
    method: 'POST',
  });
}

export async function fetchNarrativeSourceEvents(
  narrativeId: number
): Promise<NarrativeSourceEventsResponse> {
  return fetchAPI<NarrativeSourceEventsResponse>(`/api/narratives/${narrativeId}`);
}

// Feedback
export async function submitFeedback(data: CreateFeedbackRequest): Promise<{ feedback: Feedback }> {
  return fetchAPI<{ feedback: Feedback }>('/api/feedback', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getFeedback(narrativeId: number): Promise<FeedbackResponse> {
  return fetchAPI<FeedbackResponse>(`/api/feedback/${narrativeId}`);
}

// Techniques
export async function fetchTechniques(): Promise<TechniquesResponse> {
  return fetchAPI<TechniquesResponse>('/api/techniques');
}

// Stats
export async function fetchStats(): Promise<Record<string, unknown>> {
  return fetchAPI<Record<string, unknown>>('/api/stats');
}
