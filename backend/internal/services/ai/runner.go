// Async AI runner: decouples HTTP handlers from the upstream model call. A
// fixed pool of goroutines drains the job channel, invokes Prompts.* per job
// kind, and persists both the raw audit trail (via Queue) and the structured
// Result (so clients polling GET /admin/ai/tasks/:id can pick it up when the
// `ai:task:done` event fires).
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

// Job is the unit of work the Runner consumes. TaskID is the row already
// created (via Queue.Start) in "running" state — the Runner is responsible
// for advancing it to done/failed.
type Job struct {
	TaskID       uint
	Kind         string
	UserID       uint
	ProblemID    uint // authoring flows (tag / gen_*)
	SubmissionID uint // Analyze
	Raw          string
	// ForceAnalyze 表示本次 analyze 任务由管理员触发，需跳过 prompt_wrong_answer
	// 里的 "乱写" 判定——模型若返回 ok=false 仍照常把 explanation 写回，哪怕
	// reason 非空。管理员手动点的解析多半是帮学生兜底或示范，不应因模型误伤
	// 打回去。
	ForceAnalyze bool
}

// Runner executes enqueued AI jobs on a bounded worker pool. One instance per
// process, shared across handlers. `ErrQueueFull` surfaces from Enqueue when
// the buffer is saturated — handlers MUST call Queue.End with this error so
// the audit row reflects the failure.
type Runner struct {
	db         *gorm.DB
	queue      *Queue
	prompts    *Prompts
	ch         chan Job
	wg         sync.WaitGroup
	workers    int
	// maxWait 是每个 AI 任务从入队执行到返回的上限，来自 config.toml
	// [ai].max_wait_seconds；0 即沿用 kindTimeout 的旧内置档位。
	maxWait    time.Duration
}

var ErrQueueFull = errors.New("AI 任务队列已满，请稍后再试")

// NewRunner boots the worker pool. workers ≥ 1 controls upstream concurrency
// (Bifrost is I/O-bound so 2–4 is plenty for a school deployment). cap is the
// channel buffer; a request that arrives while the buffer is full fails fast
// rather than holding open an HTTP connection.
//
// maxWait 是所有 kind 共用的单次调用上限（来自 config.toml）。传 0 则回退
// 到 kindTimeout 的旧逐档预算，保留向后兼容，不会意外把短 kind 等到饱和。
func NewRunner(db *gorm.DB, queue *Queue, prompts *Prompts, workers, cap int, maxWait time.Duration) *Runner {
	if workers < 1 {
		workers = 2
	}
	if cap < 1 {
		cap = 32
	}
	r := &Runner{
		db: db, queue: queue, prompts: prompts,
		ch:      make(chan Job, cap),
		workers: workers, maxWait: maxWait,
	}
	for i := 0; i < workers; i++ {
		r.wg.Add(1)
		go r.loop()
	}
	return r
}

// Enqueue is non-blocking: it returns ErrQueueFull rather than waiting on a
// full buffer. HTTP clients prefer a fast "try later" over a stalled POST.
func (r *Runner) Enqueue(j Job) error {
	select {
	case r.ch <- j:
		return nil
	default:
		return ErrQueueFull
	}
}

func (r *Runner) loop() {
	defer r.wg.Done()
	for j := range r.ch {
		r.run(j)
	}
}

func (r *Runner) run(j Job) {
	ctx, cancel := context.WithTimeout(context.Background(), r.jobTimeout(j.Kind))
	defer cancel()
	result, errMsg := r.execute(ctx, j)
	if errMsg == "" && result != nil {
		if b, err := json.Marshal(result); err == nil {
			r.queue.SetResult(j.TaskID, string(b))
		}
	}
	r.queue.End(j.TaskID, errMsg)
	if errMsg != "" {
		log.Printf("ai runner: task %d kind=%s failed: %s", j.TaskID, j.Kind, errMsg)
	}
}

