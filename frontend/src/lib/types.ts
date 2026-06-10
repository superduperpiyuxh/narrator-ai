// TypeScript interfaces matching Go API response shapes exactly

export interface TechniqueRef {
  technique_id: string;
  name: string;
  tactic: string;
  event_count: number;
}

export interface Incident {
  id: number;
  title: string;
  description: string;
  source_ip: string;
  start_time: string;
  end_time: string;
  event_count: number;
  unique_users: string[];
  unique_ips: string[];
  unique_hostnames: string[];
  severity: 'low' | 'medium' | 'high' | 'critical';
  status: string;
  techniques: TechniqueRef[];
  tactics: string[];
  mitre_attack_ids: string[];
  confidence: number;
  raw_summary: string;
  created_at: string;
  updated_at: string;
}

export interface Narrative {
  id: number;
  incident_id: number;
  summary: string;
  confidence: number;
  sentences: string; // JSON string containing Sentence[]
  model_used: string;
  temperature: number;
  tokens_used: number;
  generation_time_ms: number;
  created_at: string;
  updated_at: string;
}

export interface Sentence {
  text: string;
  timestamp: string;
  source_event_ids: number[];
  confidence: number;
  technique?: string;
}

export interface Event {
  id: number;
  timestamp: string;
  hostname: string;
  event_type: string;
  event_id: string;
  user_name: string;
  source_ip: string;
  dest_ip: string;
  process_name: string;
  command_line: string;
  parent_process: string;
  log_type: string;
  session_id: string;
  department: string;
  location: string;
  device_type: string;
  success: boolean;
  port: string;
  protocol: string;
  file_path: string;
  severity: string;
  error: string;
  raw_json: Record<string, unknown>;
  created_at: string;
}

export interface Feedback {
  id: number;
  narrative_id: number;
  incident_id: number;
  rating: number; // -1 or 1
  notes: string;
  user_id: string;
  created_at: string;
}

export interface IncidentStats {
  total_incidents: number;
  by_severity: Record<string, number>;
  avg_events_per_incident: number;
}

// API Response types
export interface IncidentsResponse {
  incidents: Incident[];
  total: number;
  limit: number;
  offset: number;
}

export interface IncidentResponse {
  incident: Incident;
}

export interface NarrativeResponse {
  narrative: Narrative;
  cached?: boolean;
}

export interface EventsResponse {
  events: Event[];
  total: number;
}

export interface IncidentEventsResponse {
  events: Event[];
  total: number;
  incident_id: number;
}

export interface NarrativeSourceEventsResponse {
  events: Event[];
  total: number;
  narrative_id: number;
}

export interface FeedbackResponse {
  feedback: Feedback | null;
}

export interface CreateFeedbackRequest {
  narrative_id: number;
  incident_id: number;
  rating: number;
  notes?: string;
}

export interface Technique {
  id: string;
  name: string;
  description: string;
  tactic: string;
  url: string;
}

export interface TechniquesResponse {
  techniques: Technique[];
}
