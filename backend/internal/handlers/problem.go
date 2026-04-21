package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/cache"
	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/db"
	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
)

type ProblemHandler struct {
	DB    *gorm.DB
	C     *config.Config
	Cache *cache.Cache
}

const (
	statusNone      = ""
	statusAttempted = "attempted"
	statusAC        = "AC"
	statusACFaded   = "AC_FADED"
)

func (h *ProblemHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := h.DB.Model(&models.Problem{})
	if middleware.CurrentRole(c) != models.RoleAdmin {
		q = q.Where("visible = ?", true)
	}
	if kw := strings.TrimSpace(c.Query("q")); kw != "" {
		q = q.Where("title LIKE ?", "%"+kw+"%")
	}
	if diff := strings.TrimSpace(c.Query("difficulty")); diff != "" {
		q = q.Where("difficulty = ?", diff)
	}
	if tag := strings.TrimSpace(c.Query("tag_id")); tag != "" {
		// TagsJSON stores an array of uints; LIKE is enough for a lightweight match.
		q = q.Where("tags_json LIKE ?", "%"+tag+"%")
	}

	var total int64
	q.Count(&total)
	var items []models.Problem
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	statuses := h.computeStatuses(middleware.CurrentUserID(c), collectIDs(items))
	tagNameMap := h.resolveTagNames(items)

	// 批量查 "problem #<id>" 主题下正在运行的 AI 任务——admin 列表用来给行
	// 打 spinner + 禁止进入编辑页，防止一键解析还没落盘就点开看到旧字段。
	aiPending := map[uint]bool{}
	if len(items) > 0 {
		subjects := make([]string, 0, len(items))
		for _, p := range items {
			subjects = append(subjects, fmt.Sprintf("problem #%d", p.ID))
		}
		type tr struct{ Subject string }
		var trs []tr
		h.DB.Table("ai_tasks").
			Select("subject").
			Where("status = ? AND subject IN ?", models.AITaskStatusRunning, subjects).
			Scan(&trs)
		for _, row := range trs {
			var id uint
			if _, err := fmt.Sscanf(row.Subject, "problem #%d", &id); err == nil {
				aiPending[id] = true
			}
		}
	}

	// 聚合对本页题目的限制：
	// - 学生：仅统计 "visible + 当前用户已加入 + 至少一个 disable_* 为 true" 的题单；
	// - admin：放宽为"任意题单（不论 visible / 是否加入）+ 至少一个 disable_*"，
	//   用于后台题目管理页提示"这题在某题单里处于限制状态"。
	// 结果按 problem_id 打 3 个布尔位：restricted_idea / solution / ai。
	type restriction struct{ idea, solution, ai bool }
	restrictions := map[uint]restriction{}
	if len(items) > 0 {
		pids := collectIDs(items)
		type row struct {
			ProblemID uint
			RIdea     bool
			RSol      bool
			RAI       bool
		}
		var rows []row
		if middleware.CurrentRole(c) == models.RoleAdmin {
			h.DB.Raw(`
				SELECT psi.problem_id,
				       MAX(ps.disable_idea)     AS r_idea,
				       MAX(ps.disable_solution) AS r_sol,
				       MAX(ps.disable_ai)       AS r_ai
				  FROM problem_set_items psi
				  JOIN problem_sets ps ON ps.id = psi.problem_set_id
				 WHERE psi.problem_id IN ?
				 GROUP BY psi.problem_id`, pids).Scan(&rows)
		} else if uid := middleware.CurrentUserID(c); uid > 0 {
			h.DB.Raw(`
				SELECT psi.problem_id,
				       MAX(ps.disable_idea)     AS r_idea,
				       MAX(ps.disable_solution) AS r_sol,
				       MAX(ps.disable_ai)       AS r_ai
				  FROM problem_set_items psi
				  JOIN problem_sets ps
				    ON ps.id = psi.problem_set_id AND ps.visible = 1
				  JOIN problem_set_members m
				    ON m.problem_set_id = ps.id AND m.user_id = ?
				 WHERE psi.problem_id IN ?
				 GROUP BY psi.problem_id`, uid, pids).Scan(&rows)
		}
		for _, r := range rows {
			restrictions[r.ProblemID] = restriction{idea: r.RIdea, solution: r.RSol, ai: r.RAI}
		}
	}

	type problemRow struct {
		models.Problem
		MyStatus           string   `json:"my_status"`
		TagNames           []string `json:"tag_names"`
		AIPending          bool     `json:"ai_pending"`
		RestrictedIdea     bool     `json:"restricted_idea"`
		RestrictedSolution bool     `json:"restricted_solution"`
		RestrictedAI       bool     `json:"restricted_ai"`
	}
	out := make([]problemRow, len(items))
	for i, p := range items {
		rst := restrictions[p.ID]
		out[i] = problemRow{
			Problem: p, MyStatus: statuses[p.ID], TagNames: tagNameMap[p.ID],
			AIPending:          aiPending[p.ID],
			RestrictedIdea:     rst.idea,
			RestrictedSolution: rst.solution,
			RestrictedAI:       rst.ai,
		}
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": out, "page": page, "page_size": pageSize})
}

func (h *ProblemHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.Problem
	if err := h.DB.First(&p, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	if !p.Visible && middleware.CurrentRole(c) != models.RoleAdmin {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	var tcs []models.Testcase
	h.DB.Where("problem_id = ?", p.ID).Order("order_index ASC, id ASC").Find(&tcs)

	tagIDs := parseTagsJSON(p.TagsJSON)
	var tags []models.Tag
	if len(tagIDs) > 0 {
		h.DB.Where("id IN ?", tagIDs).Find(&tags)
	}

	// If opened in the context of a problem set, narrow the language choices
	// to the intersection of the set's allow-list and the globally configured
	// judge langs. Empty allow-list means "no restriction" (fall through to
	// cfg.JudgeLangs). Direct access (no problemset_id) is never restricted —
	// enforcement happens on submit, so this only shapes the UI.
	// 题单上下文还承担 Disable{Idea,Solution,AI} 三个开关：通过置空 markdown
	// 字段让前端已有的 v-if 分支自然隐藏相应 tab；同时把布尔值透出，前端据此
	// 在标题行渲染"禁用思路/题解/AI"标签告知学生。
	langs := h.C.JudgeLangs
	disableIdea := false
	disableSolution := false
	disableAI := false
	if psid, _ := strconv.Atoi(c.Query("problemset_id")); psid > 0 {
		var ps models.ProblemSet
		if err := h.DB.First(&ps, psid).Error; err == nil {
			allowed := decodeAllowedLangs(ps.AllowedLangsJSON)
			if len(allowed) > 0 {
				allowSet := make(map[string]bool, len(allowed))
				for _, l := range allowed {
					allowSet[l] = true
				}
				narrowed := make([]string, 0, len(allowed))
				for _, l := range h.C.JudgeLangs {
					if allowSet[l] {
						narrowed = append(narrowed, l)
					}
				}
				langs = narrowed
			}
			if ps.DisableIdea {
				p.SolutionIdeaMD = ""
				disableIdea = true
			}
			if ps.DisableSolution {
				p.SolutionMD = ""
				disableSolution = true
			}
			disableAI = ps.DisableAI
		}
	}

	// 上次 AI 痕迹：只在**当前上下文**里回忆。
	//   - 独立页（无 problemset_id）：只取 problem_set_id IS NULL 的提交——
	//     题单里做的 AI 不会泄漏到独立页。
	//   - 题单页（有 problemset_id）：只取 problem_set_id = 该题单 的提交；
	//     题单禁用 AI（disable_ai）时整段跳过。
	// 返回的 type 字段（'analyze' / 'optimize'）让前端区分标签"上一次解析"
	// 还是"上一次优化"；AC 的 ai_explanation 是 OptimizeAC 产生的，否则是
	// AnalyzeWrongAnswer 产生的。
	var myLatestAI any
	if uid := middleware.CurrentUserID(c); uid > 0 && !disableAI {
		type row struct {
			ID            uint   `json:"submission_id"`
			Verdict       string `json:"verdict"`
			AIExplanation string `json:"explanation"`
			Type          string `json:"type"`
		}
		var r row
		q := h.DB.Table("submissions").
			Select("id, verdict, ai_explanation").
			Where("user_id = ? AND problem_id = ? AND ai_rejected = ? AND ai_explanation <> ?",
				uid, p.ID, false, "")
		if psidParam := c.Query("problemset_id"); psidParam != "" {
			q = q.Where("problem_set_id = ?", psidParam)
		} else {
			q = q.Where("problem_set_id IS NULL")
		}
		if err := q.Order("id DESC").Limit(1).Scan(&r).Error; err == nil && r.ID > 0 {
			if r.Verdict == models.VerdictAC {
				r.Type = "optimize"
			} else {
				r.Type = "analyze"
			}
			myLatestAI = r
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"problem":          p,
		"testcases":        tcs,
		"tag_ids":          tagIDs,
		"tags":             tags,
		"languages":        langs,
		"disable_idea":     disableIdea,
		"disable_solution": disableSolution,
		"disable_ai":       disableAI,
		"my_latest_ai":     myLatestAI,
	})
}

// computeStatuses returns id -> my_status for the given problem IDs.
func (h *ProblemHandler) computeStatuses(uid uint, ids []uint) map[uint]string {
	out := map[uint]string{}
	if uid == 0 || len(ids) == 0 {
		return out
	}
	// db.NullTime: aggregate aliases (MIN/MAX of a datetime) lose GORM's model
	// type hint, so the pure-Go sqlite driver hands back plain strings. NullTime
	// parses them. See backend/internal/db/nulltime.go.
	type row struct {
		ProblemID uint
		FirstAC   db.NullTime
		LastAny   db.NullTime
		LastWrong db.NullTime
	}
	var rows []row
	h.DB.Raw(`
		SELECT problem_id,
		       MIN(CASE WHEN verdict = 'AC' THEN created_at END) AS first_ac,
		       MAX(created_at) AS last_any,
		       MAX(CASE WHEN verdict IN ('WA','TLE','MLE','OLE','RE','CE','PE') THEN created_at END) AS last_wrong
		FROM submissions
		WHERE user_id = ? AND problem_id IN ?
		GROUP BY problem_id`, uid, ids).Scan(&rows)
	for _, r := range rows {
		if !r.LastAny.Valid {
			continue
		}
		if !r.FirstAC.Valid {
			out[r.ProblemID] = statusAttempted
			continue
		}
		if r.LastWrong.Valid && r.LastWrong.Time.After(r.FirstAC.Time) {
			out[r.ProblemID] = statusACFaded
		} else {
			out[r.ProblemID] = statusAC
		}
	}
	return out
}

func (h *ProblemHandler) resolveTagNames(items []models.Problem) map[uint][]string {
	out := map[uint][]string{}
	idSet := map[uint]bool{}
	perProblem := map[uint][]uint{}
	for _, p := range items {
		ids := parseTagsJSON(p.TagsJSON)
		perProblem[p.ID] = ids
		for _, id := range ids {
			idSet[id] = true
		}
	}
	if len(idSet) == 0 {
		return out
	}
	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var tags []models.Tag
	h.DB.Where("id IN ?", ids).Find(&tags)
	nameByID := map[uint]string{}
	for _, t := range tags {
		nameByID[t.ID] = t.Name
	}
	for pid, tagIDs := range perProblem {
		names := make([]string, 0, len(tagIDs))
		for _, tid := range tagIDs {
			if n, ok := nameByID[tid]; ok {
				names = append(names, n)
			}
		}
		if len(names) > 0 {
			out[pid] = names
		}
	}
	return out
}

func collectIDs(items []models.Problem) []uint {
	out := make([]uint, 0, len(items))
	for _, p := range items {
		out = append(out, p.ID)
	}
	return out
}

func parseTagsJSON(s string) []uint {
	if s == "" {
		return nil
	}
	var ids []uint
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		return nil
	}
	return ids
}

func serializeTagsJSON(ids []uint) string {
	if len(ids) == 0 {
		return ""
	}
	b, _ := json.Marshal(ids)
	return string(b)
}
