package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/davidyoung/savers-app/backend/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ContentHandler struct {
	DB *sql.DB
}

var allowedStepKeys = map[string]bool{
	"silence": true, "affirmations": true, "visualization": true,
	"exercise": true, "reading": true, "scribing": true,
}

func scanContent(row interface {
	Scan(dest ...any) error
}) (models.ContentItem, error) {
	var c models.ContentItem
	var userID sql.NullString
	var isDef int
	err := row.Scan(&c.ID, &userID, &c.StepKey, &c.Title, &c.Body, &isDef, &c.CreatedAt)
	if err != nil {
		return c, err
	}
	if userID.Valid {
		c.UserID = &userID.String
	}
	c.IsDefault = isDef == 1
	return c, nil
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
	// User items first, then system defaults
	q += ` ORDER BY CASE WHEN user_id IS NULL THEN 1 ELSE 0 END, created_at DESC`
	rows, err := h.DB.Query(q, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	items := []models.ContentItem{}
	for rows.Next() {
		c, err := scanContent(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "读取失败")
			return
		}
		items = append(items, c)
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *ContentHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	id := chi.URLParam(r, "id")
	c, err := scanContent(h.DB.QueryRow(
		`SELECT id, user_id, step_key, title, body, is_default, created_at FROM content_items
		 WHERE id = ? AND (user_id IS NULL OR user_id = ?)`, id, uid,
	))
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "内容不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取失败")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *ContentHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var body struct {
		StepKey string `json:"step_key"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	body.StepKey = strings.TrimSpace(body.StepKey)
	body.Title = strings.TrimSpace(body.Title)
	body.Body = strings.TrimSpace(body.Body)
	if body.StepKey == "" {
		body.StepKey = "affirmations"
	}
	if !allowedStepKeys[body.StepKey] {
		writeError(w, http.StatusBadRequest, "无效的步骤类型")
		return
	}
	if body.Title == "" || body.Body == "" {
		writeError(w, http.StatusBadRequest, "标题和正文不能为空")
		return
	}
	id := uuid.NewString()
	now := time.Now().UTC()
	_, err := h.DB.Exec(
		`INSERT INTO content_items (id, user_id, step_key, title, body, is_default, created_at)
		 VALUES (?, ?, ?, ?, ?, 0, ?)`,
		id, uid, body.StepKey, body.Title, body.Body, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建失败")
		return
	}
	writeJSON(w, http.StatusCreated, models.ContentItem{
		ID: id, UserID: &uid, StepKey: body.StepKey,
		Title: body.Title, Body: body.Body, IsDefault: false, CreatedAt: now,
	})
}

func (h *ContentHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	id := chi.URLParam(r, "id")

	var owner sql.NullString
	err := h.DB.QueryRow(`SELECT user_id FROM content_items WHERE id = ?`, id).Scan(&owner)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "内容不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	if !owner.Valid {
		writeError(w, http.StatusForbidden, "系统内置内容不可修改，请复制后编辑")
		return
	}
	if owner.String != uid {
		writeError(w, http.StatusForbidden, "无权修改此内容")
		return
	}

	var body struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	body.Body = strings.TrimSpace(body.Body)
	if body.Title == "" || body.Body == "" {
		writeError(w, http.StatusBadRequest, "标题和正文不能为空")
		return
	}
	_, err = h.DB.Exec(
		`UPDATE content_items SET title = ?, body = ? WHERE id = ? AND user_id = ?`,
		body.Title, body.Body, id, uid,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}
	c, err := scanContent(h.DB.QueryRow(
		`SELECT id, user_id, step_key, title, body, is_default, created_at FROM content_items WHERE id = ?`, id,
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取失败")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *ContentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	id := chi.URLParam(r, "id")

	var owner sql.NullString
	err := h.DB.QueryRow(`SELECT user_id FROM content_items WHERE id = ?`, id).Scan(&owner)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "内容不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	if !owner.Valid {
		writeError(w, http.StatusForbidden, "系统内置内容不可删除")
		return
	}
	if owner.String != uid {
		writeError(w, http.StatusForbidden, "无权删除此内容")
		return
	}
	_, err = h.DB.Exec(`DELETE FROM content_items WHERE id = ? AND user_id = ?`, id, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true", "id": id})
}

// Clone copies a system (or any readable) item into a user-owned affirmation/content.
func (h *ContentHandler) Clone(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	id := chi.URLParam(r, "id")
	src, err := scanContent(h.DB.QueryRow(
		`SELECT id, user_id, step_key, title, body, is_default, created_at FROM content_items
		 WHERE id = ? AND (user_id IS NULL OR user_id = ?)`, id, uid,
	))
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "内容不存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取失败")
		return
	}
	newID := uuid.NewString()
	now := time.Now().UTC()
	title := src.Title
	if src.UserID == nil {
		title = src.Title + "（我的副本）"
	}
	_, err = h.DB.Exec(
		`INSERT INTO content_items (id, user_id, step_key, title, body, is_default, created_at)
		 VALUES (?, ?, ?, ?, ?, 0, ?)`,
		newID, uid, src.StepKey, title, src.Body, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "复制失败")
		return
	}
	writeJSON(w, http.StatusCreated, models.ContentItem{
		ID: newID, UserID: &uid, StepKey: src.StepKey,
		Title: title, Body: src.Body, IsDefault: false, CreatedAt: now,
	})
}
