package models

import "time"

// Verdict values: AC / WA / TLE / MLE / OLE / RE / CE / PE / SE / UKE / PENDING.
// SE covers DangerousSyscall + NonzeroExit; UKE covers judge-side InternalError
// or empty status. PE is "output matches after whitespace normalization but
// not byte-exact" — decided by compareOutput in the judge runner.
const (
	VerdictPending = "PENDING"
	VerdictAC      = "AC"
	VerdictWA      = "WA"
	VerdictTLE     = "TLE"
	VerdictMLE     = "MLE"
	VerdictOLE     = "OLE"
	VerdictRE      = "RE"
	VerdictCE      = "CE"
	VerdictPE      = "PE"
	VerdictSE      = "SE"
	VerdictUKE     = "UKE"
)

type Submission struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	UserID             uint   `gorm:"index;index:idx_user_created,priority:1;index:idx_ps_pid_user,priority:3;not null" json:"user_id"`
	ProblemID          uint   `gorm:"index;index:idx_problem_user,priority:1;index:idx_ps_pid_verdict_created,priority:2;index:idx_ps_pid_user,priority:2;not null" json:"problem_id"`
	ProblemSetID       *uint  `gorm:"index;index:idx_ps_pid_verdict_created,priority:1;index:idx_ps_pid_user,priority:1" json:"problemset_id,omitempty"`
	Language           string `gorm:"size:16;not null" json:"language"`
	Code               string `gorm:"type:text" json:"code"`
	Verdict            string `gorm:"size:16;index;index:idx_ps_pid_verdict_created,priority:3" json:"verdict"`
	Message            string `gorm:"type:text" json:"message"`
	TimeUsedMS         int    `json:"time_used_ms"`
	MemoryUsedKB       int    `json:"memory_used_kb"`
	TestcaseResultJSON string `gorm:"type:text" json:"testcase_result_json"`
	AIExplanation      string `gorm:"type:text" json:"ai_explanation"`
	// AIRejected 表示 AI 判定此次提交属于"未认真作答"（过短、空模板、随机
	// 字符、与题目无关 copy-paste 等），不予分析。rejected 的提交 ai_explanation
	// 为空，AIRejectReason 存了 AI 给出的分类理由。学生改代码重提或改写
	// 后可以再次触发 Analyze——见 handlers/ai.go Analyze。
	AIRejected     bool   `gorm:"default:false" json:"ai_rejected"`
	AIRejectReason string `gorm:"size:64" json:"ai_reject_reason,omitempty"`
	// idx_ps_pid_verdict_created / idx_ps_pid_user 覆盖题单排名与题单内
	// my_status 聚合的两条热查询，避免在 ranking.go / problemset.go 的 raw SQL
	// 上走全表扫。UserID 在 idx_ps_pid_user 里 priority=3 由 GORM 以 tag 再标
	// 一次的方式补齐（见下）。
	CreatedAt time.Time `gorm:"index;index:idx_user_created,priority:2;index:idx_ps_pid_verdict_created,priority:4" json:"created_at"`
}
