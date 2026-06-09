package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

func (h *ProblemSetHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	uid := middleware.CurrentUserID(c)
	isAdmin := middleware.CurrentRole(c) == models.RoleAdmin
	access, err := loadProblemSetAccess(h.DB, uint(id), uid, isAdmin, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !access.Found || (!isAdmin && !access.VisibleToStudent()) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}

	hasPwd := access.HasPassword()
	ps := sanitizeProblemSetForResponse(access.ProblemSet)

	// 非成员且非 admin：返回元信息 + locked，不暴露题目列表，但仍告知题目总数
	// 和榜首——这些在 List 接口已对全员可见，没必要再隐藏。前端据此把非成员
	// 视图渲染得跟成员页一致，仅少掉三个 tab。
	if !access.IsMember {
		itemCount, topName, topAC := h.problemSetHeadline(ps.ID)
		c.JSON(http.StatusOK, gin.H{
			"problemset":   ps,
			"has_password": hasPwd,
			"is_member":    false,
			"is_banned":    access.IsBanned,
			"locked":       true,
			"lock_reason":  access.StudentLockReason(),
			"item_count":   itemCount,
			"top_ac_name":  topName,
			"top_ac_count": topAC,
		})
		return
	}

	var items []models.ProblemSetItem
	h.DB.Where("problem_set_id = ?", ps.ID).Order("order_index ASC, id ASC").Find(&items)
	ids := make([]uint, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ProblemID)
	}
	var problems []models.Problem
	if len(ids) > 0 {
		h.DB.Where("id IN ?", ids).Find(&problems)
	}
	byID := make(map[uint]models.Problem, len(problems))
	for _, p := range problems {
		byID[p.ID] = p
	}
	// Problems are returned with a per-set `code` (A, B, C…) derived from
	// order_index. The global problem.ID stays authoritative for routing —
	// `code` is purely a display label so students see ICPC-style labels
	// inside a set rather than the raw DB id.
	type problemOut struct {
		models.Problem
		Code string `json:"code"`
	}
	ordered := make([]problemOut, 0, len(ids))
	for i, pid := range ids {
		if p, ok := byID[pid]; ok {
			ordered = append(ordered, problemOut{Problem: p, Code: problemCode(i)})
		}
	}

	mystatus := map[uint]string{}
	if uid > 0 && len(ids) > 0 {
		// Per-problem status must be computed from submissions tagged with
		// THIS problem_set_id — AC from the standalone problem page must not
		// leak into the set's progress.
		type row struct {
			ProblemID uint
			Verdict   string
		}
		var rows []row
		h.DB.Table("submissions").Select("problem_id, verdict").
			Where("user_id = ? AND problem_set_id = ? AND problem_id IN ?", uid, ps.ID, ids).
			Find(&rows)
		for _, r := range rows {
			if r.Verdict == models.VerdictAC {
				mystatus[r.ProblemID] = "AC"
			} else if mystatus[r.ProblemID] == "" {
				mystatus[r.ProblemID] = "attempted"
			}
		}
	}

	topName, topAC := h.problemSetLeader(ps.ID)

	c.JSON(http.StatusOK, gin.H{
		"problemset":   ps,
		"problems":     ordered,
		"my_status":    mystatus,
		"has_password": hasPwd,
		"is_member":    true,
		"top_ac_name":  topName,
		"top_ac_count": topAC,
		"locked":       false,
	})
}

// problemSetHeadline 返回非敏感的"门面"统计：题目数量、当前榜首姓名和其 AC
// 数。用于非成员 Detail / List 这类无需暴露题目明细但想展示概况的场景。
// 无榜首时 topName 为空字符串，topAC 为 0。
func (h *ProblemSetHandler) problemSetHeadline(psID uint) (itemCount int, topName string, topAC int) {
	var ic int64
	h.DB.Model(&models.ProblemSetItem{}).Where("problem_set_id = ?", psID).Count(&ic)
	itemCount = int(ic)
	topName, topAC = h.problemSetLeader(psID)
	return
}

// problemSetLeader mirrors the per-problemset ranking rule used on the ranking
// page: AC desc, then penalty asc. Bonus is display-only and does not affect
// who appears as the headline leader.
func (h *ProblemSetHandler) problemSetLeader(psID uint) (topName string, topAC int) {
	perUser, err := loadProblemSetRankRows(h.DB, psID, time.Time{})
	if err != nil {
		return
	}
	best, ok := bestProblemSetRankRow(perUser)
	if !ok {
		return
	}

	topAC = best.ACCount
	var u models.User
	if err := h.DB.First(&u, best.UserID).Error; err == nil {
		if u.Name != "" {
			topName = u.Name
		} else {
			topName = u.Username
		}
	}
	return
}
