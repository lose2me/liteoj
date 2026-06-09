package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

func (h *ProblemSetHandler) List(c *gin.Context) {
	var sets []models.ProblemSet
	q := h.DB.Order("id DESC")
	// 非 admin 只看到已开放的题单；admin 可见全部。与 Problem.visible 同语义。
	if middleware.CurrentRole(c) != models.RoleAdmin {
		q = q.Where("visible = ?", true)
	}
	if err := q.Find(&sets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Preaggregate item counts for every problem set in one query.
	countByPS := map[uint]int{}
	{
		type r struct {
			ProblemSetID uint
			N            int
		}
		var rows []r
		h.DB.Table("problem_set_items").
			Select("problem_set_id, COUNT(*) as n").
			Group("problem_set_id").Scan(&rows)
		for _, row := range rows {
			countByPS[row.ProblemSetID] = row.N
		}
	}

	// Per-user AC count across each problem set's items. Runs only when the
	// request is authenticated (every route under /problemsets is behind Auth,
	// so uid>0 in practice). AC via the standalone problem page does NOT
	// count here — only submissions tagged with this set's problemset_id.
	acByPS := map[uint]int{}
	memberByPS := map[uint]bool{}
	bannedByPS := map[uint]bool{}
	if uid := middleware.CurrentUserID(c); uid > 0 {
		type r struct {
			ProblemSetID uint
			N            int
		}
		var rows []r
		h.DB.Raw(`
			SELECT psi.problem_set_id, COUNT(DISTINCT psi.problem_id) AS n
			  FROM problem_set_items psi
			  JOIN submissions s
			    ON s.problem_id = psi.problem_id
			   AND s.problem_set_id = psi.problem_set_id
			   AND s.user_id = ?
			   AND s.verdict = 'AC'
			 WHERE NOT EXISTS (
			     SELECT 1 FROM problem_set_bans b
			      WHERE b.problem_set_id = psi.problem_set_id
			        AND b.user_id = ?
			   )
			 GROUP BY psi.problem_set_id`, uid, uid).Scan(&rows)
		for _, row := range rows {
			acByPS[row.ProblemSetID] = row.N
		}
		// 成员关系：一次查询拉当前用户所有加入记录。
		var mrows []struct{ ProblemSetID uint }
		h.DB.Table("problem_set_members").Where("user_id = ?", uid).Select("problem_set_id").Scan(&mrows)
		for _, m := range mrows {
			memberByPS[m.ProblemSetID] = true
		}
		// 封禁关系：同理，批量拉当前用户所有被踢记录。
		var brows []struct{ ProblemSetID uint }
		h.DB.Table("problem_set_bans").Where("user_id = ?", uid).Select("problem_set_id").Scan(&brows)
		for _, b := range brows {
			bannedByPS[b.ProblemSetID] = true
		}
	}

	type leaderInfo struct {
		name string
		ac   int
	}
	leaderByPS := make(map[uint]leaderInfo, len(sets))
	for _, s := range sets {
		name, ac := h.problemSetLeader(s.ID)
		leaderByPS[s.ID] = leaderInfo{name: name, ac: ac}
	}

	out := make([]psListRow, len(sets))
	for i, s := range sets {
		hasPwd := s.Password != ""
		s = sanitizeProblemSetForResponse(s)
		top := leaderByPS[s.ID]
		out[i] = psListRow{
			ProblemSet:  s,
			HasPassword: hasPwd,
			IsMember:    memberByPS[s.ID],
			IsBanned:    bannedByPS[s.ID],
			ItemCount:   countByPS[s.ID],
			MyACCount:   acByPS[s.ID],
			TopACCount:  top.ac,
			TopACName:   top.name,
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}
