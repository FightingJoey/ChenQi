export interface User {
  id: string
  username: string
  email: string
}

export interface RoutineStep {
  key: string
  name: string
  enabled: boolean
  duration_seconds: number
  order: number
}

export interface Routine {
  id: string
  user_id: string
  steps: RoutineStep[]
  updated_at: string
}

export interface SessionStep {
  id: string
  session_id: string
  key: string
  name: string
  order: number
  planned_duration_seconds: number
  actual_duration_seconds: number
  status: 'pending' | 'active' | 'completed' | 'skipped'
  content_id?: string
  started_at?: string
  completed_at?: string
}

export interface Session {
  id: string
  user_id: string
  date: string
  status: 'in_progress' | 'completed' | 'abandoned'
  mood?: number
  note?: string
  started_at: string
  completed_at?: string
  steps: SessionStep[]
}

export interface ContentItem {
  id: string
  user_id?: string | null
  step_key: string
  title: string
  body: string
  is_default: boolean
  created_at?: string
}

export interface Journal {
  id: string
  session_id?: string
  prompt: string
  body: string
  created_at: string
}

export interface Stats {
  current_streak: number
  longest_streak: number
  total_sessions: number
  completion_rate: number
  step_rates: Record<string, number>
}

export interface CalendarDay {
  date: string
  status: string
}
