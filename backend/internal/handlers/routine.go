package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/davidyoung/savers-app/backend/internal/models"
)

type RoutineHandler struct {
	DB *sql.DB
}

func (h *RoutineHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var id, stepsJSON string
	var updatedAt time.Time
	err := h.DB.QueryRow(
		`SELECT id, steps_json, updated_at FROM routines WHERE user_id = ?`, uid,
	).Scan(&id, &stepsJSON, &updatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "未找到例行配置")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取失败")
		return
	}
	var steps []models.RoutineStep
	_ = json.Unmarshal([]byte(stepsJSON), &steps)
	writeJSON(w, http.StatusOK, models.Routine{
		ID: id, UserID: uid, Steps: steps, UpdatedAt: updatedAt,
	})
}

func (h *RoutineHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var body struct {
		Steps []models.RoutineStep `json:"steps"`
	}
	if err := decodeJSON(r, &body); err != nil || len(body.Steps) == 0 {
		writeError(w, http.StatusBadRequest, "请提供有效的步骤配置")
		return
	}
	for i := range body.Steps {
		if body.Steps[i].Duration < 30 {
			body.Steps[i].Duration = 30
		}
		if body.Steps[i].Duration > 7200 {
			body.Steps[i].Duration = 7200
		}
		body.Steps[i].Order = i + 1
	}
	b, _ := json.Marshal(body.Steps)
	now := time.Now().UTC()
	res, err := h.DB.Exec(
		`UPDATE routines SET steps_json = ?, updated_at = ? WHERE user_id = ?`,
		string(b), now, uid,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "更新失败")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "未找到例行配置")
		return
	}
	h.Get(w, r)
}
