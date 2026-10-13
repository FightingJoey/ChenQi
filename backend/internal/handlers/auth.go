package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/davidyoung/savers-app/backend/internal/auth"
	"github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/davidyoung/savers-app/backend/internal/models"
	"github.com/google/uuid"
)

type AuthHandler struct {
	DB *sql.DB
}

type registerReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Username == "" || req.Email == "" || len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "用户名/邮箱不能为空，密码至少 6 位")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "密码处理失败")
		return
	}
	id := uuid.NewString()
	now := time.Now().UTC()
	_, err = h.DB.Exec(
		`INSERT INTO users (id, username, email, password_hash, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, req.Username, req.Email, hash, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			writeError(w, http.StatusConflict, "用户名或邮箱已存在")
			return
		}
		writeError(w, http.StatusInternalServerError, "注册失败")
		return
	}
	steps, _ := json.Marshal(models.DefaultSteps)
	_, _ = h.DB.Exec(
		`INSERT INTO routines (id, user_id, steps_json, updated_at) VALUES (?, ?, ?, ?)`,
		uuid.NewString(), id, string(steps), now,
	)
	token, err := auth.IssueToken(id, req.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "签发令牌失败")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user": map[string]string{
			"id": id, "username": req.Username, "email": req.Email,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	var u models.User
	err := h.DB.QueryRow(
		`SELECT id, username, email, password_hash, created_at FROM users WHERE username = ? OR email = ?`,
		req.Username, req.Username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows || !auth.CheckPassword(u.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "登录失败")
		return
	}
	token, err := auth.IssueToken(u.ID, u.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "签发令牌失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]string{
			"id": u.ID, "username": u.Username, "email": u.Email,
		},
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var u models.User
	err := h.DB.QueryRow(
		`SELECT id, username, email, created_at FROM users WHERE id = ?`, uid,
	).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "用户不存在")
		return
	}
	writeJSON(w, http.StatusOK, u)
}
