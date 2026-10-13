package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/davidyoung/savers-app/backend/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type JournalHandler struct {
	DB *sql.DB
}

func (h *JournalHandler) List(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	rows, err := h.DB.Query(
		`SELECT id, user_id, session_id, prompt, body, created_at, updated_at
		 FROM journals WHERE user_id = ? ORDER BY created_at DESC LIMIT 50`, uid,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	list := []models.Journal{}
	for rows.Next() {
		var j models.Journal
		var sid sql.NullString
		if err := rows.Scan(&j.ID, &j.UserID, &sid, &j.Prompt, &j.Body, &j.CreatedAt, &j.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "读取失败")
			return
		}
		if sid.Valid {
			j.SessionID = &sid.String
		}
		list = append(list, j)
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *JournalHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var body struct {
		SessionID *string `json:"session_id"`
		Prompt    string  `json:"prompt"`
		Body      string  `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	now := time.Now().UTC()
	id := uuid.NewString()
	_, err := h.DB.Exec(
		`INSERT INTO journals (id, user_id, session_id, prompt, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, uid, body.SessionID, body.Prompt, body.Body, now, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "保存失败")
		return
	}
	writeJSON(w, http.StatusCreated, models.Journal{
		ID: id, UserID: uid, SessionID: body.SessionID,
		Prompt: body.Prompt, Body: body.Body, CreatedAt: now, UpdatedAt: now,
	})
}

func (h *JournalHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	id := chi.URLParam(r, "id")
	var body struct {
		Prompt string `json:"prompt"`
		Body   string `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	now := time.Now().UTC()
	res, err := h.DB.Exec(
		`UPDATE journals SET prompt = ?, body = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		body.Prompt, body.Body, now, id, uid,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "日记不存在")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"id": id, "prompt": body.Prompt, "body": body.Body, "updated_at": now})
}