// execute dispatches to the right Prompts.* call based on Kind. Returns the
// structured result (to be JSON-encoded into AITask.Result) plus an error
// message (empty on success). Per-call prompt + raw response are captured via
// Queue.SetPrompt / SetOutput so the audit log stays complete even on failure.
func (r *Runner) execute(ctx context.Context, j Job) (any, string) {
	switch j.Kind {
	case models.AITaskKindAnalyze:
		var sub models.Submission
		if err := r.db.First(&sub, j.SubmissionID).Error; err != nil {
			return nil, fmt.Sprintf("提交不存在：%v", err)
		}
		var prob models.Problem
		if err := r.db.First(&prob, sub.ProblemID).Error; err != nil {
			return nil, fmt.Sprintf("题目不存在：%v", err)
		}
		text, prompt, raw, err := r.prompts.AnalyzeWrongAnswer(ctx, &prob, &sub, sub.TestcaseResultJSON)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		// 约定模型返回 {ok, reason, explanation}；解析失败则退化为"整段当
		// explanation"，兼容老 config / 偶尔不守约的模型。
		type analyzeOut struct {
			OK          bool   `json:"ok"`
			Reason      string `json:"reason"`
			Explanation string `json:"explanation"`
		}
		var parsed analyzeOut
		rejected := false
		rejectReason := ""
		explanation := text
		if jerr := json.Unmarshal([]byte(text), &parsed); jerr == nil {
			if !parsed.OK {
				if j.ForceAnalyze {
					// 管理员强制模式：即便 ok=false 也不走拒绝流程；
					// explanation 大概率是空串，回落到原文 text 展示，
					// 并把模型写的 reason 当作注释附在页脚，方便老师手动
					// 把控。绕过是故意的——触发方已经是管理员，知道自己
					// 在做什么。
					rejected = false
					rejectReason = ""
					if strings.TrimSpace(parsed.Explanation) != "" {
						explanation = parsed.Explanation
					} else {
						explanation = text
					}
				} else {
					rejected = true
					rejectReason = parsed.Reason
					explanation = ""
				}
			} else {
				explanation = parsed.Explanation
			}
		}
		// 一次 UPDATE 写回：explanation / rejected / reason 三列一起同步，
		// 避免两次写引起的短暂不一致（学生端 race 查到 explanation 为空但
		// rejected=false 之类）。
		r.db.Model(&sub).Updates(map[string]any{
			"ai_explanation":    explanation,
			"ai_rejected":       rejected,
			"ai_reject_reason":  rejectReason,
		})
		return map[string]any{
			"explanation": explanation,
			"rejected":    rejected,
			"reason":      rejectReason,
		}, ""

	case models.AITaskKindOptimize:
		var sub models.Submission
		if err := r.db.First(&sub, j.SubmissionID).Error; err != nil {
			return nil, fmt.Sprintf("提交不存在：%v", err)
		}
		var prob models.Problem
		if err := r.db.First(&prob, sub.ProblemID).Error; err != nil {
			return nil, fmt.Sprintf("题目不存在：%v", err)
		}
		text, prompt, raw, err := r.prompts.OptimizeAC(ctx, &prob, &sub)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		// AC 没有"乱写"判定——直接把优化建议写回 ai_explanation（与错因分析
		// 共用同一字段，AC / 非 AC 互斥）。
		r.db.Model(&sub).Updates(map[string]any{
			"ai_explanation":   text,
			"ai_rejected":      false,
			"ai_reject_reason": "",
		})
		return map[string]any{"explanation": text}, ""

	case models.AITaskKindTag:
		var groups []models.TagGroup
		r.db.Order("order_index ASC, id ASC").Find(&groups)
		var tags []models.Tag
		r.db.Find(&tags)
		byGroup := map[uint][]models.Tag{}
		for _, t := range tags {
			byGroup[t.GroupID] = append(byGroup[t.GroupID], t)
		}
		sug, prompt, raw, err := r.prompts.TagProblem(ctx, j.Raw, groups, byGroup)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		return sug, ""

	case models.AITaskKindGenTitle:
		text, prompt, raw, err := r.prompts.GenTitle(ctx, j.Raw)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		// Admin 已经离开 ProblemEdit，结果要直接回写到 Problem 行，否则下次
		// 进编辑页什么都看不到。一键解析 / 单项生成 都同理。
		if j.ProblemID > 0 && text != "" {
			r.db.Model(&models.Problem{}).Where("id = ?", j.ProblemID).Update("title", text)
		}
		return map[string]string{"title": text}, ""

	case models.AITaskKindGenDesc:
		text, prompt, raw, err := r.prompts.GenDesc(ctx, j.Raw)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		if j.ProblemID > 0 && text != "" {
			r.db.Model(&models.Problem{}).Where("id = ?", j.ProblemID).Update("description", text)
		}
		return map[string]string{"description": text}, ""

	case models.AITaskKindGenIdea:
		text, prompt, raw, err := r.prompts.GenIdea(ctx, j.Raw)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		if j.ProblemID > 0 {
			r.db.Model(&models.Problem{}).Where("id = ?", j.ProblemID).Update("solution_idea_md", text)
		}
		return map[string]string{"solution_idea_md": text}, ""

	case models.AITaskKindGenExplain:
		text, prompt, raw, err := r.prompts.GenExplain(ctx, j.Raw)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		if j.ProblemID > 0 {
			r.db.Model(&models.Problem{}).Where("id = ?", j.ProblemID).Update("solution_md", text)
		}
		return map[string]string{"solution_md": text}, ""

	case models.AITaskKindGenAll:
		res, prompt, raw, err := r.prompts.GenAll(ctx, j.Raw)
		r.queue.SetPrompt(j.TaskID, prompt)
		r.queue.SetOutput(j.TaskID, raw)
		if err != nil {
			return nil, err.Error()
		}
		// 把一键解析出来的四个字段回写到 Problem 行，只更新非空字段，admin
		// 之前手动调整过但 AI 这轮没生成的字段不受影响。
		if j.ProblemID > 0 && res != nil {
			updates := map[string]any{}
			if res.Title != "" {
				updates["title"] = res.Title
			}
			if res.Description != "" {
				updates["description"] = res.Description
			}
			if res.SolutionIdeaMD != "" {
				updates["solution_idea_md"] = res.SolutionIdeaMD
			}
			if res.SolutionMD != "" {
				updates["solution_md"] = res.SolutionMD
			}
			if len(updates) > 0 {
				r.db.Model(&models.Problem{}).Where("id = ?", j.ProblemID).Updates(updates)
			}
		}
		return res, ""
	}
	return nil, fmt.Sprintf("unknown kind: %s", j.Kind)
}

// jobTimeout chooses the per-job context deadline. When r.maxWait is set via
// config.toml (> 0) it wins for every kind — operators running slow-thinking
// models (DeepSeek-V3 / R1) can lift the cap globally without touching code.
// Falling back to kindTimeout preserves pre-config defaults for deployments
// that don't override it.
func (r *Runner) jobTimeout(kind string) time.Duration {
	if r.maxWait > 0 {
		return r.maxWait
	}
	return kindTimeout(kind)
}

// kindTimeout preserves the per-flow ceilings from the old synchronous
// handlers: longer budgets for the richer markdown generators, tighter for
// one-shot calls like title. The model itself enforces no timeout — the
// bound is purely to keep stuck calls from sitting in the worker forever.
func kindTimeout(kind string) time.Duration {
	switch kind {
	case models.AITaskKindGenAll:
		return 180 * time.Second
	case models.AITaskKindGenDesc, models.AITaskKindGenExplain:
		return 120 * time.Second
	case models.AITaskKindGenIdea:
		return 90 * time.Second
	default:
		return 60 * time.Second
	}
}
