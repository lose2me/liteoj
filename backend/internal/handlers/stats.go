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
