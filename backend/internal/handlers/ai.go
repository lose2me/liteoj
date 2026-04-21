package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/i18n"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/models"
	"github.com/liteoj/liteoj/backend/internal/services/ai"
)

type AIHandler struct {
	DB     *gorm.DB
	Queue  *ai.Queue
	Runner *ai.Runner
}

// rawBody is the common payload for every problem-authoring AI endpoint.
// The admin pastes the original problem text into ProblemEdit's 详细 field;
// the frontend forwards that string here so all AI flows share one source
// of truth (the "详细" blob) regardless of what's in the structured fields.
type rawBody struct {
	Raw string `json:"raw"`
}

// Analyze kicks off a non-AC-submission analysis. Cached explanations short-
// circuit synchronously so the client doesn't wait on a new model call for a
// result we already have. Fresh runs are enqueued: the handler returns a
// task_id that the client pairs with the `ai:task:done` SSE event.
func (h *AIHandler) Analyze(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var sub models.Submission
	if err := h.DB.First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	uid := middleware.CurrentUserID(c)
	if middleware.CurrentRole(c) != models.RoleAdmin && sub.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
		return
	}
	if sub.Verdict == models.VerdictAC {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrAIAcNoAnalyze})
		return
	}
	// 题单维度的 AI 禁用：如果该提交挂在某个禁用 AI 的题单里，直接拒绝。
	if sub.ProblemSetID != nil && *sub.ProblemSetID > 0 {
		var ps models.ProblemSet
		if err := h.DB.Select("disable_ai").First(&ps, *sub.ProblemSetID).Error; err == nil && ps.DisableAI {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
			return
		}
	}
	if sub.AIExplanation != "" {
		c.JSON(http.StatusOK, gin.H{"explanation": sub.AIExplanation, "cached": true})
		return
	}
	// 此前被 AI 判为"未认真作答"而拒绝分析的提交，允许学生再次触发——
	// 但先把 rejected 状态清掉，这样新任务写回 explanation/rejected 时
	// 不会因为旧 reason 残留让前端误判。
	if sub.AIRejected {
		h.DB.Model(&sub).Updates(map[string]any{
			"ai_rejected":      false,
			"ai_reject_reason": "",
		})
	}
	// Pre-flight: the target problem must exist before we accept the job.
	var prob models.Problem
	if err := h.DB.First(&prob, sub.ProblemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrProblemNotFound})
		return
	}
	_ = prob // existence check only; Runner re-fetches inside the worker
	// 管理员触发的 analyze：跳过 prompt_wrong_answer 里的"乱写判定"。模型仍
	// 可能输出 ok=false，但 runner 看到 ForceAnalyze 后不走拒绝流程、照样
	// 把 explanation（或原文）写回 ai_explanation。学生点自己的提交仍然按
	// 正常判定。
	isAdmin := middleware.CurrentRole(c) == models.RoleAdmin
	taskID := h.Queue.Start(models.AITaskKindAnalyze, uid, middleware.CurrentUsername(c),
		fmt.Sprintf("submission #%d", sub.ID))
	if err := h.Runner.Enqueue(ai.Job{
		TaskID: taskID, Kind: models.AITaskKindAnalyze,
		UserID: uid, SubmissionID: sub.ID,
		ForceAnalyze: isAdmin,
	}); err != nil {
		h.Queue.End(taskID, err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID, "status": models.AITaskStatusRunning})
}

// startAuthoringTask is the shared accept-and-enqueue path for every admin
// problem-authoring flow (tag / gen_title / gen_desc / gen_idea / gen_explain
// / gen_all). It validates the body, records a running AITask row, and hands
// the job to the Runner. The HTTP response is always 202 + task_id — results
// are delivered later via `ai:task:done` and fetched via GET /admin/ai/tasks/:id.
func (h *AIHandler) startAuthoringTask(c *gin.Context, kind string) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body rawBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.ErrBadRequest})
		return
	}
	uid := middleware.CurrentUserID(c)
	taskID := h.Queue.Start(kind, uid, middleware.CurrentUsername(c),
		fmt.Sprintf("problem #%d", id))
	if err := h.Runner.Enqueue(ai.Job{
		TaskID: taskID, Kind: kind,
		UserID: uid, ProblemID: uint(id), Raw: body.Raw,
	}); err != nil {
		h.Queue.End(taskID, err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID, "status": models.AITaskStatusRunning})
}

