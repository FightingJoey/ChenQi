package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/davidyoung/savers-app/backend/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SessionHandler struct {
	DB *sql.DB
}

func todayDate(r *http.Request) string {
	tz := r.URL.Query().Get("tz")
	if tz == "" {
		tz = r.Header.Get("X-Timezone")
	}
	loc := time.Local
	if tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}
	return time.Now().In(loc).Format("2006-01-02")
}

func (h *SessionHandler) loadSteps(sessionID string) ([]models.SessionStep, error) {
	rows, err := h.DB.Query(
		`SELECT id, session_id, key, name, step_order, planned_duration_seconds, actual_duration_seconds,
		        status, content_id, started_at, completed_at
		 FROM session_steps WHERE session_id = ? ORDER BY step_order ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	steps := []models.SessionStep{}
	for rows.Next() {
		var s models.SessionStep
		var contentID sql.NullString
		var started, completed sql.NullTime
		if err := rows.Scan(&s.ID, &s.SessionID, &s.Key, &s.Name, &s.Order, &s.PlannedDuration,
			&s.ActualDuration, &s.Status, &contentID, &started, &completed); err != nil {
			return nil, err
		}
		if contentID.Valid {
			s.ContentID = &contentID.String
		}
		if started.Valid {
			s.StartedAt = &started.Time
		}
		if completed.Valid {
			s.CompletedAt = &completed.Time
		}
		steps = append(steps, s)
	}
	return steps, nil
}

func (h *SessionHandler) getSessionByUserDate(uid, date string) (*models.Session, error) {
	var s models.Session
	var mood sql.NullInt64
	var completed sql.NullTime
	err := h.DB.QueryRow(
		`SELECT id, user_id, date, status, mood, note, started_at, completed_at
		 FROM sessions WHERE user_id = ? AND date = ?`, uid, date,
	).Scan(&s.ID, &s.UserID, &s.Date, &s.Status, &mood, &s.Note, &s.StartedAt, &completed)
	if err != nil {
		return nil, err
	}
	if mood.Valid {
		m := int(mood.Int64)
		s.Mood = &m
	}
	if completed.Valid {
		s.CompletedAt = &completed.Time
	}
	steps, err := h.loadSteps(s.ID)
	if err != nil {
		return nil, err
	}
	s.Steps = steps
	return &s, nil
}

func (h *SessionHandler) Today(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	date := todayDate(r)
	s, err := h.getSessionByUserDate(uid, date)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"date": date, "session": nil,
		})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"date": date, "session": s,
	})
}

func (h *SessionHandler) Start(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	date := todayDate(r)

	existing, err := h.getSessionByUserDate(uid, date)
	if err == nil && existing != nil {
		writeJSON(w, http.StatusOK, existing)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	var stepsJSON string
	err = h.DB.QueryRow(`SELECT steps_json FROM routines WHERE user_id = ?`, uid).Scan(&stepsJSON)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取例行配置失败")
		return
	}
	var routineSteps []models.RoutineStep
	_ = json.Unmarshal([]byte(stepsJSON), &routineSteps)

	now := time.Now().UTC()
	sid := uuid.NewString()
	tx, err := h.DB.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "开启事务失败")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO sessions (id, user_id, date, status, note, started_at) VALUES (?, ?, ?, 'in_progress', '', ?)`,
		sid, uid, date, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建会话失败")
		return
	}

	order := 0
	for _, rs := range routineSteps {
		if !rs.Enabled {
			continue
		}
		order++
		_, err = tx.Exec(
			`INSERT INTO session_steps (id, session_id, key, name, step_order, planned_duration_seconds, actual_duration_seconds, status)
			 VALUES (?, ?, ?, ?, ?, ?, 0, 'pending')`,
			uuid.NewString(), sid, rs.Key, rs.Name, order, rs.Duration,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "创建步骤失败")
			return
		}
	}
	if order == 0 {
		writeError(w, http.StatusBadRequest, "请至少启用一个 SAVERS 步骤")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "提交失败")
		return
	}
	s, err := h.getSessionByUserDate(uid, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取会话失败")
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *SessionHandler) UpdateStep(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	sessionID := chi.URLParam(r, "id")
	stepKey := chi.URLParam(r, "key")

	var owner string
	err := h.DB.QueryRow(`SELECT user_id FROM sessions WHERE id = ?`, sessionID).Scan(&owner)
	if err == sql.ErrNoRows || owner != uid {
		writeError(w, http.StatusNotFound, "会话不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	var body struct {
		Status         string  `json:"status"`
		ActualDuration *int    `json:"actual_duration_seconds"`
		ContentID      *string `json:"content_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	allowed := map[string]bool{"active": true, "completed": true, "skipped": true, "pending": true}
	if !allowed[body.Status] {
		writeError(w, http.StatusBadRequest, "无效状态")
		return
	}

	now := time.Now().UTC()
	var startedAt, completedAt interface{}
	if body.Status == "active" {
		startedAt = now
	}
	if body.Status == "completed" || body.Status == "skipped" {
		completedAt = now
	}

	actual := 0
	if body.ActualDuration != nil {
		actual = *body.ActualDuration
	}

	res, err := h.DB.Exec(
		`UPDATE session_steps SET status = ?, actual_duration_seconds = COALESCE(?, actual_duration_seconds),
		 content_id = COALESCE(?, content_id),
		 started_at = COALESCE(started_at, ?),
		 completed_at = COALESCE(?, completed_at)
		 WHERE session_id = ? AND key = ?`,
		body.Status, actual, body.ContentID, startedAt, completedAt, sessionID, stepKey,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新步骤失败")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "步骤不存在")
		return
	}

	steps, err := h.loadSteps(sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取失败")
		return
	}
	var updated *models.SessionStep
	for i := range steps {
		if steps[i].Key == stepKey {
			updated = &steps[i]
			break
		}
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *SessionHandler) Complete(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	sessionID := chi.URLParam(r, "id")

	var owner, status string
	err := h.DB.QueryRow(`SELECT user_id, status FROM sessions WHERE id = ?`, sessionID).Scan(&owner, &status)
	if err == sql.ErrNoRows || owner != uid {
		writeError(w, http.StatusNotFound, "会话不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	var body struct {
		Mood *int   `json:"mood"`
		Note string `json:"note"`
	}
	_ = decodeJSON(r, &body)

	now := time.Now().UTC()
	_, err = h.DB.Exec(
		`UPDATE sessions SET status = 'completed', mood = ?, note = ?, completed_at = ? WHERE id = ?`,
		body.Mood, body.Note, now, sessionID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "完成会话失败")
		return
	}
	// mark remaining pending as skipped
	_, _ = h.DB.Exec(
		`UPDATE session_steps SET status = 'skipped', completed_at = ? WHERE session_id = ? AND status IN ('pending','active')`,
		now, sessionID,
	)

	s, err := h.getSessionByUserDate(uid, "")
	// reload by id
	var date string
	_ = h.DB.QueryRow(`SELECT date FROM sessions WHERE id = ?`, sessionID).Scan(&date)
	s, err = h.getSessionByUserDate(uid, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取失败")
		return
	}
	writeJSON(w, http.StatusOK, s)
}
