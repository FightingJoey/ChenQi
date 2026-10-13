package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type RoutineStep struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Duration int    `json:"duration_seconds"`
	Order    int    `json:"order"`
}

type Routine struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id"`
	Steps     []RoutineStep `json:"steps"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type Session struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Date        string         `json:"date"` // YYYY-MM-DD in user local sense
	Status      string         `json:"status"` // in_progress | completed | abandoned
	Mood        *int           `json:"mood,omitempty"`
	Note        string         `json:"note,omitempty"`
	StartedAt   time.Time      `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Steps       []SessionStep  `json:"steps,omitempty"`
}

type SessionStep struct {
	ID              string     `json:"id"`
	SessionID       string     `json:"session_id"`
	Key             string     `json:"key"`
	Name            string     `json:"name"`
	Order           int        `json:"order"`
	PlannedDuration int        `json:"planned_duration_seconds"`
	ActualDuration  int        `json:"actual_duration_seconds"`
	Status          string     `json:"status"` // pending | active | completed | skipped
	ContentID       *string    `json:"content_id,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type ContentItem struct {
	ID        string    `json:"id"`
	UserID    *string   `json:"user_id,omitempty"` // nil = system seed
	StepKey   string    `json:"step_key"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type Journal struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	SessionID *string   `json:"session_id,omitempty"`
	Prompt    string    `json:"prompt"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Stats struct {
	CurrentStreak   int            `json:"current_streak"`
	LongestStreak   int            `json:"longest_streak"`
	TotalSessions   int            `json:"total_sessions"`
	CompletionRate  float64        `json:"completion_rate"`
	StepRates       map[string]float64 `json:"step_rates"`
}

type CalendarDay struct {
	Date   string `json:"date"`
	Status string `json:"status"` // completed | in_progress | none
}

var DefaultSteps = []RoutineStep{
	{Key: "silence", Name: "静默", Enabled: true, Duration: 300, Order: 1},
	{Key: "affirmations", Name: "肯定语", Enabled: true, Duration: 300, Order: 2},
	{Key: "visualization", Name: "观想", Enabled: true, Duration: 300, Order: 3},
	{Key: "exercise", Name: "运动", Enabled: true, Duration: 600, Order: 4},
	{Key: "reading", Name: "阅读", Enabled: true, Duration: 600, Order: 5},
	{Key: "scribing", Name: "书写", Enabled: true, Duration: 300, Order: 6},
}
