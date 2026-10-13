package handlers

import (
	"database/sql"
	"net/http"

	"github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/davidyoung/savers-app/backend/internal/models"
	"github.com/go-chi/chi/v5"
)

type ContentHandler struct {
	DB *sql.DB
}

func (h *ContentHandler) List(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	stepKey := r.URL.Query().Get("step_key")
	q := `SELECT id, user_id, step_key, title, body, is_default, created_at
	      FROM content_items WHERE (user_id IS NULL OR user_id = ?)`
	args := []interface{}{uid}
	if stepKey != "" {
		q += ` AND step_key = ?`
		args = append(args, stepKey)
	}
	q += ` ORDER BY is_default DESC, created_at ASC`
	rows, err := h.DB.Query(q, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	items := []models.ContentItem{}
	for rows.Next() {
		var c models.ContentItem
		var userID sql.NullString
		var isDef int
		if err := rows.Scan(&c.ID, &userID, &c.StepKey, &c.Title, &c.Body, &isDef, &c.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "读取失败")
			return
		}
		if userID.Valid {
			c.UserID = &userID.String
		}
		c.IsDefault = isDef == 1
		items = append(items, c)
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *ContentHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	id := chi.URLParam(r, "id")
	var c models.ContentItem
	var userID sql.NullString
	var isDef int
	err := h.DB.QueryRow(
		`SELECT id, user_id, step_key, title, body, is_default, created_at FROM content_items
		 WHERE id = ? AND (user_id IS NULL OR user_id = ?)`, id, uid,
	).Scan(&c.ID, &userID, &c.StepKey, &c.Title, &c.Body, &isDef, &c.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "内容不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取失败")
		return
	}
	if userID.Valid {
		c.UserID = &userID.String
	}
	c.IsDefault = isDef == 1
	writeJSON(w, http.StatusOK, c)
}
