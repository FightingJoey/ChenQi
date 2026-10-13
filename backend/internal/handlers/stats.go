package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/davidyoung/savers-app/backend/internal/models"
)

type StatsHandler struct {
	DB *sql.DB
}

func (h *StatsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())

	rows, err := h.DB.Query(
		`SELECT date, status FROM sessions WHERE user_id = ? AND status = 'completed' ORDER BY date ASC`, uid,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	dates := []string{}
	for rows.Next() {
		var d, st string
		if err := rows.Scan(&d, &st); err != nil {
			writeError(w, http.StatusInternalServerError, "读取失败")
			return
		}
		dates = append(dates, d)
	}

	current, longest := computeStreaks(dates)
	var total int
	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM sessions WHERE user_id = ?`, uid).Scan(&total)
	completed := len(dates)
	rate := 0.0
	if total > 0 {
		rate = float64(completed) / float64(total)
	}

	stepRates := map[string]float64{}
	stepRows, err := h.DB.Query(
		`SELECT ss.key,
		        SUM(CASE WHEN ss.status = 'completed' THEN 1 ELSE 0 END) * 1.0 / COUNT(*),
		        COUNT(*)
		 FROM session_steps ss
		 JOIN sessions s ON s.id = ss.session_id
		 WHERE s.user_id = ?
		 GROUP BY ss.key`, uid,
	)
	if err == nil {
		defer stepRows.Close()
		for stepRows.Next() {
			var key string
			var rate float64
			var cnt int
			if err := stepRows.Scan(&key, &rate, &cnt); err == nil {
				stepRates[key] = rate
			}
		}
	}

	writeJSON(w, http.StatusOK, models.Stats{
		CurrentStreak:  current,
		LongestStreak:  longest,
		TotalSessions:  completed,
		CompletionRate: rate,
		StepRates:      stepRates,
	})
}

func computeStreaks(dates []string) (current, longest int) {
	if len(dates) == 0 {
		return 0, 0
	}
	set := map[string]bool{}
	for _, d := range dates {
		set[d] = true
	}
	parse := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}
	// longest
	run := 1
	longest = 1
	for i := 1; i < len(dates); i++ {
		prev := parse(dates[i-1])
		cur := parse(dates[i])
		if cur.Sub(prev) == 24*time.Hour {
			run++
			if run > longest {
				longest = run
			}
		} else if cur.Equal(prev) {
			continue
		} else {
			run = 1
		}
	}
	// current: walk back from today or yesterday
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	start := today
	if !set[today] {
		if set[yesterday] {
			start = yesterday
		} else {
			return 0, longest
		}
	}
	t := parse(start)
	for {
		if set[t.Format("2006-01-02")] {
			current++
			t = t.AddDate(0, 0, -1)
		} else {
			break
		}
	}
	return current, longest
}

func (h *StatsHandler) Calendar(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		now := time.Now()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		to = time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	rows, err := h.DB.Query(
		`SELECT date, status FROM sessions WHERE user_id = ? AND date >= ? AND date <= ?`,
		uid, from, to,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	days := []models.CalendarDay{}
	for rows.Next() {
		var d models.CalendarDay
		if err := rows.Scan(&d.Date, &d.Status); err != nil {
			writeError(w, http.StatusInternalServerError, "读取失败")
			return
		}
		days = append(days, d)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"from": from, "to": to, "days": days,
	})
}