// Optimize kicks off optimization-suggestion generation for an AC submission.
// Mirrors Analyze but requires verdict=AC and uses a different prompt (no
// "乱写" JSON envelope — AC 代码已经通过所有用例). Result lands in
// ai_explanation, same field as the WA analysis — they are mutually exclusive
// by verdict so they can share the column.
func (h *AIHandler) Optimize(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var sub models.Submission
	if err := h.DB.First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	uid := middleware.CurrentUserID(c)
	if middleware.CurrentRole(c) != models.RoleAdmin && sub.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
		return
	}
	if sub.Verdict != models.VerdictAC {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅 AC 提交可生成优化建议"})
		return
	}
	// 题单维度的 AI 禁用：如果该提交挂在某个禁用 AI 的题单里，直接拒绝。
	if sub.ProblemSetID != nil && *sub.ProblemSetID > 0 {
		var ps models.ProblemSet
		if err := h.DB.Select("disable_ai").First(&ps, *sub.ProblemSetID).Error; err == nil && ps.DisableAI {
			c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
			return
		}
	}
	if sub.AIExplanation != "" {
		c.JSON(http.StatusOK, gin.H{"explanation": sub.AIExplanation, "cached": true})
		return
	}
	var prob models.Problem
	if err := h.DB.First(&prob, sub.ProblemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrProblemNotFound})
		return
	}
	_ = prob
	taskID := h.Queue.Start(models.AITaskKindOptimize, uid, middleware.CurrentUsername(c),
		fmt.Sprintf("submission #%d", sub.ID))
	if err := h.Runner.Enqueue(ai.Job{
		TaskID: taskID, Kind: models.AITaskKindOptimize,
		UserID: uid, SubmissionID: sub.ID,
	}); err != nil {
		h.Queue.End(taskID, err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID, "status": models.AITaskStatusRunning})
}

func (h *AIHandler) AITag(c *gin.Context)        { h.startAuthoringTask(c, models.AITaskKindTag) }
func (h *AIHandler) AIGenTitle(c *gin.Context)   { h.startAuthoringTask(c, models.AITaskKindGenTitle) }
func (h *AIHandler) AIGenDesc(c *gin.Context)    { h.startAuthoringTask(c, models.AITaskKindGenDesc) }
func (h *AIHandler) AIGenIdea(c *gin.Context)    { h.startAuthoringTask(c, models.AITaskKindGenIdea) }
func (h *AIHandler) AIGenExplain(c *gin.Context) { h.startAuthoringTask(c, models.AITaskKindGenExplain) }
func (h *AIHandler) AIGenAll(c *gin.Context)     { h.startAuthoringTask(c, models.AITaskKindGenAll) }

// ListTasks paginates the persistent AI task history for the admin page.
func (h *AIHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := h.DB.Model(&models.AITask{})
	if k := c.Query("kind"); k != "" {
		q = q.Where("kind = ?", k)
	}
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if kw := c.Query("username"); kw != "" {
		q = q.Where("username LIKE ?", "%"+kw+"%")
	}

	var total int64
	q.Count(&total)
	// List payload excludes the heavy prompt + output + result blobs to keep
	// responses small; use GetTask for the full row.
	var items []models.AITask
	q.Omit("prompt", "output", "result").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items)
	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": total, "page": page, "page_size": pageSize,
	})
}

// GetTask returns a single AI task row including the full prompt + output +
// result. Clients fetch this after receiving an `ai:task:done` SSE event for
// the task_id they enqueued.
func (h *AIHandler) GetTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var task models.AITask
	if err := h.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.ErrNotFound})
		return
	}
	// Restrict non-admins to tasks they initiated themselves. Analyze is the
	// only student-facing kind; exposing somebody else's prompt/output would
	// leak their code + the admin-paste raw sources for authoring flows.
	if middleware.CurrentRole(c) != models.RoleAdmin && task.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": i18n.ErrForbidden})
		return
	}
	c.JSON(http.StatusOK, task)
}

// RunningForProblem reports whether any AI task is currently running against
// the given problem. The admin UI uses this to disable every AI button while
// a flow is in flight — preventing overlapping GenAll/Tag/etc. calls that
// would fight for the same source-of-truth fields and potentially clobber
// each other on apply. Cheap enough to poll; pushed via `ai:tasks:changed`
// SSE so the UI also updates immediately when a task finishes.
func (h *AIHandler) RunningForProblem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	subject := fmt.Sprintf("problem #%d", id)
	var count int64
	h.DB.Model(&models.AITask{}).
		Where("subject = ? AND status = ?", subject, models.AITaskStatusRunning).
		Count(&count)
	c.JSON(http.StatusOK, gin.H{"running": count})
}
