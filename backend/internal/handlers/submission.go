package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
	"github.com/liteoj/liteoj/backend/internal/services/judge"
)

type SubmissionHandler struct {
	DB     *gorm.DB
	C      *config.Config
	Queue  *judge.Queue
	Broker *events.Broker
}

type submitReq struct {
	Language     string `json:"language" binding:"required"`
	Code         string `json:"code" binding:"required"`
	ProblemSetID *uint  `json:"problemset_id,omitempty"`
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Param("id"))
	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}
	var p models.Problem
	if err := h.DB.First(&p, pid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrProblemNotFound})
		return
	}
	if !p.Visible && middleware.CurrentRole(c) != models.RoleAdmin {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrProblemNotFound})
		return
	}
	allowed := false
	for _, l := range h.C.JudgeLangs {
		if l == req.Language {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrLangNotAllowed})
		return
	}

	// Enforce per-problemset language whitelist when the client claims the
	// submission is happening inside a problem set. Empty allow-list = no
	// restriction; unknown problemset_id = silently ignore (student can't
	// forge a pass by pointing at a non-existent set).
	if req.ProblemSetID != nil && *req.ProblemSetID > 0 {
		var ps models.ProblemSet
		if err := h.DB.First(&ps, *req.ProblemSetID).Error; err == nil {
			// 必须是成员（或 admin）才能以题单上下文提交。
			if middleware.CurrentRole(c) != models.RoleAdmin {
				var n int64
				h.DB.Model(&models.ProblemSetMember{}).
					Where("problem_set_id = ? AND user_id = ?", *req.ProblemSetID, middleware.CurrentUserID(c)).
					Count(&n)
				if n == 0 {
					c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
					return
				}
			}
			allowedLangs := decodeAllowedLangs(ps.AllowedLangsJSON)
			if len(allowedLangs) > 0 {
				ok := false
				for _, l := range allowedLangs {
					if l == req.Language {
						ok = true
						break
					}
				}
				if !ok {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": i18n.ErrProblemsetLangBlock + req.Language,
					})
					return
				}
			}
		}
	}
	var tcs []models.Testcase
	h.DB.Where("problem_id = ?", p.ID).Order("order_index ASC, id ASC").Find(&tcs)
	if len(tcs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrNoTestData})
		return
	}

	sub := &models.Submission{
		UserID:       middleware.CurrentUserID(c),
		ProblemID:    p.ID,
		ProblemSetID: req.ProblemSetID,
		Language:     req.Language,
		Code:         req.Code,
		Verdict:      models.VerdictPending,
	}
	if err := h.DB.Create(sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cpu := p.TimeLimitMS
	if cpu == 0 {
		cpu = h.C.JudgeDefaultCPU
	}
	mem := p.MemoryLimitMB
	if mem == 0 {
		mem = h.C.JudgeDefaultMem
	}
	h.Queue.Enqueue(sub, tcs, cpu, mem)

	if h.Broker != nil {
		h.Broker.Publish(events.Event{
			Type: "submission:new",
			Data: map[string]any{
				"id":            sub.ID,
				"user_id":       sub.UserID,
				"problem_id":    sub.ProblemID,
				"problemset_id": sub.ProblemSetID,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"submission_id": sub.ID,
		"verdict":       models.VerdictPending,
	})
}

// List returns a paginated submission list. All authenticated users see every
// submission (to browse activity), but the `code` field is only exposed to
// admins. Students get meta (who/what/verdict/time/mem) and can click through
// to their own submission detail (code is gated there too).
func (h *SubmissionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	isAdmin := middleware.CurrentRole(c) == models.RoleAdmin
	q := h.DB.Model(&models.Submission{})
	if uid := c.Query("user_id"); uid != "" {
		q = q.Where("user_id = ?", uid)
	}
	if pid := c.Query("problem_id"); pid != "" {
		q = q.Where("problem_id = ?", pid)
	}
	if v := c.Query("verdict"); v != "" {
		q = q.Where("verdict = ?", v)
	}
	if psid := c.Query("problemset_id"); psid != "" {
		q = q.Where("problem_set_id = ?", psid)
	}
	if lang := c.Query("language"); lang != "" {
		q = q.Where("language = ?", lang)
	}
	if kw := c.Query("username"); kw != "" {
		// Match by display name OR login — 其实都转成 user_id IN (...) 避免 JOIN。
		like := "%" + kw + "%"
		var ids []uint
		h.DB.Model(&models.User{}).
			Where("name LIKE ? OR username LIKE ?", like, like).
			Pluck("id", &ids)
		if len(ids) == 0 {
			c.JSON(http.StatusOK, gin.H{"total": 0, "items": []any{}, "page": page, "page_size": pageSize})
			return
		}
		q = q.Where("user_id IN ?", ids)
	}

	var total int64
	q.Count(&total)
	var rows []models.Submission
	q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows)

	// Build username/name lookup — used for both admin and public views now.
	type uinfo struct{ Username, Name string }
	users := map[uint]uinfo{}
	if len(rows) > 0 {
		ids := make([]uint, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.UserID)
		}
		var us []models.User
		h.DB.Where("id IN ?", ids).Find(&us)
		for _, u := range us {
			users[u.ID] = uinfo{Username: u.Username, Name: u.Name}
		}
	}

	type summary struct {
		ID               uint   `json:"id"`
		UserID           uint   `json:"user_id"`
		Username         string `json:"username,omitempty"`
		Name             string `json:"name,omitempty"`
		ProblemID        uint   `json:"problem_id"`
		Language         string `json:"language"`
		Verdict          string `json:"verdict"`
		TimeUsedMS       int    `json:"time_used_ms"`
		MemoryKB         int    `json:"memory_used_kb"`
		CreatedAt        string `json:"created_at"`
		Code             string `json:"code,omitempty"`
		HasAIExplanation bool   `json:"has_ai_explanation"`
		AIPending        bool   `json:"ai_pending"`
		AIRejected       bool   `json:"ai_rejected"`
		AIRejectReason   string `json:"ai_reject_reason,omitempty"`
	}

	// 批量查"这一页提交里正在跑 AI 分析任务的是哪几条"——ai_tasks.subject 以
	// "submission #<id>" 形式存；一次 IN 查询取所有 running 行即可。
	aiPending := map[uint]bool{}
	if len(rows) > 0 {
		subjects := make([]string, 0, len(rows))
		for _, r := range rows {
			subjects = append(subjects, fmt.Sprintf("submission #%d", r.ID))
		}
		type tr struct{ Subject string }
		var trs []tr
		h.DB.Table("ai_tasks").
			Select("subject").
			Where("status = ? AND kind = ? AND subject IN ?",
				models.AITaskStatusRunning, models.AITaskKindAnalyze, subjects).
			Scan(&trs)
		for _, r := range trs {
			var id uint
			if _, err := fmt.Sscanf(r.Subject, "submission #%d", &id); err == nil {
				aiPending[id] = true
			}
		}
	}

	items := make([]summary, len(rows))
	for i, r := range rows {
		u := users[r.UserID]
		items[i] = summary{
			ID: r.ID, UserID: r.UserID, Username: u.Username, Name: u.Name,
			ProblemID: r.ProblemID, Language: r.Language, Verdict: r.Verdict,
			TimeUsedMS: r.TimeUsedMS, MemoryKB: r.MemoryUsedKB,
			CreatedAt:        r.CreatedAt.Format("2006-01-02 15:04:05"),
			HasAIExplanation: r.AIExplanation != "",
			AIPending:        aiPending[r.ID],
			AIRejected:       r.AIRejected,
			AIRejectReason:   r.AIRejectReason,
		}
		if isAdmin {
			items[i].Code = r.Code
		}
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": items, "page": page, "page_size": pageSize})
}

// Detail returns a single submission. Students may only view their own
// submissions (and see their own code). Admins see everything including the
// code field.
func (h *SubmissionHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var s models.Submission
	if err := h.DB.First(&s, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	if middleware.CurrentRole(c) != models.RoleAdmin && s.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
		return
	}
	// has_prev 告诉前端该用户在本题上是否已有更早的同语言提交——用于隐藏
	// "与上次对比"按钮。对比不同语言没有意义，因此要求 language 相同。
	var prevCount int64
	h.DB.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ? AND language = ? AND id < ?",
			s.UserID, s.ProblemID, s.Language, s.ID).
		Count(&prevCount)
	// ai_disabled：该提交所属题单禁用了 AI 解析时为 true，前端据此隐藏按钮；
	// 同时 Analyze 端点也会二次校验，防止客户端直接调 API 绕过。
	aiDisabled := false
	if s.ProblemSetID != nil && *s.ProblemSetID > 0 {
		var ps models.ProblemSet
		if err := h.DB.Select("disable_ai").First(&ps, *s.ProblemSetID).Error; err == nil {
			aiDisabled = ps.DisableAI
		}
	}
	type detailResp struct {
		models.Submission
		HasPrev    bool `json:"has_prev"`
		AIDisabled bool `json:"ai_disabled"`
	}
	c.JSON(http.StatusOK, detailResp{Submission: s, HasPrev: prevCount > 0, AIDisabled: aiDisabled})
}

// Diff returns two submissions' code for the frontend to render in a diff viewer.
func (h *SubmissionHandler) Diff(c *gin.Context) {
	aID, _ := strconv.Atoi(c.Param("id"))
	bID, _ := strconv.Atoi(c.Param("other"))
	var a, b models.Submission
	if err := h.DB.First(&a, aID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrSubmissionANotFound})
		return
	}
	role := middleware.CurrentRole(c)
	uid := middleware.CurrentUserID(c)
	if role != models.RoleAdmin && a.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
		return
	}
	if bID == 0 {
		// 默认对比上一次同语言提交——不同语言比 diff 没意义。
		h.DB.Where("user_id = ? AND problem_id = ? AND language = ? AND id < ?",
			a.UserID, a.ProblemID, a.Language, a.ID).
			Order("id DESC").First(&b)
	} else {
		if err := h.DB.First(&b, bID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrSubmissionBNotFound})
			return
		}
		if role != models.RoleAdmin && b.UserID != uid {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"a": a, "b": b})
}
