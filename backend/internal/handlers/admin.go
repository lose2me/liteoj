package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/auth"
	"github.com/liteoj/liteoj/backend/internal/cache"
	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type AdminHandler struct {
	DB     *gorm.DB
	C      *config.Config
	Cache  *cache.Cache
	Broker *events.Broker
}

// publishPS / publishProblem / publishMembers 是 SSE 广播的轻薄包装——让各个
// 写操作 handler 在成功后一行把事件推出去。前端据此对相应页面做重拉。
func (h *AdminHandler) publishPS(psid uint) {
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "problemset:changed", Data: map[string]any{"id": psid}})
	}
}
func (h *AdminHandler) publishProblem(pid uint) {
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "problem:changed", Data: map[string]any{"id": pid}})
	}
}
func (h *AdminHandler) publishMembers(psid uint, uid uint) {
	if h.Broker != nil {
		h.Broker.Publish(events.Event{
			Type: "problemset:members:changed",
			Data: map[string]any{"id": psid, "user_id": uid},
		})
	}
}

// ---------- Users ----------

type userUpsertReq struct {
	Username string      `json:"username"`
	Name     string      `json:"name"`
	Password string      `json:"password"`
	Role     models.Role `json:"role"`
}

// userListRow extends the bare User record with per-user submission stats
// that the admin console table surfaces (AC 数 / 尝试 / 总提交 / AC 率 / AK).
// AK reuses the same aggregator that powers the ranking page so both screens
// agree on "did the user AK this set".
type userListRow struct {
	models.User
	DistinctAC  int     `json:"distinct_ac"`
	DistinctTry int     `json:"distinct_tried"`
	TotalSubs   int     `json:"total_submissions"`
	ACRate      float64 `json:"ac_rate"`
	AK          int     `json:"ak"`
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []models.User
	q := h.DB.Model(&models.User{})
	if r := c.Query("role"); r != "" {
		q = q.Where("role = ?", r)
	}
	if kw := c.Query("q"); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("name LIKE ? OR username LIKE ?", like, like)
	}
	q.Order("id ASC").Find(&users)

	// One aggregate query covers total / ac_sub / distinct_tried / distinct_ac
	// for every user in one pass. Running this unbounded is fine: users table
	// is small (school / club scale) and the backing index on submissions.user_id
	// handles the GROUP BY cheaply.
	type agg struct {
		UserID uint
		Total  int
		ACSub  int
		Tried  int
		ACProb int
	}
	var aggs []agg
	h.DB.Raw(`
		SELECT user_id,
		       COUNT(*)                                                   AS total,
		       SUM(CASE WHEN verdict = 'AC' THEN 1 ELSE 0 END)             AS ac_sub,
		       COUNT(DISTINCT problem_id)                                  AS tried,
		       COUNT(DISTINCT CASE WHEN verdict = 'AC' THEN problem_id END) AS ac_prob
		  FROM submissions
		 GROUP BY user_id`).Scan(&aggs)
	aggByUser := make(map[uint]agg, len(aggs))
	for _, a := range aggs {
		aggByUser[a.UserID] = a
	}

	akPerUser := computeAKPerUser(h.DB, time.Time{})

	out := make([]userListRow, len(users))
	for i, u := range users {
		a := aggByUser[u.ID]
		rate := 0.0
		if a.Total > 0 {
			rate = float64(a.ACSub) / float64(a.Total)
		}
		out[i] = userListRow{
			User:        u,
			DistinctAC:  a.ACProb,
			DistinctTry: a.Tried,
			TotalSubs:   a.Total,
			ACRate:      rate,
			AK:          akPerUser[u.ID],
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

// UserProfile returns a single user's profile + submission stats for the admin
// "view other user" page. Mirrors the shape of GET /me/stats (so the frontend
// can reuse the same cards) but takes the uid from the URL instead of session.
func (h *AdminHandler) UserProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var u models.User
	if err := h.DB.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	uid := u.ID
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
	var distinctAC int64
	h.DB.Table("submissions").Where("user_id = ? AND verdict = ?", uid, models.VerdictAC).
		Distinct("problem_id").Count(&distinctAC)
	var distinctTried int64
	h.DB.Table("submissions").Where("user_id = ?", uid).
		Distinct("problem_id").Count(&distinctTried)
	acRate := 0.0
	if total > 0 {
		acRate = float64(dist[models.VerdictAC]) / float64(total)
	}
	ak := computeAKPerUser(h.DB, time.Time{})[uid]
	c.JSON(http.StatusOK, gin.H{
		"user":              u,
		"total_submissions": total,
		"distinct_ac":       distinctAC,
		"distinct_tried":    distinctTried,
		"ac_rate":           acRate,
		"distribution":      dist,
		"ak":                ak,
	})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req userUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}
	if req.Role == "" {
		req.Role = models.RoleStudent
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u := &models.User{Username: req.Username, Name: req.Name, PasswordHash: hash, Role: req.Role}
	if err := h.DB.Create(u).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req userUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		updates["password_hash"] = hash
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- Problems ----------

type problemUpsertReq struct {
	models.Problem
	TagIDs []uint `json:"tag_ids"`
}

func (h *AdminHandler) CreateProblem(c *gin.Context) {
	var r problemUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := r.Problem
	if p.TimeLimitMS == 0 {
		p.TimeLimitMS = h.C.JudgeDefaultCPU
	}
	if p.MemoryLimitMB == 0 {
		p.MemoryLimitMB = h.C.JudgeDefaultMem
	}
	p.TagsJSON = serializeTagsJSON(r.TagIDs)
	if err := h.DB.Create(&p).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.Cache != nil {
		h.Cache.Invalidate("problems:")
	}
	h.publishProblem(p.ID)
	c.JSON(http.StatusOK, p)
}

func (h *AdminHandler) UpdateProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var r problemUpsertReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := r.Problem
	p.ID = uint(id)
	p.TagsJSON = serializeTagsJSON(r.TagIDs)
	if err := h.DB.Model(&models.Problem{}).Where("id = ?", id).Updates(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.Cache != nil {
		h.Cache.Invalidate("problems:")
	}
	// 题目改过后，先前生成的 AI 解析可能跟新版题面/用例不一致——一键清空该
	// 题所有 submissions 的 AI 字段（ai_explanation / ai_rejected /
	// ai_reject_reason），让学生重新点"AI 解析"触发新一轮。verdict 保持原
	// 样——AC 历史不动。
	h.DB.Model(&models.Submission{}).Where("problem_id = ?", id).Updates(map[string]any{
		"ai_explanation":   "",
		"ai_rejected":      false,
		"ai_reject_reason": "",
	})
	h.publishProblem(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Problem{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Where("problem_id = ?", id).Delete(&models.Testcase{})
	// 级联清理所有题单里对该题的引用：原题删后题单内自然不再展示。
	h.DB.Where("problem_id = ?", id).Delete(&models.ProblemSetItem{})
	if h.Cache != nil {
		h.Cache.Invalidate("problems:")
	}
	h.publishProblem(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- Testcases ----------

func (h *AdminHandler) ListTestcases(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Param("id"))
	var tcs []models.Testcase
	h.DB.Where("problem_id = ?", pid).Order("order_index ASC, id ASC").Find(&tcs)
	c.JSON(http.StatusOK, gin.H{"items": tcs})
}

func (h *AdminHandler) CreateTestcase(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Param("id"))
	var tc models.Testcase
	if err := c.ShouldBindJSON(&tc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tc.ProblemID = uint(pid)
	if err := h.DB.Create(&tc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishProblem(uint(pid))
	c.JSON(http.StatusOK, tc)
}

func (h *AdminHandler) UpdateTestcase(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("tcid"))
	var tc models.Testcase
	if err := c.ShouldBindJSON(&tc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tc.ID = uint(id)
	if err := h.DB.Model(&models.Testcase{}).Where("id = ?", id).Updates(&tc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 用例更新后拿回 ProblemID 以广播事件——前端收到后刷新题目详情/列表。
	var reloaded models.Testcase
	if err := h.DB.Select("problem_id").First(&reloaded, id).Error; err == nil {
		h.publishProblem(reloaded.ProblemID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteTestcase(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("tcid"))
	// 删之前先读 problem_id 以便事件带出。
	var existing models.Testcase
	h.DB.Select("problem_id").First(&existing, id)
	if err := h.DB.Delete(&models.Testcase{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing.ProblemID > 0 {
		h.publishProblem(existing.ProblemID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- Problem sets ----------

func (h *AdminHandler) CreateProblemSet(c *gin.Context) {
	var ps models.ProblemSet
	if err := c.ShouldBindJSON(&ps); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ps.AllowedLangsJSON = encodeAllowedLangs(ps.AllowedLangs)
	if err := h.DB.Create(&ps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishPS(ps.ID)
	c.JSON(http.StatusOK, ps)
}

func (h *AdminHandler) UpdateProblemSet(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ps models.ProblemSet
	if err := c.ShouldBindJSON(&ps); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Use an explicit field map so that clearing allowed_langs (→ "") and
	// zero-value fields (empty password, false booleans) actually persist;
	// Updates with a struct skips Go zero values.
	updates := map[string]any{
		"title":              ps.Title,
		"password":           ps.Password,
		"start_time":         ps.StartTime,
		"end_time":           ps.EndTime,
		"allowed_langs_json": encodeAllowedLangs(ps.AllowedLangs),
		"visible":            ps.Visible,
		"disable_idea":       ps.DisableIdea,
		"disable_solution":   ps.DisableSolution,
		"disable_ai":         ps.DisableAI,
	}
	if err := h.DB.Model(&models.ProblemSet{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AdminHandler) DeleteProblemSet(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.ProblemSet{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetItem{})
	h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetMember{})
	h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetBan{})
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListProblemSetMembers 列出某题单的全部成员，admin 专用。
func (h *AdminHandler) ListProblemSetMembers(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	var members []models.ProblemSetMember
	if err := h.DB.Where("problem_set_id = ?", psid).Order("joined_at DESC").Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(members) == 0 {
		c.JSON(http.StatusOK, gin.H{"items": []any{}})
		return
	}
	uids := make([]uint, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UserID)
	}
	var users []models.User
	h.DB.Where("id IN ?", uids).Find(&users)
	byID := map[uint]models.User{}
	for _, u := range users {
		byID[u.ID] = u
	}
	type row struct {
		UserID   uint      `json:"user_id"`
		Username string    `json:"username"`
		Name     string    `json:"name"`
		JoinedAt time.Time `json:"joined_at"`
	}
	out := make([]row, 0, len(members))
	for _, m := range members {
		u := byID[m.UserID]
		out = append(out, row{UserID: m.UserID, Username: u.Username, Name: u.Name, JoinedAt: m.JoinedAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

// RemoveProblemSetMember 管理员把某用户踢出题单：删成员关系 + 写入封禁。
// **不删** submissions / ai_tasks——/submissions 全站列表仍需看到他的历史，
// 题单层面的"消失"通过 ranking / problemset list 的 anti-join bans 实现
// （见 handlers/ranking.go 与 handlers/problemset.go 里的 NOT EXISTS 子句）。
// 幂等：若已在 ban 表则只更新不重复插入。
func (h *AdminHandler) RemoveProblemSetMember(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	uid, _ := strconv.Atoi(c.Param("uid"))
	actor := middleware.CurrentUserID(c)
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_set_id = ? AND user_id = ?", psid, uid).
			Delete(&models.ProblemSetMember{}).Error; err != nil {
			return err
		}
		// 写入封禁记录（幂等）。
		var existing int64
		tx.Model(&models.ProblemSetBan{}).
			Where("problem_set_id = ? AND user_id = ?", psid, uid).Count(&existing)
		if existing == 0 {
			ban := models.ProblemSetBan{
				ProblemSetID: uint(psid),
				UserID:       uint(uid),
				BannedAt:     time.Now(),
				BannedBy:     actor,
			}
			if err := tx.Create(&ban).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishMembers(uint(psid), uint(uid))
	c.JSON(http.StatusOK, gin.H{"ok": true, "banned": true})
}

// ListProblemSetBans 列出某题单已封禁（被踢出）的用户。
func (h *AdminHandler) ListProblemSetBans(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	var bans []models.ProblemSetBan
	if err := h.DB.Where("problem_set_id = ?", psid).Order("banned_at DESC").Find(&bans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(bans) == 0 {
		c.JSON(http.StatusOK, gin.H{"items": []any{}})
		return
	}
	uids := make([]uint, 0, len(bans))
	for _, b := range bans {
		uids = append(uids, b.UserID)
	}
	var users []models.User
	h.DB.Where("id IN ?", uids).Find(&users)
	byID := map[uint]models.User{}
	for _, u := range users {
		byID[u.ID] = u
	}
	type row struct {
		UserID   uint      `json:"user_id"`
		Username string    `json:"username"`
		Name     string    `json:"name"`
		BannedAt time.Time `json:"banned_at"`
	}
	out := make([]row, 0, len(bans))
	for _, b := range bans {
		u := byID[b.UserID]
		out = append(out, row{UserID: b.UserID, Username: u.Username, Name: u.Name, BannedAt: b.BannedAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

// UnbanProblemSetMember 管理员解除某用户的封禁，解除后该用户可再次加入。
// 不会自动重新加入——解除后需要用户自行点击"加入"。
func (h *AdminHandler) UnbanProblemSetMember(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	uid, _ := strconv.Atoi(c.Param("uid"))
	if err := h.DB.Where("problem_set_id = ? AND user_id = ?", psid, uid).Delete(&models.ProblemSetBan{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishMembers(uint(psid), uint(uid))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ToggleProblemSetVisibility 翻转题单 visible 开关。关闭后非 admin 看不到列表、
// 访问详情和 Join 均 404。不影响已加入成员的历史数据。
func (h *AdminHandler) ToggleProblemSetVisibility(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ps models.ProblemSet
	if err := h.DB.First(&ps, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	next := !ps.Visible
	if err := h.DB.Model(&models.ProblemSet{}).Where("id = ?", id).Update("visible", next).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true, "visible": next})
}

// CopyProblemSet duplicates a problem set (title appended with "（副本）") and
// all of its item links. Password / allowed_langs / start/end times are carried
// over; the admin can edit afterwards to diverge.
func (h *AdminHandler) CopyProblemSet(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var src models.ProblemSet
	if err := h.DB.First(&src, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	dup := models.ProblemSet{
		Title:            src.Title + "（副本）",
		AllowedLangsJSON: src.AllowedLangsJSON,
		Password:         src.Password,
		StartTime:        src.StartTime,
		EndTime:          src.EndTime,
		CreatedBy:        middleware.CurrentUserID(c),
	}
	if err := h.DB.Create(&dup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var items []models.ProblemSetItem
	h.DB.Where("problem_set_id = ?", id).Order("order_index ASC, id ASC").Find(&items)
	if len(items) > 0 {
		cloned := make([]models.ProblemSetItem, len(items))
		for i, it := range items {
			cloned[i] = models.ProblemSetItem{
				ProblemSetID: dup.ID,
				ProblemID:    it.ProblemID,
				OrderIndex:   it.OrderIndex,
			}
		}
		if err := h.DB.Create(&cloned).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	h.publishPS(dup.ID)
	c.JSON(http.StatusOK, gin.H{"id": dup.ID, "title": dup.Title})
}

type setProblemsReq struct {
	ProblemIDs []uint `json:"problem_ids"`
}

func (h *AdminHandler) SetProblemSetItems(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req setProblemsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Where("problem_set_id = ?", id).Delete(&models.ProblemSetItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.ProblemSetItem, 0, len(req.ProblemIDs))
	for i, pid := range req.ProblemIDs {
		items = append(items, models.ProblemSetItem{
			ProblemSetID: uint(id), ProblemID: pid, OrderIndex: i,
		})
	}
	if len(items) > 0 {
		if err := h.DB.Create(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	h.publishPS(uint(id))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- Home page ----------

type homeReq struct {
	Content string `json:"content"`
}

// UpdateHome 覆盖写入首页单例。Get 路径是公开的 /api/home，这里挂在 admin
// group 下所以只有管理员能改。写入成功后广播 home:changed，让正打开 / 的学
// 生/游客立即看到新内容。
func (h *AdminHandler) UpdateHome(c *gin.Context) {
	var req homeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 单例行：FirstOrCreate 取 id=1，然后覆盖 content。
	var hp models.HomePage
	h.DB.FirstOrCreate(&hp, models.HomePage{ID: 1})
	if err := h.DB.Model(&models.HomePage{}).Where("id = ?", 1).Updates(map[string]any{
		"content":    req.Content,
		"updated_at": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.Broker != nil {
		h.Broker.Publish(events.Event{Type: "home:changed", Data: nil})
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
