package models

import "time"

// AITask is an audit log for every AI call: who, what kind, against which
// target (problem/submission), and how it ended. Rows persist across restarts
// so the admin "AI 队列" page shows full history, not just in-flight calls.
const (
	AITaskKindAnalyze    = "analyze"     // student: explain a non-AC submission
	AITaskKindOptimize   = "optimize"    // student: optimize suggestions for an AC submission
	AITaskKindTag        = "tag"         // admin: suggest tags for a problem
	AITaskKindGenTitle   = "gen_title"   // admin: generate a problem title from raw
	AITaskKindGenDesc    = "gen_desc"    // admin: generate the problem description (incl. IO sections)
	AITaskKindGenIdea    = "gen_idea"    // admin: generate a code-free solution idea
	AITaskKindGenExplain = "gen_explain" // admin: generate the full solution markdown
	AITaskKindGenAll     = "gen_all"     // admin: 一键填充 — merged single-call flow

	AITaskStatusRunning = "running"
	AITaskStatusDone    = "done"
	AITaskStatusFailed  = "failed"
	AITaskStatusAborted = "aborted"
)

type AITask struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Kind     string `gorm:"size:16;index;index:idx_status_kind_subject,priority:2;not null" json:"kind"`
	UserID   uint   `gorm:"index" json:"user_id"`
	Username string `gorm:"size:64" json:"username"`
	// idx_status_subject 覆盖 "某对象当前是否有 running 任务" 和列表页批量
	// 查 subject IN (...) 这两类查询，避免 ai_tasks 历史累积后每次都扫全部
	// running 行再做字符串过滤。
	Subject string `gorm:"size:128;index:idx_ai_tasks_subject;index:idx_status_subject,priority:2;index:idx_status_kind_subject,priority:3" json:"subject"`
	// idx_status_started 支撑 admin "运行中任务"筛选 + "近期任务"排序这对热查询。
	Status     string     `gorm:"size:16;index;index:idx_status_started,priority:1;index:idx_status_subject,priority:1;index:idx_status_kind_subject,priority:1;not null" json:"status"`
	StartedAt  time.Time  `gorm:"index;index:idx_status_started,priority:2" json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	DurationMS int        `json:"duration_ms"`
	Error      string     `gorm:"type:text" json:"error,omitempty"`
	Prompt     string     `gorm:"type:text" json:"prompt,omitempty"`
	// Output stores the raw model response body (or the upstream error body
	// when the call failed). Lets the admin see what came back even when the
	// JSON could not be parsed — invaluable for diagnosing flaky / truncated
	// upstream responses (e.g. timeouts that cut the body mid-stream).
	Output string `gorm:"type:text" json:"output,omitempty"`
	// Result holds the structured per-kind payload the HTTP API used to return
	// synchronously (e.g. {"title":...}, {"description":...}, the parsed tag
	// suggestion). Async clients fetch it via GET /admin/ai/tasks/:id once the
	// corresponding ai:task:done event arrives. JSON-encoded string.
	Result string `gorm:"type:text" json:"result,omitempty"`
}
