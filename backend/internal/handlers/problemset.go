package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type ProblemSetHandler struct {
	DB     *gorm.DB
	C      *config.Config
	Broker *events.Broker
}

// decodeAllowedLangs parses the persisted JSON into the transient slice. Empty
// string means "no restriction"; unreadable JSON is treated the same (safe
// degrade rather than 500 on a corrupted row).
func decodeAllowedLangs(s string) []string {
	if s == "" {
		return nil
	}
	var v []string
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil
	}
	return v
}

// encodeAllowedLangs stores the slice as JSON; nil or empty slice persists as
// "" so the "no restriction" case round-trips cleanly.
func encodeAllowedLangs(v []string) string {
	if len(v) == 0 {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// psListRow wraps ProblemSet with UI-facing status that the client cares about
// but we never want to persist: whether a password is set (derived, not the
// actual secret), how many problems the set contains, how many of them the
// current user has AC'd, and who the current leader is.
type psListRow struct {
	models.ProblemSet
	HasPassword bool   `json:"has_password"`
	IsMember    bool   `json:"is_member"`
	IsBanned    bool   `json:"is_banned"`
	ItemCount   int    `json:"item_count"`
	MyACCount   int    `json:"my_ac_count"`
	TopACCount  int    `json:"top_ac_count"`
	TopACName   string `json:"top_ac_name"`
}

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

	// Leader per problem set: the user who AC'd the most distinct problems in
	// that set via in-set submissions. Single aggregation over (set, user)
	// pairs; Go picks the max per set in memory.
	// anti-join bans：被踢的学生即使提交仍在，也不应该再顶榜首。
	type topInfo struct {
		uid uint
		ac  int
	}
	topByPS := map[uint]topInfo{}
	{
		type r struct {
			ProblemSetID uint
			UserID       uint
			AC           int
		}
		var rows []r
		h.DB.Raw(`
			SELECT psi.problem_set_id, s.user_id, COUNT(DISTINCT psi.problem_id) AS ac
			  FROM problem_set_items psi
			  JOIN submissions s
			    ON s.problem_id = psi.problem_id
			   AND s.problem_set_id = psi.problem_set_id
			   AND s.verdict = 'AC'
			   AND NOT EXISTS (
			     SELECT 1 FROM problem_set_bans b
			      WHERE b.problem_set_id = s.problem_set_id
			        AND b.user_id       = s.user_id
			   )
			 GROUP BY psi.problem_set_id, s.user_id`).Scan(&rows)
		for _, row := range rows {
			cur, ok := topByPS[row.ProblemSetID]
			if !ok || row.AC > cur.ac {
				topByPS[row.ProblemSetID] = topInfo{uid: row.UserID, ac: row.AC}
			}
		}
	}

	// Resolve leader names in a single IN query.
	nameByUser := map[uint]string{}
	if len(topByPS) > 0 {
		ids := make([]uint, 0, len(topByPS))
		for _, t := range topByPS {
			ids = append(ids, t.uid)
		}
		var us []models.User
		h.DB.Where("id IN ?", ids).Find(&us)
		for _, u := range us {
			if u.Name != "" {
				nameByUser[u.ID] = u.Name
			} else {
				nameByUser[u.ID] = u.Username
			}
		}
	}

	out := make([]psListRow, len(sets))
	for i, s := range sets {
		hasPwd := s.Password != ""
		s.Password = "" // never leak
		s.AllowedLangs = decodeAllowedLangs(s.AllowedLangsJSON)
		top := topByPS[s.ID]
		out[i] = psListRow{
			ProblemSet:  s,
			HasPassword: hasPwd,
			IsMember:    memberByPS[s.ID],
			IsBanned:    bannedByPS[s.ID],
			ItemCount:   countByPS[s.ID],
			MyACCount:   acByPS[s.ID],
			TopACCount:  top.ac,
			TopACName:   nameByUser[top.uid],
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *ProblemSetHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ps models.ProblemSet
	if err := h.DB.First(&ps, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}

	uid := middleware.CurrentUserID(c)
	isAdmin := middleware.CurrentRole(c) == models.RoleAdmin
	hasPwd := ps.Password != ""

	// 可见性：非 admin 访问已关闭的题单 → 404。与 Problem.visible 同语义。
	if !isAdmin && !ps.Visible {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}

	// 成员判定：admin 视为已加入；否则查询 problem_set_members。
	isMember := isAdmin
	if !isAdmin && uid > 0 {
		var n int64
		h.DB.Model(&models.ProblemSetMember{}).
			Where("problem_set_id = ? AND user_id = ?", ps.ID, uid).
			Count(&n)
		isMember = n > 0
	}

	// 封禁判定：非 admin 且 uid > 0 时查询 problem_set_bans。
	// admin 不受封禁影响。
	isBanned := false
	if !isAdmin && uid > 0 {
		var n int64
		h.DB.Model(&models.ProblemSetBan{}).
			Where("problem_set_id = ? AND user_id = ?", ps.ID, uid).
			Count(&n)
		isBanned = n > 0
	}

	ps.Password = "" // never leak
	ps.AllowedLangs = decodeAllowedLangs(ps.AllowedLangsJSON)

	// 非成员且非 admin：返回元信息 + locked，不暴露题目列表，但仍告知题目总数
	// 和榜首——这些在 List 接口已对全员可见，没必要再隐藏。前端据此把非成员
	// 视图渲染得跟成员页一致，仅少掉三个 tab。
	if !isMember {
		reason := "not_member"
		if isBanned {
			reason = "banned"
		}
		itemCount, topName, topAC := h.problemSetHeadline(ps.ID)
		c.JSON(http.StatusOK, gin.H{
			"problemset":   ps,
			"has_password": hasPwd,
			"is_member":    false,
			"is_banned":    isBanned,
			"locked":       true,
			"lock_reason":  reason,
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

	// Leader for this set: user with most distinct in-set AC. The
	// problem_set_id filter mirrors `mystatus` above — same independence
	// rule for both the student's own progress and the leaderboard head.
	var topName string
	topAC := 0
	if len(ids) > 0 {
		type r struct {
			UserID uint
			AC     int
		}
		var rows []r
		h.DB.Raw(`
			SELECT user_id, COUNT(DISTINCT problem_id) AS ac
			  FROM submissions s
			 WHERE verdict = 'AC' AND problem_set_id = ? AND problem_id IN ?
			   AND NOT EXISTS (
			     SELECT 1 FROM problem_set_bans b
			      WHERE b.problem_set_id = s.problem_set_id
			        AND b.user_id       = s.user_id
			   )
			 GROUP BY user_id
			 ORDER BY ac DESC LIMIT 1`, ps.ID, ids).Scan(&rows)
		if len(rows) > 0 && rows[0].AC > 0 {
			topAC = rows[0].AC
			var u models.User
			if err := h.DB.First(&u, rows[0].UserID).Error; err == nil {
				if u.Name != "" {
					topName = u.Name
				} else {
					topName = u.Username
				}
			}
		}
	}

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
	type r struct {
		UserID uint
		AC     int
	}
	var rows []r
	h.DB.Raw(`
		SELECT s.user_id, COUNT(DISTINCT s.problem_id) AS ac
		  FROM submissions s
		  JOIN problem_set_items psi
		    ON psi.problem_set_id = s.problem_set_id
		   AND psi.problem_id    = s.problem_id
		 WHERE s.verdict = 'AC' AND s.problem_set_id = ?
		   AND NOT EXISTS (
		     SELECT 1 FROM problem_set_bans b
		      WHERE b.problem_set_id = s.problem_set_id
		        AND b.user_id       = s.user_id
		   )
		 GROUP BY s.user_id
		 ORDER BY ac DESC LIMIT 1`, psID).Scan(&rows)
	if len(rows) == 0 || rows[0].AC == 0 {
		return
	}
	topAC = rows[0].AC
	var u models.User
	if err := h.DB.First(&u, rows[0].UserID).Error; err == nil {
		if u.Name != "" {
			topName = u.Name
		} else {
			topName = u.Username
		}
	}
	return
}

// Join 让当前用户加入题单。如果题单设有密码，必须提供正确密码。
// 已加入时幂等。被管理员踢出后会写入 problem_set_bans，此时拒绝加入
// （admin 自身不受封禁限制）。
func (h *ProblemSetHandler) Join(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	var ps models.ProblemSet
	if err := h.DB.First(&ps, psid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	uid := middleware.CurrentUserID(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.ErrForbidden})
		return
	}
	isAdmin := middleware.CurrentRole(c) == models.RoleAdmin
	// 可见性：非 admin 访问已关闭题单 → 404（与 Detail 保持一致，避免"列表
	// 看不到却能加入"的路径）。
	if !isAdmin && !ps.Visible {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	// 封禁校验：非 admin 被踢出后永久失去加入权。
	if !isAdmin {
		var nb int64
		h.DB.Model(&models.ProblemSetBan{}).
			Where("problem_set_id = ? AND user_id = ?", psid, uid).Count(&nb)
		if nb > 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "banned"})
			return
		}
	}
	// 密码校验：仅对非管理员生效。
	if ps.Password != "" && !isAdmin {
		var body struct {
			Password string `json:"password"`
		}
		_ = c.ShouldBindJSON(&body)
		if body.Password != ps.Password {
			c.JSON(http.StatusForbidden, gin.H{"error": "password_incorrect"})
			return
		}
	}
	// 幂等加入：若已存在 membership，直接返回。
	var n int64
	h.DB.Model(&models.ProblemSetMember{}).Where("problem_set_id = ? AND user_id = ?", psid, uid).Count(&n)
	if n == 0 {
		m := models.ProblemSetMember{ProblemSetID: uint(psid), UserID: uid, JoinedAt: time.Now()}
		if err := h.DB.Create(&m).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if h.Broker != nil {
			h.Broker.Publish(events.Event{
				Type: "problemset:members:changed",
				Data: map[string]any{"id": uint(psid), "user_id": uid},
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"joined": true})
}
