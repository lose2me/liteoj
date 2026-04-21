package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

type RankingHandler struct {
	DB *gorm.DB
}

// rankRow holds both global-ranking and per-problemset-ranking fields; the
// irrelevant ones carry `omitempty` so each mode produces a clean payload.
type rankRow struct {
	UserID        uint              `json:"user_id"`
	Username      string            `json:"username"`
	Name          string            `json:"name"`
	ACCount       int               `json:"ac_count"`
	TotalAttempts int               `json:"total_attempts,omitempty"`
	// Global only:
	ACRate       float64    `json:"ac_rate,omitempty"`
	AK           int        `json:"ak,omitempty"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	// Problemset only:
	PenaltyMin int               `json:"penalty_min,omitempty"`
	Results    map[uint]string   `json:"results,omitempty"` // problem_id → best verdict in-set
}

// Global ranks all users across every problem. Sort: AC count desc → AK desc
// → AC rate desc. AK counts problem sets where the user AC'd every item via
// in-set submissions (Submission.ProblemSetID = that set).
func (h *RankingHandler) Global(c *gin.Context) {
	scope := c.DefaultQuery("scope", "all")
	h.serveGlobal(c, scope)
}

// Problemset ranks the members of a single problem set. Sort: AC count desc →
// penalty asc. Only submissions tagged with this problemset_id count — a
// student who AC'd the problem on its standalone page does NOT get credit
// here. Each row also carries per-problem verdicts so the frontend can draw
// an ICPC-style A/B/C… grid.
func (h *RankingHandler) Problemset(c *gin.Context) {
	psid, _ := strconv.Atoi(c.Param("id"))
	h.serveProblemset(c, uint(psid), c.DefaultQuery("scope", "all"))
}

// --- Global ---------------------------------------------------------------

func (h *RankingHandler) serveGlobal(c *gin.Context, scope string) {
	since := scopeStart(scope)

	type rawRow struct {
		UserID       uint
		ProblemID    uint
		ProblemSetID uint // COALESCE(problem_set_id, 0) — 0 表示独立页上下文
		Verdict      string
		TimeUsedMS   int
		CreatedAt    time.Time
	}
	q := h.DB.Table("submissions").
		Select("user_id, problem_id, COALESCE(problem_set_id, 0) AS problem_set_id, verdict, time_used_ms, created_at").
		Where("verdict IN ?", []string{
			models.VerdictAC, models.VerdictWA, models.VerdictTLE,
			models.VerdictMLE, models.VerdictOLE, models.VerdictRE,
			models.VerdictCE, models.VerdictPE,
		}).
		Order("user_id ASC, problem_id ASC, problem_set_id ASC, created_at ASC")
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	var raw []rawRow
	q.Scan(&raw)

	type agg struct {
		FirstAC    bool
		Attempts   int
		ACAttempts int // attempts up to and including first AC
	}
	// 以 (user, problem, context) 为键——同一题在不同题单 / 独立页各自计数。
	type tripleKey struct{ u, p, ps uint }
	perPair := map[tripleKey]*agg{}
	for _, r := range raw {
		k := tripleKey{r.UserID, r.ProblemID, r.ProblemSetID}
		a, ok := perPair[k]
		if !ok {
			a = &agg{}
			perPair[k] = a
		}
		a.Attempts++
		if !a.FirstAC {
			a.ACAttempts++
			if r.Verdict == models.VerdictAC {
				a.FirstAC = true
			}
		}
	}

	perUser := map[uint]*rankRow{}
	for k, a := range perPair {
		u, ok := perUser[k.u]
		if !ok {
			u = &rankRow{UserID: k.u, Results: nil}
			perUser[k.u] = u
		}
		u.TotalAttempts += a.Attempts
		if a.FirstAC {
			u.ACCount++
		}
	}

	// AC rate: AC submissions / total submissions (ACCount is distinct-problem
	// AC'd, so we need the raw count).
	acSubs := map[uint]int{}
	totalSubs := map[uint]int{}
	for _, r := range raw {
		totalSubs[r.UserID]++
		if r.Verdict == models.VerdictAC {
			acSubs[r.UserID]++
		}
	}
	for uid, row := range perUser {
		if totalSubs[uid] > 0 {
			row.ACRate = float64(acSubs[uid]) / float64(totalSubs[uid])
		}
	}

	// AK: for each user, count problem sets where every item has ≥1 AC
	// submission filed with that set's problemset_id. The JOIN on
	// problem_set_items prevents a rogue problemset_id on a submission from
	// counting toward a problem that isn't in the set.
	akPerUser := computeAKPerUser(h.DB, since)
	for uid, n := range akPerUser {
		if row, ok := perUser[uid]; ok {
			row.AK = n
		}
	}

	// LastActiveAt：独立聚合 MAX(created_at) per user（在当前 scope 下）。
	// 用 string 接收 + 手动解析——SQLite 聚合函数（MAX/MIN 等）会丢掉
	// 列的 declared type，modernc.org/sqlite 会把结果以原始 string 形式
	// 返回，无法直接 scan 到 time.Time。
	type lastSubRow struct {
		UserID uint
		Latest string
	}
	var lastSubs []lastSubRow
	qLast := h.DB.Model(&models.Submission{}).
		Select("user_id, MAX(created_at) AS latest").
		Group("user_id")
	if !since.IsZero() {
		qLast = qLast.Where("created_at >= ?", since)
	}
	qLast.Scan(&lastSubs)
	for _, ls := range lastSubs {
		t, ok := parseDBTime(ls.Latest)
		if !ok {
			continue
		}
		if row, rok := perUser[ls.UserID]; rok {
			tt := t
			row.LastActiveAt = &tt
		}
	}

	attachUserNames(h.DB, perUser)

	rows := make([]rankRow, 0, len(perUser))
	for _, r := range perUser {
		rows = append(rows, *r)
	}
	sortSliceStable(rows, func(a, b rankRow) bool {
		if a.ACCount != b.ACCount {
			return a.ACCount > b.ACCount
		}
		if a.AK != b.AK {
			return a.AK > b.AK
		}
		return a.ACRate > b.ACRate
	})
	c.JSON(http.StatusOK, gin.H{"items": rows, "scope": scope})
}

// computeAKPerUser returns user_id → AK count. An AK is when a user has
// AC'd every item of a problem set via submissions whose problemset_id equals
// that set. `since` optionally bounds AC time.
func computeAKPerUser(db *gorm.DB, since time.Time) map[uint]int {
	type total struct {
		ProblemSetID uint
		N            int
	}
	var totals []total
	db.Table("problem_set_items").
		Select("problem_set_id, COUNT(*) AS n").
		Group("problem_set_id").Scan(&totals)
	totalByPS := map[uint]int{}
	for _, t := range totals {
		totalByPS[t.ProblemSetID] = t.N
	}

	type acRow struct {
		ProblemSetID uint
		UserID       uint
		N            int
	}
	var acRows []acRow
	qry := db.Table("submissions AS s").
		Joins("JOIN problem_set_items psi ON psi.problem_set_id = s.problem_set_id AND psi.problem_id = s.problem_id").
		Where("s.verdict = ? AND s.problem_set_id IS NOT NULL", models.VerdictAC).
		Select("s.problem_set_id, s.user_id, COUNT(DISTINCT s.problem_id) AS n").
		Group("s.problem_set_id, s.user_id")
	if !since.IsZero() {
		qry = qry.Where("s.created_at >= ?", since)
	}
	qry.Scan(&acRows)

	ak := map[uint]int{}
	for _, r := range acRows {
		if r.N > 0 && r.N == totalByPS[r.ProblemSetID] {
			ak[r.UserID]++
		}
	}
	return ak
}

// --- Problemset -----------------------------------------------------------

func (h *RankingHandler) serveProblemset(c *gin.Context, psid uint, scope string) {
	since := scopeStart(scope)

	// Item order drives the A/B/C… labels the frontend renders.
	var items []models.ProblemSetItem
	h.DB.Where("problem_set_id = ?", psid).Order("order_index ASC, id ASC").Find(&items)
	if len(items) == 0 {
		c.JSON(http.StatusOK, gin.H{"items": []rankRow{}, "problems": []any{}})
		return
	}
	problemIDs := make([]uint, 0, len(items))
	for _, it := range items {
		problemIDs = append(problemIDs, it.ProblemID)
	}

	var problems []models.Problem
	h.DB.Where("id IN ?", problemIDs).Find(&problems)
	byID := map[uint]models.Problem{}
	for _, p := range problems {
		byID[p.ID] = p
	}
	type probSummary struct {
		ID    uint   `json:"id"`
		Code  string `json:"code"`
		Title string `json:"title"`
	}
	orderedProblems := make([]probSummary, 0, len(items))
	for i, it := range items {
		p, ok := byID[it.ProblemID]
		if !ok {
			continue
		}
		orderedProblems = append(orderedProblems, probSummary{
			ID: p.ID, Code: problemCode(i), Title: p.Title,
		})
	}

	// Pull every in-set submission for these problems, in time order so we can
	// reconstruct "attempts before first AC" and the freshest non-AC verdict.
	// ai_explanation / ai_rejected are pulled too because an AI Analyze used
	// in-set also contributes +20 min to penalty (mirroring a wrong attempt).
	type rawRow struct {
		UserID        uint
		ProblemID     uint
		Verdict       string
		TimeUsedMS    int
		CreatedAt     time.Time
		AIExplanation string
		AIRejected    bool
	}
	q := h.DB.Table("submissions").
		Select("user_id, problem_id, verdict, time_used_ms, created_at, ai_explanation, ai_rejected").
		Where("problem_set_id = ? AND problem_id IN ?", psid, problemIDs).
		Where("verdict IN ?", []string{
			models.VerdictAC, models.VerdictWA, models.VerdictTLE,
			models.VerdictMLE, models.VerdictOLE, models.VerdictRE,
			models.VerdictCE, models.VerdictPE,
		}).
		Order("user_id ASC, problem_id ASC, created_at ASC")
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	var raw []rawRow
	q.Scan(&raw)

	type pairAgg struct {
		FirstAC        bool
		WABeforeAC     int    // counts WA/TLE/MLE/RE/CE attempts strictly before first AC
		AIUsedBeforeAC int    // counts successful AI analyses on non-AC subs before first AC
		LatestNonAC    string // most recent non-AC verdict (used if no AC yet)
		LatestNonACAt  time.Time
	}
	pairs := map[struct{ u, p uint }]*pairAgg{}
	for _, r := range raw {
		k := struct{ u, p uint }{r.UserID, r.ProblemID}
		a, ok := pairs[k]
		if !ok {
			a = &pairAgg{}
			pairs[k] = a
		}
		if r.Verdict == models.VerdictAC {
			if !a.FirstAC {
				a.FirstAC = true
			}
			continue
		}
		if !a.FirstAC {
			a.WABeforeAC++
			// AI usage penalty only counts a *successful* analysis: rejected
			// runs produced no explanation for the student so they shouldn't
			// be charged. Optimize / tag / gen_* tasks don't write
			// ai_explanation on submissions at all, so this check is safe.
			if r.AIExplanation != "" && !r.AIRejected {
				a.AIUsedBeforeAC++
			}
		}
		if r.CreatedAt.After(a.LatestNonACAt) {
			a.LatestNonAC = r.Verdict
			a.LatestNonACAt = r.CreatedAt
		}
	}

	perUser := map[uint]*rankRow{}
	// 把题单的全体成员先都塞进 perUser——即使 0 提交也要在排行榜里露面。
	// admin 视角可能看"全局" scope 里会出现非成员的提交（特殊情况：学生退
	// 出题单后历史提交仍在），所以聚合阶段遇到新 user_id 仍然会 append，成
	// 员表只负责兜底"未提交过的人"。
	var memberIDs []uint
	h.DB.Model(&models.ProblemSetMember{}).
		Where("problem_set_id = ?", psid).
		Pluck("user_id", &memberIDs)
	for _, uid := range memberIDs {
		perUser[uid] = &rankRow{UserID: uid, Results: map[uint]string{}}
	}
	for k, a := range pairs {
		u, ok := perUser[k.u]
		if !ok {
			u = &rankRow{UserID: k.u, Results: map[uint]string{}}
			perUser[k.u] = u
		}
		if a.FirstAC {
			u.ACCount++
			// ICPC-style penalty: 20 minutes per wrong attempt preceding the
			// first AC, per problem. Additionally, each successful AI analysis
			// used on an in-set non-AC submission before that first AC adds
			// another 20 min — discourages offloading to AI. No wall-clock
			// time component — keeps the math local to attempts so
			// ps.StartTime being optional is fine.
			u.PenaltyMin += (a.WABeforeAC + a.AIUsedBeforeAC) * 20
			u.Results[k.p] = models.VerdictAC
		} else if a.LatestNonAC != "" {
			u.Results[k.p] = a.LatestNonAC
		}
	}

	attachUserNames(h.DB, perUser)

	rows := make([]rankRow, 0, len(perUser))
	for _, r := range perUser {
		rows = append(rows, *r)
	}
	sortSliceStable(rows, func(a, b rankRow) bool {
		if a.ACCount != b.ACCount {
			return a.ACCount > b.ACCount
		}
		return a.PenaltyMin < b.PenaltyMin
	})
	c.JSON(http.StatusOK, gin.H{
		"items":    rows,
		"scope":    scope,
		"problems": orderedProblems,
	})
}

// attachUserNames fills in Username/Name on every rankRow in-place.
func attachUserNames(db *gorm.DB, perUser map[uint]*rankRow) {
	if len(perUser) == 0 {
		return
	}
	ids := make([]uint, 0, len(perUser))
	for id := range perUser {
		ids = append(ids, id)
	}
	var users []models.User
	db.Where("id IN ?", ids).Find(&users)
	for _, u := range users {
		if r, ok := perUser[u.ID]; ok {
			r.Username = u.Username
			r.Name = u.Name
		}
	}
}

// problemCode maps a 0-based index to an Excel-style column label: 0→A,
// 1→B, …, 25→Z, 26→AA, 27→AB, …. Used for per-problem-set labels that
// the admin configured via order.
func problemCode(idx int) string {
	if idx < 0 {
		return ""
	}
	var out []byte
	n := idx
	for {
		out = append([]byte{byte('A' + n%26)}, out...)
		n = n/26 - 1
		if n < 0 {
			break
		}
	}
	return string(out)
}

func scopeStart(scope string) time.Time {
	now := time.Now()
	switch scope {
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "year":
		return now.AddDate(-1, 0, 0)
	}
	return time.Time{}
}

// parseDBTime 尝试以常见格式解析 SQLite 回传的时间字符串。
// modernc.org/sqlite 对聚合结果（MAX/MIN）不保留 declared type，会把
// 时间值以存储原样的字符串形式返回，无法自动映射到 time.Time。
func parseDBTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	// GORM 写入 SQLite 常见的几种格式，优先更宽松 / 带时区的。
	layouts := []string{
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	// 最后兜底用本地时区解 "YYYY-MM-DD HH:MM:SS"
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); err == nil {
		return t, true
	}
	return time.Time{}, false
}
