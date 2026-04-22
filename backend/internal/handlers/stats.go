package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type StatsHandler struct {
	DB *gorm.DB
}

// Stats returns verdict distribution and summary counters for the current user.
func (h *StatsHandler) Stats(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	type row struct {
		Verdict string
		Count   int
	}
	var rows []row
	h.DB.Raw(`SELECT verdict, COUNT(*) AS count FROM submissions WHERE user_id = ? GROUP BY verdict`, uid).Scan(&rows)
	dist := map[string]int{}
	var total int
	for _, r := range rows {
		dist[r.Verdict] = r.Count
		total += r.Count
	}
	// AC / Tried 按 (problem, context) 去重：题单内外相互独立，同题在不同题单
	// AC 各算一次。COALESCE(problem_set_id, 0) 把独立页当作 "context=0"。
	var distinctAC int64
	h.DB.Raw(`
		SELECT COUNT(*) FROM (
		  SELECT DISTINCT problem_id, COALESCE(problem_set_id, 0) AS ctx
		    FROM submissions WHERE user_id = ? AND verdict = 'AC'
		) t`, uid).Scan(&distinctAC)
	var distinctTried int64
	h.DB.Raw(`
		SELECT COUNT(*) FROM (
		  SELECT DISTINCT problem_id, COALESCE(problem_set_id, 0) AS ctx
		    FROM submissions WHERE user_id = ?
		) t`, uid).Scan(&distinctTried)
	acRate := 0.0
	if total > 0 {
		acRate = float64(dist[models.VerdictAC]) / float64(total)
	}
	// AK: reuse the shared aggregator so the personal-center number matches
	// whatever the ranking page shows (same "in-set AC" rule, no drift).
	ak := computeAKPerUser(h.DB, time.Time{})[uid]
	c.JSON(http.StatusOK, gin.H{
		"total_submissions": total,
		"distinct_ac":       distinctAC,
		"distinct_tried":    distinctTried,
		"ac_rate":           acRate,
		"distribution":      dist,
		"ak":                ak,
	})
}

// Contribution returns the current user's per-day submission / AC counts over
// the last 365 days, GitHub-heatmap style. SQLite-only: uses `strftime` to
// normalize created_at into a YYYY-MM-DD bucket. When the frontend draws a
// 52×7 grid for the past year, days with zero submissions simply stay absent
// and render as empty cells. PG support is tracked as follow-up.
func (h *StatsHandler) Contribution(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	since := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	type bucket struct {
		Day   string `json:"date"`
		Total int    `json:"count"`
		AC    int    `json:"ac"`
	}
	var items []bucket
	h.DB.Raw(`
		SELECT strftime('%Y-%m-%d', created_at) AS day,
		       COUNT(*)                                               AS total,
		       SUM(CASE WHEN verdict = 'AC' THEN 1 ELSE 0 END)         AS ac
		  FROM submissions
		 WHERE user_id = ?
		   AND created_at >= ?
		 GROUP BY day
		 ORDER BY day ASC`, uid, since).Scan(&items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}
